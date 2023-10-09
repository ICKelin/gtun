package route

import (
	"fmt"
	"github.com/ICKelin/optw/transport/transport_api"
	"time"

	"github.com/ICKelin/gtun/internal/logs"
	"github.com/ICKelin/optw/transport"
)

type ConnectionConfig struct {
	Region     string
	ServerAddr string
	AuthKey    string
}

type Connection struct {
	dialer     transport.Dialer
	region     string
	scheme     string
	serverAddr string
}

func CreateConnection(region, scheme, serverAddr, authKey string) (*Connection, error) {
	dialer, err := transport_api.NewDialer(scheme, serverAddr, authKey)
	if err != nil {
		return nil, err
	}

	return &Connection{
		dialer:     dialer,
		region:     region,
		scheme:     scheme,
		serverAddr: serverAddr,
	}, nil
}

func (c *Connection) ConnectNextHop() {
	for {
		conn, err := c.dialer.Dial()
		if err != nil {
			logs.Error("connect to %s fail: %v", c.String(), err)
			time.Sleep(time.Second * 3)
			continue
		}

		logs.Info("connect to %s success", c.String())
		hopConn := &HopInfo{Conn: conn}

		// add next hop connection to route
		GetRouteManager().AddRoute(c.region, hopConn)
		tick := time.NewTicker(time.Second * 1)
		for range tick.C {
			if hopConn.IsClosed() {
				break
			}
		}

		// delete next hop connection to route
		GetRouteManager().DeleteRoute(c.region, hopConn)
		logs.Warn("reconnect %s", c.String())
	}
}

func (c *Connection) String() string {
	return fmt.Sprintf("regions[%s] %s://%s", c.region, c.scheme, c.serverAddr)
}
