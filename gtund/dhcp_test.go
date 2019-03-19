package gtund

import (
	"fmt"
	"testing"
)

var (
	poolCount = 240
)

func TestNewDHCP(t *testing.T) {
	dhcp, err := NewDHCP(&DHCPConfig{
		CIDR: "192.168.10.1/24",
	})

	if err != nil {
		t.Error(err)
		return
	}

	t.Log(dhcp.gateway, dhcp.mask)
	for ip, _ := range dhcp.table {
		fmt.Println(ip)
	}
}

func TestSelect(t *testing.T) {
	cfg := &DHCPConfig{
		CIDR: "192.168.10.1/24",
	}

	dhcp, err := NewDHCP(cfg)
	if err != nil {
		t.Error(err)
		return
	}

	for i := 0; i < poolCount; i++ {
		ip, err := dhcp.SelectIP()
		if err != nil {
			t.Error(err)
			return
		}

		if dhcp.table[ip] == false {
			t.Errorf("ip %s expect to be in used, got not", ip)
			return
		}
	}
}

// func TestRecycle(t *testing.T) {
// 	cfg := &DHCPConfig{
// 		gateway: "192.168.6.1",
// 	}

// 	dhcp, err := NewDHCP(cfg)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	for i := 0; i < poolCount; i++ {
// 		ip, err := dhcp.SelectIP("")
// 		assert.Equal(t, nil, err)
// 		assert.NotEqual(t, "", ip)
// 		assert.Equal(t, true, dhcp.InUsed(ip))
// 		dhcp.RecycleIP(ip)
// 		assert.Equal(t, false, dhcp.InUsed(ip))
// 	}
// }
