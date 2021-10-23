package kcp

import (
	"net"

	"github.com/ICKelin/gtun/transport"
	kcpgo "github.com/xtaci/kcp-go"
	"github.com/xtaci/smux"
)

var _ transport.Listener = &Listener{}

type Listener struct {
	*kcpgo.Listener
}

func (l *Listener) Accept() (transport.Conn, error) {
	conn, err := l.Listener.AcceptKCP()
	if err != nil {
		return nil, err
	}

	conn.SetStreamMode(true)
	conn.SetWriteDelay(false)
	conn.SetNoDelay(1, 10, 2, 1)
	conn.SetWindowSize(1024, 1024)
	conn.SetMtu(1350)
	conn.SetACKNoDelay(true)
	mux, err := smux.Server(conn, nil)
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

func Listen(laddr string) (transport.Listener, error) {
	listener, err := kcpgo.ListenWithOptions(laddr, nil, 10, 3)
	if err != nil {
		return nil, err
	}
	listener.SetReadBuffer(4194304)
	listener.SetWriteBuffer(4194304)
	return &Listener{Listener: listener}, nil
}
