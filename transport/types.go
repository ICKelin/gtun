package transport

import (
	"net"
	"time"
)

type Dialer interface {
	Dial(target string) (Conn, error)
}

type Listener interface {
	Accept() (Conn, error)
	Close() error
	Addr() net.Addr
}

type Conn interface {
	OpenStream() (Stream, error)
	AcceptStream() (Stream, error)
	Close()
}

type Stream interface {
	Write(buf []byte) (int, error)
	Read(buf []byte) (int, error)
	Close() error
	SetWriteDeadline(time.Time) error
	SetReadDeadline(time.Time) error
}
