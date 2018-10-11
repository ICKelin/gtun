package gtund

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGod(t *testing.T) {
	option = &Options{
		listenAddr: "127.0.0.1:9876",
	}
	g := NewGod(&GodConfig{
		GodAddr:  "127.0.0.1:9876",
		GodToken: "gtun-sg-token",
	})

	g.Run()
}

func TestUpdateClientCount(t *testing.T) {
	option = &Options{
		listenAddr: "127.0.0.1:9876",
	}
	g := NewGod(&GodConfig{
		GodAddr:  "127.0.0.1:9876",
		GodToken: "gtun-sg-token",
	})

	go g.Run()

	for {
		time.Sleep(time.Second * 1)
		err := g.UpdateClientCount(1)
		assert.Equal(t, nil, err)
	}
}
