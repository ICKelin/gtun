package gtund

import (
	"testing"
)

func TestGod(t *testing.T) {
	g := NewGod(&GodConfig{
		GodAddr: "127.0.0.1:9876",
	})

	g.Run()
}
