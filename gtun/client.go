package gtun

import (
	"encoding/json"
	"fmt"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/ICKelin/gtun/common"
	"github.com/ICKelin/gtun/pkg/logs"
	"github.com/songgao/water"
)

var (
	defaultServer        = "127.0.0.1:9091"
	defaultClientAuthKey = "gtun-cs-token"
	defaultLayer2        = false
	defaultClientConfig  = &ClientConfig{
		ServerAddr: defaultServer,
		AuthKey:    defaultClientAuthKey,
	}
)

type ClientConfig struct {
	ServerAddr string `toml:"server"`
	AuthKey    string `toml:"auth"`
}

type Client struct {
	serverAddr string
	authKey    string
	myip       string
	gw         string
}

func NewClient(cfg *ClientConfig) *Client {
	if cfg == nil {
		cfg = defaultClientConfig
	}

	addr := cfg.ServerAddr
	if addr == "" {
		addr = defaultServer
	}

	authkey := cfg.AuthKey
	if authkey == "" {
		authkey = defaultClientAuthKey
	}

	return &Client{
		serverAddr: addr,
		authKey:    authkey,
	}
}

func (client *Client) Run() {
	ifce, err := NewIfce()
	if err != nil {
		logs.Error("new interface fail: %v", err)
		return
	}

	for {
		server := client.serverAddr
		conn, err := net.DialTimeout("tcp", server, time.Second*10)
		if err != nil {
			logs.Error("connect to server fail: %v", err)
			time.Sleep(time.Second * 3)
			continue
		}

		s2c, err := authorize(conn, client.authKey)
		if err != nil {
			logs.Error("auth fail: %v", err)
			time.Sleep(time.Second * 3)
			continue
		}

		logs.Info("connect to %s success, assign ip %s", server, s2c.AccessIP)

		client.myip = s2c.AccessIP
		client.gw = s2c.Gateway
		sndqueue := make(chan []byte)
		err = SetIfaceIP(ifce, s2c.AccessIP)
		if err != nil {
			logs.Error("setup iface fail: %v", err)
			time.Sleep(time.Second * 3)
			continue
		}

		done := make(chan struct{})
		wg := &sync.WaitGroup{}
		wg.Add(3)

		go ifaceRead(ifce, sndqueue)
		go heartbeat(sndqueue, done, wg)
		go snd(conn, sndqueue, done, wg)
		go rcv(conn, ifce, wg)
		wg.Wait()

		RemoveIfaceIP(ifce, s2c.AccessIP)
		logs.Info("reconnecting")
	}
}

func authorize(conn net.Conn, key string) (s2cauthorize *common.S2CAuthorize, err error) {
	c2sauthorize := &common.C2SAuthorize{
		OS:      common.OSID(runtime.GOOS),
		Version: common.Version(),
		Key:     key,
	}

	payload, err := json.Marshal(c2sauthorize)
	if err != nil {
		return nil, err
	}

	buff, _ := common.Encode(common.C2S_AUTHORIZE, payload)

	_, err = conn.Write(buff)
	if err != nil {
		return nil, err
	}

	cmd, resp, err := common.Decode(conn)
	if err != nil {
		return nil, err
	}

	if cmd != common.S2C_AUTHORIZE {
		err = fmt.Errorf("invalid authorize cmd %d", cmd)
		return nil, err
	}

	s2cauthorize = &common.S2CAuthorize{}
	err = json.Unmarshal(resp, s2cauthorize)
	if err != nil {
		return nil, err
	}

	return s2cauthorize, nil
}

func rcv(conn net.Conn, ifce *water.Interface, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()

	for {
		cmd, pkt, err := common.Decode(conn)
		if err != nil {
			logs.Info("decode fail: %v", err)
			break
		}
		switch cmd {
		case common.S2C_HEARTBEAT:
			logs.Debug("heartbeat from srv")

		case common.C2C_DATA:
			_, err := ifce.Write(pkt)
			if err != nil {
				logs.Error("read from iface fail: %v", err)
			}

		default:
			logs.Info("unimplement cmd %d %v", int(cmd), pkt)
		}
	}
}

func snd(conn net.Conn, sndqueue chan []byte, done chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()
	defer close(done)

	for {
		pkt := <-sndqueue
		conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
		_, err := conn.Write(pkt)
		conn.SetWriteDeadline(time.Time{})
		if err != nil {
			logs.Error("send packet fail: %v", err)
			break
		}
	}
}

func heartbeat(sndqueue chan []byte, done chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	tick := time.NewTicker(time.Second * 3)
	defer tick.Stop()

	for {
		select {
		case <-done:
			return

		case <-tick.C:
			bytes, _ := common.Encode(common.C2S_HEARTBEAT, nil)
			sndqueue <- bytes
		}
	}
}

func ifaceRead(ifce *water.Interface, sndqueue chan []byte) {
	packet := make([]byte, 65536)
	for {
		n, err := ifce.Read(packet)
		if err != nil {
			logs.Error("read from iface fail: %v", err)
			break
		}

		bytes, _ := common.Encode(common.C2C_DATA, packet[:n])
		sndqueue <- bytes
	}
}
