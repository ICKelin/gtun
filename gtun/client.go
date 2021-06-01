package gtun

import (
	"time"

	"github.com/ICKelin/gtun/internal/logs"
	"github.com/ICKelin/gtun/transport"
)

type ClientConfig struct {
	Region     string
	ServerAddr string
	AuthKey    string
}

type Client struct {
	dialer     transport.Dialer
	sessionMgr *SessionManager
}

func NewClient(dialer transport.Dialer) *Client {
	return &Client{
		dialer:     dialer,
		sessionMgr: GetSessionManager(),
	}
}

func (c *Client) Run(region, remote string) {
	for {
		conn, err := c.dialer.Dial(remote)
		if err != nil {
			logs.Error("connect to %s fail: %v", remote, err)
			time.Sleep(time.Second * 3)
			continue
		}

		logs.Info("connect to region %s success", region)
		sess := newSession(conn, region)
		c.sessionMgr.AddSession(region, sess)
		tick := time.NewTicker(time.Second * 1)
		for range tick.C {
			if sess.conn.IsClosed() {
				break
			}
		}

		c.sessionMgr.DeleteSession(region)
		logs.Warn("reconnect")
	}
}
