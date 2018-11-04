package kcpserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	cnt := `
		{
			"listen": ":2010",
			"target":"127.0.0.1:2012",
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
			"nodelay": false,
			"interval":50,
			"resend":0,
			"nc": 0,
			"sockbuf":4194304,
			"keepalive":10,
			"snmpperiod":60
		}
	`
	config, err := parseConfig([]byte(cnt))
	assert.Equal(t, nil, err)
	KCPServer(config)
}
