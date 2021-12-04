package mux

import (
	"net"
	"time"

	"github.com/ICKelin/gtun/transport"
	"github.com/xtaci/smux"
)

var _ transport.Listener = &Listener{}
var _ transport.Dialer = &Dialer{}
var _ transport.Conn = &Conn{}

type Dialer struct {
	remote string
}

type Listener struct {
	laddr string
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

func (c *Conn) IsClosed() bool {
	return c.mux.IsClosed()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.mux.RemoteAddr()
}

func (c *Conn) LocalAddr() net.Addr {
	return c.mux.LocalAddr()
}

func NewDialer(remote string) transport.Dialer {
	return &Dialer{remote}
}

func (d *Dialer) Dial() (transport.Conn, error) {
	conn, err := net.Dial("tcp", d.remote)
	if err != nil {
		return nil, err
	}

	cfg := smux.DefaultConfig()
	cfg.KeepAliveTimeout = time.Second * 10
	cfg.KeepAliveInterval = time.Second * 3
	mux, err := smux.Client(conn, cfg)
	if err != nil {
		return nil, err
	}

	return &Conn{mux: mux}, nil
}

func NewListener(laddr string) *Listener {
	return &Listener{laddr: laddr}
}

func (l *Listener) Accept() (transport.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	cfg := smux.DefaultConfig()
	cfg.KeepAliveTimeout = time.Second * 10
	cfg.KeepAliveInterval = time.Second * 3
	mux, err := smux.Server(conn, cfg)
	if err != nil {
		return nil, err
	}

	return &Conn{mux: mux}, nil
}

func (l *Listener) Close() error {
	return l.Listener.Close()
}

func (l *Listener) Listen() error {
	listener, err := net.Listen("tcp", l.laddr)
	if err != nil {
		return err
	}

	l.Listener = listener
	return nil
}
