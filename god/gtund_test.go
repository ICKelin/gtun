package god

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/ICKelin/gtun/common"
	"github.com/stretchr/testify/assert"
)

func TestGtund(t *testing.T) {
	go server(t)
	time.Sleep(time.Second * 1)

	conn, err := net.Dial("tcp", "127.0.0.1:9876")
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, conn)

	reg := &common.S2GRegister{}
	bytes, err := json.Marshal(reg)
	assert.Equal(t, nil, err)

	regBytes, err := common.Encode(common.S2G_REGISTER, bytes)
	assert.Equal(t, nil, err)

	conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
	n, err := conn.Write(regBytes)
	conn.SetWriteDeadline(time.Time{})
	assert.Equal(t, nil, err)
	assert.Equal(t, len(regBytes), n)

	for {
		bytes, err := common.Encode(common.S2G_HEARTBEAT, nil)
		assert.Equal(t, nil, err)
		conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
		nw, err := conn.Write(bytes)
		conn.SetWriteDeadline(time.Time{})
		assert.Equal(t, nil, err)
		assert.Equal(t, nw, len(bytes))
		time.Sleep(time.Second * 5)
	}
}

func server(t *testing.T) {
	d := NewGtund(&gtundConfig{Listener: "127.0.0.1:9876"})
	d.Run()
}
