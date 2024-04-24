package route

import (
	"fmt"
	"github.com/ICKelin/gtun/src/gtun/config"
	"github.com/ICKelin/gtun/src/internal/logs"
	"github.com/ICKelin/optw/transport/transport_api"
	"time"

	"github.com/ICKelin/optw/transport"
)

var cm = &connManager{regionConn: map[string][]*conn{}}

type connManager struct {
	regionConn map[string][]*conn
}

func (cm *connManager) startConn() {
	for _, conns := range cm.regionConn {
		for _, conn := range conns {
			go conn.connect()
		}
	}
}

type conn struct {
	dialer     transport.Dialer
	region     string
	scheme     string
	serverAddr string
}

func newConn(region, scheme, serverAddr, authKey string) (*conn, error) {
	dialer, err := transport_api.NewDialer(scheme, serverAddr, authKey)
	if err != nil {
		return nil, err
	}
	dialer.SetAccessToken(config.Default().AccessToken)

	return &conn{
		dialer:     dialer,
		region:     region,
		scheme:     scheme,
		serverAddr: serverAddr,
	}, nil
}

func (c *conn) connect() {
	for {
		conn, err := c.dialer.Dial()
		if err != nil {
			logs.Error("connect to %s fail: %v", c.String(), err)
			time.Sleep(time.Second * 3)
			continue
		}

		logs.Info("connect to %s success", c.String())

		// add next hop connection to route
		routeEle := &routeItem{
			region:     c.region,
			scheme:     c.scheme,
			serverAddr: c.serverAddr,
			Conn:       conn,
		}
		GetRouteManager().addRoute(c.region, routeEle)
		tick := time.NewTicker(time.Second * 1)
		for range tick.C {
			if conn.IsClosed() {
				break
			}
		}

		// delete next hop connection to route
		GetRouteManager().deleteRoute(c.region, routeEle)
		logs.Warn("reconnect %s", c.String())
	}
}

func (c *conn) String() string {
	return fmt.Sprintf("regions[%s] %s://%s", c.region, c.scheme, c.serverAddr)
}
