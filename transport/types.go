package transport

import (
	"net"
	"time"
)

// Dialer defines transport dialer for client side
type Dialer interface {
	Dial(target string) (Conn, error)
}

// Listener defines transport listener for server side
type Listener interface {
	Listen() error
	// Accept returns a connection
	// if an error occurs, it may suit each implements error
	Accept() (Conn, error)

	// Close close a listener
	Close() error

	// Addr returns address of listener
	Addr() net.Addr
}

// Conn defines a transport connection
type Conn interface {
	OpenStream() (Stream, error)
	AcceptStream() (Stream, error)
	Close()
	IsClosed() bool
	RemoteAddr() net.Addr
}

// Stream defines a transport stream base on
// Conn.OpenStream or Conn.AcceptStream
type Stream interface {
	Write(buf []byte) (int, error)
	Read(buf []byte) (int, error)
	Close() error
	SetWriteDeadline(time.Time) error
	SetReadDeadline(time.Time) error
}
