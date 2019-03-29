package registry

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/ICKelin/gtun/common"
	"github.com/ICKelin/gtun/logs"
)

type GtundConfig struct {
	Listener string `toml:"listen"`
	Token    string `toml:"token"` // 内部系统鉴权token
}

type gtund struct {
	listener string
	token    string
	m        *Models
}

func NewGtund(cfg *GtundConfig, m *Models) *gtund {
	return &gtund{
		listener: cfg.Listener,
		token:    cfg.Token,
		m:        m,
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

func (d *gtund) GetGtundList(w http.ResponseWriter, r *http.Request) {
	list := d.m.Status()
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(list))
}

func (d *gtund) onConn(conn net.Conn) {
	defer conn.Close()

	reg, err := d.onRegister(conn)
	if err != nil {
		logs.Error("register fail: %v", err)
		return
	}

	logs.Info("register gtund from %s %v", conn.RemoteAddr().String(), reg)
	defer logs.Info("disconnect from %s", conn.RemoteAddr().String())
	defer d.onDisconnect(conn.RemoteAddr().String())

	sndbuffer := make(chan []byte)

	done := make(chan struct{})
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go d.recv(conn, wg, sndbuffer, done)
	go d.send(conn, wg, sndbuffer, done)

	wg.Wait()
}

func (d *gtund) recv(conn net.Conn, wg *sync.WaitGroup, sndbuffer chan []byte, done chan struct{}) {
	defer logs.Info("close receive")
	defer close(done)
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

func (d *gtund) send(conn net.Conn, wg *sync.WaitGroup, sndbuffer chan []byte, done chan struct{}) {
	defer logs.Info("close send")
	defer conn.Close()
	defer wg.Done()

	for {
		select {
		case <-done:
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

	// store gtund info
	d.m.NewGtund(conn.RemoteAddr().String(), &reg)

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

func (d *gtund) onDisconnect(addr string) {
	d.m.RemoveGtund(addr)
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

	d.m.UpdateRefCount(conn.RemoteAddr().String(), obj.Count)
}

func (d *gtund) response(data interface{}, err error, sndbuffer chan []byte) {
	msg := common.Response(data, err)
	resp, _ := common.Encode(common.G2S_UPDATE_CLIENT_COUNT, msg)
	sndbuffer <- resp
}
