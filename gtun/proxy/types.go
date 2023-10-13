package proxy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/ICKelin/gtun/internal/logs"
	"github.com/ICKelin/gtun/internal/utils"
	"net/http"
	"os"
	"strings"
	"time"
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
	ips := loadIPs(file)

	for _, ip := range ips {
		AddIP(region, ip)
	}
}

func loadIPs(file string) []string {
	if len(file) <= 0 {
		return nil
	}

	ips := make([]string, 0)
	var br *bufio.Reader
	if strings.HasPrefix(file, "http://") || strings.HasPrefix(file, "https://") {
		// load from url
		req, err := http.NewRequest("GET", file, nil)
		if err != nil {
			logs.Warn("load file fail: %v", err)
			return nil
		}

		cli := http.Client{
			Timeout: time.Second * 120,
		}

		resp, err := cli.Do(req)
		if err != nil {
			logs.Warn("load file fail: %v", err)
			return nil
		}

		defer resp.Body.Close()
		br = bufio.NewReader(resp.Body)
	} else {
		// load from file
		fp, err := os.Open(file)
		if err != nil {
			logs.Warn("open rule file fail: %v", err)
			return nil
		}
		defer fp.Close()
		br = bufio.NewReader(fp)
	}

	for {
		line, _, err := br.ReadLine()
		if err != nil {
			break
		}
		ips = append(ips, string(line))
	}

	return ips
}
