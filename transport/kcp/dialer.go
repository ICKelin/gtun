package kcp

import (
	"encoding/json"

	"github.com/ICKelin/gtun/transport"
	kcpgo "github.com/xtaci/kcp-go"
	"github.com/xtaci/smux"
)

var _ transport.Dialer = &Dialer{}

var defaultConfig = KCPConfig{
	FecDataShards:   10,
	FecParityShards: 3,
	Nodelay:         1,
	Interval:        10,
	Resend:          2,
	Nc:              1,
	SndWnd:          1024,
	RcvWnd:          1024,
	Mtu:             1350,
	AckNoDelay:      true,
	Rcvbuf:          4194304,
	SndBuf:          4194304,
}

type Dialer struct {
	config KCPConfig
}

func NewDialer(rawConfig json.RawMessage) *Dialer {
	dialer := &Dialer{}
	if len(rawConfig) <= 0 {
		dialer.config = defaultConfig
	} else {
		cfg := KCPConfig{}
		json.Unmarshal(rawConfig, &cfg)
		dialer.config = cfg
	}
	return dialer
}

func (dialer *Dialer) Dial(remote string) (transport.Conn, error) {
	cfg := dialer.config
	kcpconn, err := kcpgo.DialWithOptions(remote, nil, cfg.FecDataShards, cfg.FecParityShards)
	if err != nil {
		return nil, err
	}

	kcpconn.SetStreamMode(true)
	kcpconn.SetWriteDelay(false)
	kcpconn.SetNoDelay(cfg.Nodelay, cfg.Interval, cfg.Resend, cfg.Nc)
	kcpconn.SetWindowSize(cfg.RcvWnd, cfg.SndWnd)
	kcpconn.SetMtu(cfg.Mtu)
	kcpconn.SetACKNoDelay(cfg.AckNoDelay)
	kcpconn.SetReadBuffer(cfg.Rcvbuf)
	kcpconn.SetWriteBuffer(cfg.SndBuf)

	sess, err := smux.Client(kcpconn, nil)
	if err != nil {
		return nil, err
	}
	return &Conn{sess}, err
}