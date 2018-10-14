package god

import (
	"testing"

	"time"
)

func TestGtunServer(t *testing.T) {
	g := NewGtun(&gtunConfig{
		Listener: "127.0.0.1:2002",
		Tokens:   []string{"abcdefg"},
	})

	go g.Run()
	time.Sleep(time.Second * 1)
	server(t)
}
