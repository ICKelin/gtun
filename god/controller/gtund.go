package controller

import (
	"encoding/json"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/ICKelin/glog"
	"github.com/ICKelin/gtun/common"
)

type GtundConfig struct {
	Listener string `json:"gtund_listener"`
	Token    string `json:"token"` // 内部系统鉴权token
}

type gtund struct {
	listener string
	token    string
}

func NewGtund(cfg *GtundConfig) *gtund {
	return &gtund{
		listener: cfg.Listener,
		token:    cfg.Token,
	}
}

func (d *gtund) Run() error {
	listener, err := net.Listen("tcp", d.listener)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go d.onConn(conn)
	}
}

func (d *gtund) onConn(conn net.Conn) {
	defer conn.Close()

	reg, err := d.onRegister(conn)
	if err != nil {
		glog.ERROR("register fail: ", err)
		return
	}

	GetDB().Set(conn.RemoteAddr().String(), reg)
	defer GetDB().Del(conn.RemoteAddr().String())

	glog.INFO("register gtund from ", conn.RemoteAddr().String(), reg)
	defer glog.INFO("disconnect from", conn.RemoteAddr().String())

	sndbuffer := make(chan []byte)

	stop := make(chan struct{})
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go d.recv(conn, wg, sndbuffer, stop)
	go d.send(conn, wg, sndbuffer, stop)
	wg.Wait()
}

func (d *gtund) recv(conn net.Conn, wg *sync.WaitGroup, sndbuffer chan []byte, stop chan struct{}) {
	defer glog.INFO("close receive")
	defer close(stop)
	defer conn.Close()
	defer wg.Done()

	for {
		cmd, bytes, err := common.Decode(conn)
		if err != nil {
			glog.ERROR(err)
			break
		}

		switch cmd {
		case common.S2G_HEARTBEAT:
			d.onHeartbeat(conn, bytes, sndbuffer)

		case common.S2G_UPDATE_CLIENT_COUNT:
			d.onUpdate(conn, bytes, sndbuffer)

		default:
			glog.WARM("unimplemented cmd", cmd)

		}
	}
}

func (d *gtund) send(conn net.Conn, wg *sync.WaitGroup, sndbuffer chan []byte, stop chan struct{}) {
	defer glog.INFO("close send")
	defer conn.Close()
	defer wg.Done()

	for {
		select {
		case <-stop:
			return

		case bytes := <-sndbuffer:
			conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
			_, err := conn.Write(bytes)
			conn.SetWriteDeadline(time.Time{})
			if err != nil {
				glog.ERROR("write to client: ", err)
				return
			}
		}
	}

}

func (d *gtund) onRegister(conn net.Conn) (*common.S2GRegister, error) {
	cmd, bytes, err := common.Decode(conn)
	if err != nil {
		return nil, err
	}

	if cmd != common.S2G_REGISTER {
		return nil, errors.New("invalid command")
	}

	reg := common.S2GRegister{}
	err = json.Unmarshal(bytes, &reg)
	if err != nil {
		return nil, err
	}

	if reg.Token != d.token {
		msg := common.Response(nil, errors.New("invalid token"))
		resp, err := common.Encode(common.G2S_REGISTER, msg)
		if err != nil {
			return nil, err
		}
		conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
		conn.Write(resp)
		conn.SetWriteDeadline(time.Time{})

		return nil, errors.New("invalid token")
	}

	msg := common.Response(nil, nil)
	resp, err := common.Encode(common.G2S_REGISTER, msg)
	if err != nil {
		return nil, err
	}

	conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
	conn.Write(resp)
	conn.SetWriteDeadline(time.Time{})

	return &reg, nil
}

func (d *gtund) onHeartbeat(conn net.Conn, bytes []byte, sndbuffer chan []byte) {
	glog.DEBUG("on S2G_HEARTBEAT: ", conn.RemoteAddr().String(), string(bytes))

	bytes, err := common.Encode(common.G2S_HEARTBEAT, nil)
	if err != nil {
		glog.ERROR(err)
		return
	}
	sndbuffer <- bytes
}

func (d *gtund) onUpdate(conn net.Conn, bytes []byte, sndbuffer chan []byte) {
	glog.DEBUG("on S2G_UPDATE_CLIENT_COUNT: ", conn.RemoteAddr().String(), string(bytes))

	obj := &common.S2GUpdate{}
	err := json.Unmarshal(bytes, obj)
	if err != nil {
		d.response(nil, err, sndbuffer)
		return
	}

	rec, ok := GetDB().Get(conn.RemoteAddr().String())
	if !ok {
		d.response(rec, errors.New("not register yet!"), sndbuffer)
		return
	}

	val := rec.(*common.S2GRegister)
	val.Count += obj.Count
	GetDB().Set(conn.RemoteAddr().String(), val)
	d.response(val, nil, sndbuffer)
}

func (d *gtund) response(data interface{}, err error, sndbuffer chan []byte) {
	msg := common.Response(data, err)
	resp, _ := common.Encode(common.G2S_UPDATE_CLIENT_COUNT, msg)
	sndbuffer <- resp
}
