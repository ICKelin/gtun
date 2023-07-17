package route

import (
	"time"

	"github.com/ICKelin/gtun/internal/logs"
	"github.com/ICKelin/optw/transport"
)

type ClientConfig struct {
	Region     string
	ServerAddr string
	AuthKey    string
}

type Client struct {
	dialer       transport.Dialer
	region       string
	routeManager *Manager
}

func NewClient(region string, dialer transport.Dialer) *Client {
	return &Client{
		dialer:       dialer,
		region:       region,
		routeManager: GetRouteManager(),
	}
}

func (c *Client) ConnectNextHop() {
	for {
		conn, err := c.dialer.Dial()
		if err != nil {
			logs.Error("connect to %s fail: %v", c.region, err)
			time.Sleep(time.Second * 3)
			continue
		}

		logs.Info("connect to region %s success", c.region)
		hopConn := &HopInfo{Conn: conn}
		c.routeManager.AddRoute(c.region, hopConn)
		tick := time.NewTicker(time.Second * 1)
		for range tick.C {
			if hopConn.IsClosed() {
				break
			}
		}

		c.routeManager.DeleteRoute(c.region, hopConn)
		logs.Warn("reconnect")
	}
}
