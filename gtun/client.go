package gtun

import (
	"net"
	"time"

	"github.com/ICKelin/gtun/pkg/logs"
	"github.com/hashicorp/yamux"
)

var defaultRegion = "US"

type ClientConfig struct {
	ServerAddr string `toml:"server"`
	AuthKey    string `toml:"auth"`
}

type Client struct {
	cfg        *ClientConfig
	sessionMgr *SessionManager
}

func NewClient(cfg *ClientConfig) *Client {
	return &Client{
		cfg:        cfg,
		sessionMgr: GetSessionManager(),
	}
}

func (client *Client) Run() {
	for {
		conn, err := net.DialTimeout("tcp", client.cfg.ServerAddr, time.Second*10)
		if err != nil {
			logs.Error("connect to server fail: %v", err)
			time.Sleep(time.Second * 3)
			continue
		}

		mux, err := yamux.Client(conn, nil)
		if err != nil {
			logs.Error("new yamux session fail: %v", err)
			time.Sleep(time.Second * 3)
			continue
		}

		sess := newSession(mux, defaultRegion)
		client.sessionMgr.AddSession(defaultRegion, sess)
		select {
		case <-sess.conn.CloseChan():
			break
		}

		client.sessionMgr.DeleteSession(defaultRegion)
		logs.Warn("reconnect")
	}
}
