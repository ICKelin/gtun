package gtun

import (
	"net"
	"time"

	"github.com/ICKelin/gtun/pkg/logs"
	"github.com/xtaci/smux"
)

type ClientConfig struct {
	Region     string
	ServerAddr string
	AuthKey    string
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
			logs.Error("connect to %s fail: %v", client.cfg.ServerAddr, err)
			time.Sleep(time.Second * 3)
			continue
		}

		mux, err := smux.Client(conn, nil)
		if err != nil {
			logs.Error("new yamux session fail: %v", err)
			time.Sleep(time.Second * 3)
			continue
		}

		logs.Info("connect to region %s success", client.cfg.Region)
		sess := newSession(mux, client.cfg.Region)
		client.sessionMgr.AddSession(client.cfg.Region, sess)
		tick := time.NewTicker(time.Second * 1)
		for range tick.C {
			if sess.conn.IsClosed() {
				break
			}
		}

		client.sessionMgr.DeleteSession(client.cfg.Region)
		logs.Warn("reconnect")
	}
}
