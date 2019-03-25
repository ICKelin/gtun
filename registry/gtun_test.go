package registry

import (
	"testing"

	"time"

	"github.com/ICKelin/gtun/registry/config"
)

func TestGtunServer(t *testing.T) {
	g := NewGtun(&config.GtunConfig{
		Listener: "127.0.0.1:2002",
		Tokens:   []string{"abcdefg"},
	})

	go g.Run()
	time.Sleep(time.Second * 1)
	server(t)
}
