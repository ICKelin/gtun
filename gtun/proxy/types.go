package proxy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/ICKelin/gtun/internal/logs"
	"github.com/ICKelin/gtun/internal/utils"
	"os"
)

var errRegistered = fmt.Errorf("already registered")
var errNotRegister = fmt.Errorf("proxy not register")
var ipsetNamePrefix = "GTUN-"

// Proxy defines Proxies, such as tproxy_tcp, tproxy_udp,ip_tun, ip_wireguard
type Proxy interface {
	Name() string
	Setup(cfg json.RawMessage) error
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

func Setup(region, ruleFile string, proxyConfigs map[string]string) error {
	for name, config := range proxyConfigs {
		constructor := registerProxy[name]
		if constructor == nil {
			return errNotRegister
		}
		p := constructor()
		err := p.Setup([]byte(config))
		if err != nil {
			return err
		}

		AddFromFile(region, ruleFile)
		go p.ListenAndServe()
	}
	return nil
}

func AddIP(region string, ip string) error {
	out, err := utils.ExecCmd("ipset", []string{"add", ipsetNamePrefix + region, ip})
	if err != nil {
		return fmt.Errorf("add to ipset fail: %v %s", err, out)
	}
	return nil
}

func DelIP(region, ip string) error {
	out, err := utils.ExecCmd("ipset", []string{"del", ipsetNamePrefix + region, ip})
	if err != nil {
		return fmt.Errorf("del from ipset fail: %v %s", err, out)
	}
	return nil
}

func AddApp(region, appName string) error {
	AddFromFile(region, appName)
	return nil
}

func AddFromFile(region, file string) {
	if len(file) <= 0 {
		return
	}

	fp, err := os.Open(file)
	if err != nil {
		logs.Warn("open rule file fail: %v", err)
		return
	}
	defer fp.Close()

	br := bufio.NewReader(fp)
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			break
		}

		err = AddIP(region, string(line))
		if err != nil {
			logs.Warn("add %s %s proxy fail: %v", region, string(line), err)
		}
	}
}
