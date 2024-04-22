package proxy

import (
	"encoding/json"
	"fmt"
	"github.com/ICKelin/gtun/src/internal/logs"
)

var errRegistered = fmt.Errorf("already registered")
var errNotRegister = fmt.Errorf("proxy not register")

// Proxy defines Proxies, such as tproxy_tcp, tproxy_udp,ip_tun, ip_wireguard
type Proxy interface {
	Name() string
	Setup(region string, cfg json.RawMessage) error
	ListenAndServe() error
}

var registerProxy = make(map[string]func() Proxy)

func Register(name string, constructor func() Proxy) error {
	if _, ok := registerProxy[name]; ok {
		return errRegistered
	}
	registerProxy[name] = constructor
	return nil
}

func Serve(region string, proxyConfig map[string]json.RawMessage) error {
	logs.Debug("region %s proxy config %s", region, proxyConfig)
	err := setup(region, proxyConfig)
	if err != nil {
		fmt.Printf("region[%s] setup proxy fail: %v\n", region, err)
		return err
	}
	return nil
}

func setup(region string, proxyConfigs map[string]json.RawMessage) error {
	for name, config := range proxyConfigs {
		constructor := registerProxy[name]
		if constructor == nil {
			return errNotRegister
		}
		p := constructor()
		err := p.Setup(region, config)
		if err != nil {
			return err
		}

		go p.ListenAndServe()
	}
	return nil
}
