package registry

import (
	"encoding/json"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/ICKelin/gtun/common"
	"github.com/ICKelin/gtun/god/config"
	"github.com/ICKelin/gtun/god/models"
	"github.com/ICKelin/gtun/logs"
)

type gtund struct {
	listener     string
	token        string
	gtundManager *models.GtundManager
}

func NewGtund(cfg *config.GtundConfig) *gtund {
	return &gtund{
		listener:     cfg.Listener,
		token:        cfg.Token,
		gtundManager: models.GetGtundManager(),
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
		logs.Error("register fail: %v", err)
		return
	}

	gtundInfo, err := d.gtundManager.NewGtund(reg)
	if err != nil {
		logs.Error("new gtund fail: ", err)
		return
	}

	defer d.gtundManager.RemoveGtund(gtundInfo.Id)

	logs.Info("register gtund from %s %v", conn.RemoteAddr().String(), reg)
	defer logs.Info("disconnect from %s", conn.RemoteAddr().String())

	sndbuffer := make(chan []byte)

	stop := make(chan struct{})
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go d.recv(conn, wg, sndbuffer, stop)
	go d.send(conn, wg, sndbuffer, stop)

	wg.Wait()
}

func (d *gtund) recv(conn net.Conn, wg *sync.WaitGroup, sndbuffer chan []byte, stop chan struct{}) {
	defer logs.Info("close receive")
	defer close(stop)
	defer conn.Close()
	defer wg.Done()

	for {
		cmd, bytes, err := common.Decode(conn)
		if err != nil {
			logs.Error("decode fail: %v", err)
			break
		}

		switch cmd {
		case common.S2G_HEARTBEAT:
			d.onHeartbeat(conn, bytes, sndbuffer)

		case common.S2G_UPDATE_CLIENT_COUNT:
			d.onUpdate(conn, bytes, sndbuffer)

		default:
			logs.Warn("unimplemented cmd", cmd)

		}
	}
}

func (d *gtund) send(conn net.Conn, wg *sync.WaitGroup, sndbuffer chan []byte, stop chan struct{}) {
	defer logs.Info("close send")
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
				logs.Error("write to client: ", err)
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
	logs.Debug("on S2G_HEARTBEAT: %s %s", conn.RemoteAddr().String(), string(bytes))

	bytes, err := common.Encode(common.G2S_HEARTBEAT, nil)
	if err != nil {
		logs.Error("encode fail: %v", err)
		return
	}
	sndbuffer <- bytes
}

func (d *gtund) onUpdate(conn net.Conn, bytes []byte, sndbuffer chan []byte) {
	logs.Debug("on S2G_UPDATE_CLIENT_COUNT: %s %s", conn.RemoteAddr().String(), string(bytes))

	obj := &common.S2GUpdate{}
	err := json.Unmarshal(bytes, obj)
	if err != nil {
		d.response(nil, err, sndbuffer)
		return
	}

	gtund, err := d.gtundManager.IncReferenceCount(obj.Count)
	if err != nil {
		d.response(nil, err, sndbuffer)
		return
	}

	d.response(gtund, nil, sndbuffer)
}

func (d *gtund) response(data interface{}, err error, sndbuffer chan []byte) {
	msg := common.Response(data, err)
	resp, _ := common.Encode(common.G2S_UPDATE_CLIENT_COUNT, msg)
	sndbuffer <- resp
}
