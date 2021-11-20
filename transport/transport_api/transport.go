package transport_api

import (
	"errors"
	"github.com/ICKelin/gtun/transport"
	"github.com/ICKelin/gtun/transport/kcp"
	"github.com/ICKelin/gtun/transport/mux"
)

const (
	protoKCP    = "kcp"
	protoTCPMux = "mux"
)

var (
	errUnsupported = errors.New("transport_api: unsupported protocol")
)

func NewListen(scheme, addr, cfg string) (transport.Listener, error) {
	var listener transport.Listener
	switch scheme {
	case protoKCP:
		listener = kcp.NewListener(addr, []byte(cfg))
		err := listener.Listen()
		if err != nil {
			return nil, err
		}

	case protoTCPMux:
		listener = mux.NewListener(addr)
		err := listener.Listen()
		if err != nil {
			return nil, err
		}

	default:
		return nil, errUnsupported
	}
	return listener, nil
}

func NewDialer(scheme, addr, cfg string) (transport.Dialer, error) {
	var dialer transport.Dialer
	switch scheme {
	case protoKCP:
		dialer = kcp.NewDialer(addr, []byte(cfg))
	case protoTCPMux:
		dialer = mux.NewDialer(addr)
	default:
		return nil, errUnsupported
	}

	return dialer, nil
}
