package mux

import (
	"fmt"

	"github.com/ICKelin/gtun/transport"
	"github.com/hashicorp/yamux"
)

type Session struct {
	*yamux.Session
}

type Stream struct {
	*yamux.Stream
}

func NewSession(instance interface{}) (*Session, error) {
	sess, ok := instance.(*yamux.Session)
	if !ok {
		return nil, fmt.Errorf("invalid session type")
	}
	return &Session{Session: sess}
}

func (s *Session) OpenStream() (transport.Stream, error) {
	stream, err := s.sess.OpenStream()
	if err != nil {
		return nil, err
	}
	return &Stream{Stream: stream}, nil
}

func (s *Session) AcceptStream(transport.Stream, error) {
	return s.sess.AcceptStream()
}
