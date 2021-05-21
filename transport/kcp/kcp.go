package kcp

import (
	"net"

	"github.com/ICKelin/gtun/transport"
	kcpgo "github.com/xtaci/kcp-go"
	"github.com/xtaci/smux"
)

var _ transport.Dialer = &Dialer{}
var _ transport.Conn = &Conn{}

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
	kcpconn, err := kcpgo.DialWithOptions(remote, nil, 0, 0)
	if err != nil {
		return nil, err
	}

	// kcp options
	// just for test
	kcpconn.SetStreamMode(true)
	kcpconn.SetWriteDelay(true)
	kcpconn.SetNoDelay(1, 10, 2, 1)
	kcpconn.SetWindowSize(2048, 2048)
	kcpconn.SetMtu(1480)
	kcpconn.SetACKNoDelay(false)

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
	conn.SetWriteDelay(true)
	conn.SetNoDelay(1, 10, 2, 1)
	conn.SetWindowSize(2048, 2048)
	conn.SetMtu(1480)
	conn.SetACKNoDelay(false)
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
	listener, err := kcpgo.ListenWithOptions(laddr, nil, 0, 0)
	if err != nil {
		return nil, err
	}
	return &Listener{Listener: listener}, nil
}
