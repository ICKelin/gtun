package main

import "testing"

func TestBootGtun(t *testing.T) {
	ReloadGtun([]*GtunConfigItem{
		{
			Region:          "CN",
			Scheme:          "kcp",
			ServerIP:        "127.0.0.1",
			ServerTracePort: 3003,
			ServerPort:      3002,
			ListenPort:      8524,
			Rate:            50,
		},
		{
			Region:          "US",
			Scheme:          "kcp",
			ServerIP:        "127.0.0.1",
			ServerTracePort: 4003,
			ServerPort:      4002,
			ListenPort:      8525,
			Rate:            50,
		},
	})
}
