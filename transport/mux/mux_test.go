package mux

import (
	"fmt"
	"testing"
	"time"

	"github.com/ICKelin/gtun/transport"
)

func TestYamux(t *testing.T) {
	lis, err := Listen("127.0.0.1:50051")
	if err != nil {
		t.Error(err)
		return
	}

	go func() {
		for {
			conn, err := lis.Accept()
			if err != nil {
				t.Error(err)
				return
			}

			go func() {
				defer conn.Close()
				count := 0
				for {
					stream, err := conn.AcceptStream()
					if err != nil {
						break
					}
					count += 1
					fmt.Println("Accept stream ", count)
					go func(s transport.Stream) {
						stream.Close()
					}(stream)
				}
			}()
		}
	}()
	dialer := &Dialer{}
	conn, err := dialer.Dial("127.0.0.1:50051")
	if err != nil {
		t.Log(err)
		return
	}
	defer conn.Close()
	for i := 0; i < 1000; i++ {
		stream, err := conn.OpenStream()
		if err != nil {
			t.Error(err)
			break
		}
		stream.Close()
		time.Sleep(time.Millisecond * 100)
	}
}
