package gtun

import (
	"time"

	"github.com/ICKelin/gtun/internal/logs"
	"github.com/ICKelin/gtun/transport/mux"
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
		dialer := mux.Dialer{}
		conn, err := dialer.Dial(client.cfg.ServerAddr)
		if err != nil {
			logs.Error("connect to %s fail: %v", client.cfg.ServerAddr, err)
			time.Sleep(time.Second * 3)
			continue
		}

		logs.Info("connect to region %s success", client.cfg.Region)
		sess := newSession(conn, client.cfg.Region)
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
