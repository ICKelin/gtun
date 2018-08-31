package gtund

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	poolCount = 240
)

func TestSelect(t *testing.T) {
	cfg := &DHCPConfig{
		gateway: "192.168.6.1",
	}

	dhcp, err := NewDHCP(cfg)
	if err != nil {
		t.Error(err)
	}

	inuse := make(map[string]bool)
	for i := 0; i < poolCount; i++ {
		ip, err := dhcp.SelectIP("")
		assert.Equal(t, nil, err)
		assert.Equal(t, false, inuse[ip])
		assert.NotEqual(t, "", ip)
		assert.Equal(t, true, dhcp.InUsed(ip))
		inuse[ip] = true
	}

	ip, err := dhcp.SelectIP("")
	assert.NotEqual(t, nil, err)
	assert.Equal(t, "", ip)
}

func TestSelectWithPrefer(t *testing.T) {
	cfg := &DHCPConfig{
		gateway: "192.168.6.1",
	}

	dhcp, err := NewDHCP(cfg)
	if err != nil {
		t.Error(err)
	}

	inuse := make(map[string]bool)
	for i := 0; i < poolCount; i++ {
		prefer := fmt.Sprintf("192.168.6.%d", i+10)
		ip, err := dhcp.SelectIP(prefer)
		assert.Equal(t, nil, err)
		assert.Equal(t, false, inuse[ip])
		assert.Equal(t, true, dhcp.InUsed(ip))
		assert.Equal(t, prefer, ip)
		inuse[ip] = true
	}
}

func TestRecycle(t *testing.T) {
	cfg := &DHCPConfig{
		gateway: "192.168.6.1",
	}

	dhcp, err := NewDHCP(cfg)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < poolCount; i++ {
		ip, err := dhcp.SelectIP("")
		assert.Equal(t, nil, err)
		assert.NotEqual(t, "", ip)
		assert.Equal(t, true, dhcp.InUsed(ip))
		dhcp.RecycleIP(ip)
		assert.Equal(t, false, dhcp.InUsed(ip))
	}
}
