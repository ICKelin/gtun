package gtund

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/ICKelin/glog"
	"github.com/ICKelin/gtun/common"
)

var (
	defaultHeartbeat = time.Second * 3
	defaultTimeout   = time.Second * 5
)

type godConfig struct {
	HeartbeatInterval time.Duration
	Timeout           time.Duration
	GodAddr           string
}

type god struct {
	heartbeatInterval time.Duration
	timeout           time.Duration
	godAddr           string
	stop              chan struct{}
}

func NewGod(cfg godConfig) *god {
	heartbeatInterval := cfg.HeartbeatInterval
	timeout := cfg.Timeout

	if cfg.HeartbeatInterval <= 0 {
		heartbeatInterval = defaultHeartbeat
	}

	if cfg.Timeout <= 0 {
		timeout = defaultTimeout
	}
	g := &god{
		heartbeatInterval: heartbeatInterval,
		timeout:           timeout,
		godAddr:           cfg.GodAddr,
		stop:              make(chan struct{}),
	}
	return g
}

func (g *god) Run() error {
	conn, err := net.Dial("tcp", g.godAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = g.register(conn)
	if err != nil {
		return err
	}

	for {
		select {
		case <-time.After(g.heartbeatInterval):
			err = g.heartbeat(conn)
			if err != nil {
				return err
			}

		case <-g.stop:
			return nil
		}
	}
}

func (g *god) register(conn net.Conn) (*common.G2SRegister, error) {
	reg := &common.S2GRegister{}
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

	//TODO: receive g2s register msg
	return nil, nil
}

func (g *god) heartbeat(conn net.Conn) error {
	bytes, err := common.Encode(common.S2G_HEARTBEAT, nil)
	if err != nil {
		return err
	}
	conn.SetWriteDeadline(time.Now().Add(g.timeout))
	_, err = conn.Write(bytes)
	conn.SetWriteDeadline(time.Time{})

	cmd, _, err := common.Decode(conn)
	if err != nil {
		return err
	}

	if cmd != common.G2S_HEARTBEAT {
		return fmt.Errorf("invalid cmd: %d", cmd)
	}

	glog.DEBUG("heartbeat from god:", conn.RemoteAddr().String())
	return err
}
