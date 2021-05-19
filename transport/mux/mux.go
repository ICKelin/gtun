package mux

import (
	"net"

	"github.com/ICKelin/gtun/transport"
	"github.com/xtaci/smux"
)

var _ transport.Listener = &Listener{}
var _ transport.Dialer = &Dialer{}
var _ transport.Conn = &Conn{}

type Dialer struct{}

type Listener struct {
	net.Listener
}

type Conn struct {
	mux *smux.Session
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

func (d *Dialer) Dial(remote string) (transport.Conn, error) {
	conn, err := net.Dial("tcp", remote)
	if err != nil {
		return nil, err
	}

	mux, err := smux.Client(conn, nil)
	if err != nil {
		return nil, err
	}

	return &Conn{mux: mux}, nil
}

func (l *Listener) Accept() (transport.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	mux, err := smux.Server(conn, nil)
	if err != nil {
		return nil, err
	}

	return &Conn{mux: mux}, nil
}

func (l *Listener) Close() error {
	return l.Listener.Close()
}

func Listen(laddr string) (transport.Listener, error) {
	listener, err := net.Listen("tcp", laddr)
	if err != nil {
		return nil, err
	}

	return &Listener{Listener: listener}, nil
}
