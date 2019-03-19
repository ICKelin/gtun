package gtund

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ICKelin/gtun/common"
	"github.com/ICKelin/gtun/logs"
)

var (
	defaultHeartbeat = 3
	defaultTimeout   = 5
	defautlAddr      = "127.0.0.1:9399"
)

type RegistryConfig struct {
	HeartbeatInterval int    `toml:"interval"`
	Timeout           int    `toml:"timeout"`
	Addr              string `toml:"addr"`
	Token             string `toml:"token"`
	Must              bool   `toml:"must"`
}

type Service struct {
	Name        string
	PublicIP    string
	Port        int
	ClientLimit int
	IsTap       bool
}

type Registry struct {
	heartbeatInterval time.Duration
	timeout           time.Duration
	addr              string
	token             string
	must              bool

	service   *Service
	sndbuffer chan []byte
}

func NewRegistry(cfg *RegistryConfig, service *Service) *Registry {
	hbintval := cfg.HeartbeatInterval
	if hbintval <= 0 {
		hbintval = defaultHeartbeat
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	addr := cfg.Addr
	if addr == "" {
		addr = defautlAddr
	}

	r := &Registry{
		heartbeatInterval: time.Second * time.Duration(hbintval),
		timeout:           time.Second * time.Duration(timeout),
		addr:              addr,
		token:             cfg.Token,
		must:              cfg.Must,
		service:           service,
		sndbuffer:         make(chan []byte),
	}
	return r
}

func (r *Registry) Run() error {
	for {
		err := r.run()
		logs.Error("disconnect with registry: %v", err)
		if r.must {
			time.Sleep(time.Second * 3)
			continue
		}
		return err
	}
}

func (g *Registry) Sync(inc int) error {
	s2g := &common.S2GUpdate{
		Count: inc,
	}

	msg, err := json.Marshal(s2g)
	if err != nil {
		return err
	}

	bytes, err := common.Encode(common.S2G_UPDATE_CLIENT_COUNT, msg)
	if err != nil {
		return err
	}

	g.sndbuffer <- bytes
	return nil
}

func (r *Registry) run() error {
	conn, err := net.Dial("tcp", r.addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = r.register(conn)
	if err != nil {
		return err
	}

	logs.Info("connect to %s success", r.addr)

	stop := make(chan struct{})
	go r.heartbeat(conn, stop)

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go r.send(conn, wg, stop)
	go r.recv(conn, wg)

	wg.Wait()
	close(stop)

	return fmt.Errorf("internal error")
}

func (r *Registry) register(conn net.Conn) (*common.ResponseBody, error) {
	reg := &common.S2GRegister{
		PublicIP:       r.service.PublicIP,
		Port:           r.service.Port,
		Name:           r.service.Name,
		Token:          r.token,
		MaxClientCount: r.service.ClientLimit,
		IsWindows:      r.service.IsTap,
	}

	bytes, err := json.Marshal(reg)
	if err != nil {
		return nil, err
	}

	regBytes, err := common.Encode(common.S2G_REGISTER, bytes)
	if err != nil {
		return nil, err
	}

	conn.SetWriteDeadline(time.Now().Add(r.timeout))
	_, err = conn.Write(regBytes)
	conn.SetWriteDeadline(time.Time{})
	if err != nil {
		return nil, err
	}

	cmd, resp, err := common.Decode(conn)
	if err != nil {
		return nil, err
	}

	if cmd != common.G2S_REGISTER {
		return nil, errors.New("invalid cmd")
	}

	msg := &common.ResponseBody{}
	err = json.Unmarshal(resp, msg)
	if err != nil {
		return nil, err
	}

	if msg.Code != common.CODE_SUCCESS {
		return nil, errors.New(msg.Message)
	}

	return msg, nil
}

func (r *Registry) heartbeat(conn net.Conn, stop chan struct{}) {
	for {
		select {
		case <-stop:
			logs.Info("heartbeat receive stop signal")
			return

		case <-time.After(r.heartbeatInterval):
			bytes, err := common.Encode(common.S2G_HEARTBEAT, nil)
			if err != nil {
				logs.Error("heartbear fail:", err)
				return
			}

			r.sndbuffer <- bytes
		}
	}
}

func (r *Registry) send(conn net.Conn, wg *sync.WaitGroup, stop chan struct{}) {
	defer wg.Done()
	defer conn.Close()

	for {
		select {
		case <-stop:
			logs.Info("send recv stop signal")
			return

		case msg := <-r.sndbuffer:
			conn.SetWriteDeadline(time.Now().Add(r.timeout))
			_, err := conn.Write(msg)
			conn.SetWriteDeadline(time.Time{})
			if err != nil {
				logs.Error("send fail: %v", err)
				return
			}
		}
	}
}

func (r *Registry) recv(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()

	for {
		cmd, bytes, err := common.Decode(conn)
		if err != nil {
			logs.Error("recv fail: %v", err)
			return
		}

		switch cmd {
		case common.G2S_HEARTBEAT:
			logs.Info("on G2S_HEARTBEAT: %s", conn.RemoteAddr().String())

		case common.G2S_UPDATE_CLIENT_COUNT:
			logs.Info("on G2S_UPDATE_CLIENT_COUNT: %s %s", conn.RemoteAddr().String(), string(bytes))

		default:
			logs.Warn("unimplemented cmd")
		}
	}
}
