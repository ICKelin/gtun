package kcp

import (
	"net"

	"github.com/ICKelin/gtun/transport"
	kcpgo "github.com/xtaci/kcp-go"
	"github.com/xtaci/smux"
)

var _ transport.Dialer = &Dialer{}
var _ transport.Conn = &Conn{}
var _ transport.Listener = &Listener{}

type Dialer struct{}

type Conn struct {
	mux *smux.Session
}

type Listener struct {
	*kcpgo.Listener
}

func (c *Conn) OpenStream() (transport.Stream, error) {
	stream, err := c.mux.OpenStream()
	if err != nil {
		return nil, err
	}

	return stream, nil
}

func (c *Conn) AcceptStream() (transport.Stream, error) {
	return c.mux.AcceptStream()
}

func (c *Conn) Close() {
	c.mux.Close()
}

func (c *Conn) IsClosed() bool {
	return c.mux.IsClosed()
}

func (dialer *Dialer) Dial(remote string) (transport.Conn, error) {
	kcpconn, err := kcpgo.DialWithOptions(remote, nil, 10, 3)
	if err != nil {
		return nil, err
	}

	// kcp options
	// just for test
	kcpconn.SetStreamMode(true)
	kcpconn.SetWriteDelay(false)
	kcpconn.SetNoDelay(1, 10, 2, 1)
	kcpconn.SetWindowSize(1024, 1024)
	kcpconn.SetMtu(1350)
	kcpconn.SetACKNoDelay(true)
	kcpconn.SetReadBuffer(4194304)
	kcpconn.SetWriteBuffer(4194304)

	sess, err := smux.Client(kcpconn, nil)
	if err != nil {
		return nil, err
	}
	return &Conn{sess}, err
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
