package forward

import (
	"encoding/json"
	"fmt"
)

var errRegistered = fmt.Errorf("alread registered")

// Forward defines forwarder, such as tcp,udp,tun,wireguard
type Forward interface {
	Name() string
	Setup(cfg json.RawMessage) error
	ListenAndServe() error
}

var registeredForward = make(map[string]func() Forward)

func RegisterForward(name string, constructor func() Forward) error {
	if _, ok := registeredForward[name]; ok {
		return errRegistered
	}
	registeredForward[name] = constructor
	return nil
}
