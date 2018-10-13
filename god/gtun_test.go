package god

import (
	"testing"
)

func TestGtunServer(t *testing.T) {
	g := NewGtun(&gtunConfig{
		Listener: "127.0.0.1:2002",
		Tokens:   []string{"abcdefg"},
	})

	g.Run()
}
