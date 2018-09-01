package gtund

import (
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBroadcast(t *testing.T) {
	sndqueue := make(chan *GtunClientContext)
	forward := NewForward()

	listener, err := net.Listen("tcp", ":9090")
	if err != nil {
		t.Error(err)
	}

	conns := new(sync.Map)
	go send(t, conns, sndqueue)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				t.Error(err)
				break
			}

			conns.Store(conn, true)
			forward.Add(conn.RemoteAddr().String(), conn)
			go onConn(conn, sndqueue, forward)
		}
	}()

	clients(t)

}

func clients(t *testing.T) {
	for i := 0; i < 100; i++ {
		go func() {
			_, err := net.Dial("tcp", "127.0.0.1:9090")
			if err != nil {
				t.Error(err)
			}

		}()
	}
}

func onConn(conn net.Conn, sndqueue chan *GtunClientContext, forward *Forward) {
	defer conn.Close()
	forward.Broadcast(sndqueue, []byte{})
}

func send(t *testing.T, conns *sync.Map, sndqueue chan *GtunClientContext) {
	for {
		ctx := <-sndqueue
		assert.NotEqual(t, nil, ctx.conn)
		val, ok := conns.Load(ctx.conn)
		assert.Equal(t, true, ok)
		assert.Equal(t, true, val.(bool))
	}
}
