package gtund

import (
	"testing"
)

func TestGod(t *testing.T) {
	option = &Options{
		listenAddr: "127.0.0.1:9876",
	}
	g := NewRegistry(&RegistryConfig{
		Addr:  "120.25.214.63:9623",
		Token: "gtun-sg-token",
	}, &Service{})

	g.Run()
}

// func TestUpdateClientCount(t *testing.T) {
// 	option = &Options{
// 		listenAddr: "127.0.0.1:9876",
// 	}
// 	g := NewGod(&GodConfig{
// 		GodAddr:  "127.0.0.1:9876",
// 		GodToken: "gtun-sg-token",
// 	})

// 	go g.Run()

// 	for {
// 		time.Sleep(time.Second * 1)
// 		err := g.UpdateClientCount(1)
// 		assert.Equal(t, nil, err)
// 	}
// }
