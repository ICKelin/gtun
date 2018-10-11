package gtund

import (
	"encoding/json"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/ICKelin/glog"
	"github.com/ICKelin/gtun/common"
)

var (
	defaultHeartbeat = time.Second * 3
	defaultTimeout   = time.Second * 5
)

type GodConfig struct {
	HeartbeatInterval int    `json:"god_hb_interval"`
	Timeout           int    `json:"god_conn_timeout"`
	GodAddr           string `json:"god_addr"`
	GodToken          string `json:"token"`
	Must              bool   `json:"must"`
}

type God struct {
	heartbeatInterval time.Duration
	timeout           time.Duration
	godAddr           string
	godToekn          string
	stop              chan struct{}
	sndbuffer         chan []byte
	rcvbuffer         chan []byte
}

func NewGod(cfg *GodConfig) *God {
	heartbeatInterval := time.Duration(cfg.HeartbeatInterval) * time.Second
	timeout := time.Duration(cfg.Timeout) * time.Second

	if cfg.HeartbeatInterval <= 0 {
		heartbeatInterval = defaultHeartbeat
	}

	if cfg.Timeout <= 0 {
		timeout = defaultTimeout
	}

	g := &God{
		heartbeatInterval: heartbeatInterval,
		timeout:           timeout,
		godAddr:           cfg.GodAddr,
		godToekn:          cfg.GodToken,
		stop:              make(chan struct{}),
		sndbuffer:         make(chan []byte),
		rcvbuffer:         make(chan []byte),
	}
	return g
}

func (g *God) Run() {
	for {
		err := g.run()
		glog.ERROR("disconnect with god:", err)
		time.Sleep(time.Second * 3)
	}
}

func (g *God) UpdateClientCount(inc int) error {
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

func (g *God) run() error {
	conn, err := net.Dial("tcp", g.godAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = g.register(conn)
	if err != nil {
		return err
	}

	glog.INFO("connect to ", g.godAddr, "success")

	stop := make(chan struct{})
	go g.heartbeat(conn, stop)

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go g.send(conn, wg)
	go g.recv(conn, wg)

	wg.Wait()
	close(stop)

	return nil
}

func (g *God) register(conn net.Conn) (*common.G2SResponse, error) {
	reg := &common.S2GRegister{
		PublicIP:   GetPublicIP(),
		ListenAddr: GetOpts().listenAddr,
		CIDR:       GetOpts().gateway,
		Region:     GetConfig().Region,
		Token:      g.godToekn,
	}
	bytes, err := json.Marshal(reg)
	if err != nil {
		return nil, err
	}

	regBytes, err := common.Encode(common.S2G_REGISTER, bytes)
	if err != nil {
		return nil, err
	}

	conn.SetWriteDeadline(time.Now().Add(g.timeout))
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

	msg := &common.G2SResponse{}
	err = json.Unmarshal(resp, msg)
	if err != nil {
		return nil, err
	}

	if msg.Code != common.CODE_SUCCESS {
		return nil, errors.New(msg.Message)
	}

	return msg, nil
}

func (g *God) heartbeat(conn net.Conn, stop chan struct{}) {
	for {
		select {
		case <-stop:
			return
		case <-time.After(g.heartbeatInterval):
			bytes, err := common.Encode(common.S2G_HEARTBEAT, nil)
			if err != nil {
				glog.ERROR("heartbear fail:", err)
				continue
			}

			g.sndbuffer <- bytes
		}
	}
}

func (g *God) send(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()

	for {
		msg := <-g.sndbuffer
		conn.SetWriteDeadline(time.Now().Add(g.timeout))
		_, err := conn.Write(msg)
		conn.SetWriteDeadline(time.Time{})
		if err != nil {
			glog.ERROR("send fail:", err)
			return
		}
	}
}

func (g *God) recv(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()

	for {
		cmd, bytes, err := common.Decode(conn)
		if err != nil {
			glog.ERROR("recv fail: ", err)
			return
		}

		switch cmd {
		case common.G2S_HEARTBEAT:
			glog.DEBUG("on G2S_HEARTBEAT:", conn.RemoteAddr().String())

		case common.G2S_UPDATE_CLIENT_COUNT:
			glog.DEBUG("on G2S_UPDATE_CLIENT_COUNT: ", conn.RemoteAddr().String(), string(bytes))

		default:
			glog.WARM("unimplemented cmd")
		}
	}
}
