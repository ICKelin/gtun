package kcp

import (
	"encoding/json"
	"net"

	"github.com/ICKelin/gtun/transport"
	kcpgo "github.com/xtaci/kcp-go"
	"github.com/xtaci/smux"
)

var _ transport.Listener = &Listener{}

var defaultListenConfig = KCPConfig{
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

type Listener struct {
	config KCPConfig
	*kcpgo.Listener
}

func NewListener(rawConfig json.RawMessage) *Listener {
	l := &Listener{}
	if len(rawConfig) <= 0 {
		l.config = defaultConfig
	} else {
		cfg := KCPConfig{}
		json.Unmarshal(rawConfig, &cfg)
		l.config = cfg
	}
	return l
}

func (l *Listener) Listen(laddr string) error {
	kcpLis, err := kcpgo.ListenWithOptions(laddr, nil, 10, 3)
	if err != nil {
		return err
	}
	kcpLis.SetReadBuffer(4194304)
	kcpLis.SetWriteBuffer(4194304)
	l.Listener = kcpLis
	return nil
}

func (l *Listener) Accept() (transport.Conn, error) {
	cfg := l.config
	kcpconn, err := l.Listener.AcceptKCP()
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
	mux, err := smux.Server(kcpconn, nil)
	if err != nil {
		return nil, err
	}

	return &Conn{mux: mux}, nil
}

func (l *Listener) Close() error {
	return l.Listener.Close()
}

func (l *Listener) Addr() net.Addr {
	return l.Listener.Addr()
}
