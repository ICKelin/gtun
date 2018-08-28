package gtund

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

type DHCPConfig struct {
	gateway string
	mask    string
}

type DHCP struct {
	gateway string
	mask    string
	table   *sync.Map
}

func NewDHCP(cfg *DHCPConfig) (*DHCP, error) {
	dhcp := &DHCP{
		gateway: cfg.gateway,
		mask:    defaultMask,
		table:   new(sync.Map),
	}

	sp := strings.Split(cfg.gateway, ".")
	if len(sp) != 4 {
		return nil, fmt.Errorf("invalid gateway address %s", cfg.gateway)
	}

	prefix := fmt.Sprintf("%s.%s.%s", sp[0], sp[1], sp[2])
	for i := 10; i < 250; i++ {
		ip := fmt.Sprintf("%s.%d", prefix, i)
		dhcp.table.Store(ip, false)
	}

	return dhcp, nil
}

func (dhcp *DHCP) SelectIP() (string, error) {
	ip := ""
	dhcp.table.Range(func(key, value interface{}) bool {
		if value.(bool) == false {
			dhcp.table.Store(key, true)
			ip = key.(string)
			return false
		}
		return true
	})

	if ip == "" {
		return "", errors.New("not avaliable ip in pool")
	}
	return ip, nil
}

func (dhcp *DHCP) RecycleIP(ip string) {
	dhcp.table.Store(ip, false)
}

func (dhcp *DHCP) InUsed(ip string) bool {
	v, ok := dhcp.table.Load(ip)
	if ok {
		return v.(bool)
	}

	return false
}
