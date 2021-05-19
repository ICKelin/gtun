package transport

import "time"

type Client struct{}

type Server struct{}

type Session interface {
	OpenStream() (Stream, error)
	AcceptStream() (Stream, error)
}

type Stream interface {
	Write(buf []byte) (int, error)
	Read(buf []byte) (int, error)
	Close() error
	SetWriteDeadline(time.Time) error
	SetReadDeadline(time.Time) error
}
