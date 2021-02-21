package gtund

import (
	"fmt"
	"net"
	"sync"
)

var (
	defaultCidr       = "100.64.240.1/24"
	errNotAvaliableIP = fmt.Errorf("not avaliable ip")
)

type DHCPConfig struct {
	Gateway     string `toml:"gateway"`
	Nameserver  string `toml:"nameserver"`
	CIDR        string `toml:"cidr"`
	ClientCount int    `toml:"-"`
}

type DHCP struct {
	sync.Mutex
	gateway string
	mask    int
	table   map[string]bool
}

func NewDHCP(cfg DHCPConfig) (*DHCP, error) {
	cidr := cfg.CIDR
	gateway := cfg.Gateway

	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	mask, _ := ipnet.Mask.Size()
	p := ip.To4()

	igateway := (int(p[0]) << 24) + (int(p[1]) << 16) + (int(p[2]) << 8) + int(p[3])
	ipcount := (1 << uint(32-mask)) - 2 // ignore broadcast

	dhcp := &DHCP{}
	dhcp.gateway = gateway
	dhcp.mask = mask
	dhcp.table = make(map[string]bool)

	for i := 0; i < ipcount; i++ {
		uip := igateway + i
		ele := fmt.Sprintf("%d.%d.%d.%d", byte(uip>>24), byte(uip>>16), byte(uip>>8), byte(uip))
		dhcp.table[ele] = false
	}
	dhcp.table[gateway] = true
	cfg.ClientCount = ipcount - 1 // ignore gateway

	return dhcp, nil
}

func (dhcp *DHCP) SelectIP() (string, error) {
	dhcp.Lock()
	defer dhcp.Unlock()

	for ip, inuse := range dhcp.table {
		if !inuse {
			dhcp.table[ip] = true
			return ip, nil
		}
	}

	return "", errNotAvaliableIP
}

func (dhcp *DHCP) RecycleIP(ip string) {
	dhcp.Lock()
	defer dhcp.Unlock()
	dhcp.table[ip] = false
}
