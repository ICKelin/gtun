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

type Manager struct {
	ruleFiles     map[string]proxyRule
	registerProxy map[string]func() Proxy
}

type proxyRule struct {
	regionProxyFile string
	ipProxyFile     string
	proxyIPs        map[string]struct{}
}

var manager = &Manager{
	ruleFiles:     make(map[string]proxyRule),
	registerProxy: make(map[string]func() Proxy),
}

func GetManager() *Manager {
	return manager
}

// Proxy defines Proxies, such as tproxy_tcp, tproxy_udp,ip_tun, ip_wireguard
type Proxy interface {
	Name() string
	Setup(cfg json.RawMessage) error
	ListenAndServe() error
}

func (m *Manager) Register(name string, constructor func() Proxy) error {
	if _, ok := m.registerProxy[name]; ok {
		return errRegistered
	}
	m.registerProxy[name] = constructor
	return nil
}

func (m *Manager) Setup(region, ruleFile, ipProxyFile string, proxyConfigs map[string]string) error {
	for name, config := range proxyConfigs {
		constructor := m.registerProxy[name]
		if constructor == nil {
			return errNotRegister
		}
		p := constructor()
		err := p.Setup([]byte(config))
		if err != nil {
			return err
		}

		m.AddFromFile(region, ruleFile)
		m.AddFromFile(region, ipProxyFile)

		rule := m.ruleFiles[region]
		rule.ipProxyFile = ipProxyFile
		rule.regionProxyFile = ruleFile
		go p.ListenAndServe()
	}
	return nil
}

func (m *Manager) AddIP(region string, ip string) error {
	out, err := utils.ExecCmd("ipset", []string{"add", "-!", ipsetNamePrefix + region, ip})
	if err != nil {
		return fmt.Errorf("add to ipset fail: %v %s", err, out)
	}

	rule := m.ruleFiles[region]
	if len(rule.ipProxyFile) <= 0 {
		return nil
	}

	rule.proxyIPs[ip] = struct{}{}
	// TODO: write to proxy file
	return nil
}

func (m *Manager) DelIP(region, ip string) error {
	out, err := utils.ExecCmd("ipset", []string{"del", "-!", ipsetNamePrefix + region, ip})
	if err != nil {
		return fmt.Errorf("del from ipset fail: %v %s", err, out)
	}

	rule := m.ruleFiles[region]
	if len(rule.ipProxyFile) <= 0 {
		return nil
	}
	delete(rule.proxyIPs, ip)
	// TODO: delete from proxy file
	return nil
}

func (m *Manager) IPList(region string) []string {
	rule := m.ruleFiles[region]
	ips := make([]string, 0)
	for ip, _ := range rule.proxyIPs {
		ips = append(ips, ip)
	}
	return ips
}

func (m *Manager) AddApp(region, appName string) error {
	m.AddFromFile(region, appName)
	return nil
}

func (m *Manager) AddFromFile(region, file string) {
	ips := m.loadIPs(file)

	for _, ip := range ips {
		m.AddIP(region, ip)
	}
}

func (m *Manager) loadIPs(file string) []string {
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
