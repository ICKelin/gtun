package kcpclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	cfg := `
	{
		"localaddr":":2000",
		"remoteaddr":"127.0.0.1:2002",
		"key":"ICKelin-gtun-tunnel",
		"crypt":"xor",
		"mode":"fast",
		"mtu":1350,
		"sndwnd":1024,
		"rcvwnd":1024,
		"datashard":10,
		"parityshard":3,
		"dscp":0,
		"nocomp":false,
		"acknodelay":false,
		"nodelay": false,
		"interval":50,
		"resend":0,
		"nc": 0,
		"sockbuf":4194304,
		"keepalive":10,
		"snmpperiod":60,
		"conn": 1
	}
	`
	config, err := parseConfig([]byte(cfg))
	assert.Equal(t, nil, err)
	KCPClient(config)
}
