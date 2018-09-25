package common

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

const (
	MAX_PAYLOAD = 1<<16 - 1
)

var version = "1.1.0"

func Version() string {
	return version
}

func Encode(cmd byte, payload []byte) ([]byte, error) {
	buff := make([]byte, 0)

	if len(payload) > MAX_PAYLOAD {
		return nil, fmt.Errorf("too big payload")
	}

	plen := make([]byte, 2)
	binary.BigEndian.PutUint16(plen, uint16(len(payload))+1)
	buff = append(buff, plen...)
	buff = append(buff, cmd)
	buff = append(buff, payload...)

	return buff, nil
}

func Decode(conn net.Conn) (byte, []byte, error) {
	plen := make([]byte, 2)
	conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	_, err := io.ReadFull(conn, plen)
	conn.SetReadDeadline(time.Time{})
	if err != nil {
		return 0, nil, err
	}

	payloadlength := binary.BigEndian.Uint16(plen)
	if payloadlength > 65535 {
		return 0, nil, fmt.Errorf("too big ippkt size %d", payloadlength)
	}

	resp := make([]byte, payloadlength)
	nr, err := io.ReadFull(conn, resp)
	if err != nil {
		return 0, nil, err
	}

	if nr < 1 {
		return 0, nil, fmt.Errorf("invalid pkt")
	}

	if nr != int(payloadlength) {
		return resp[0], resp[1:nr], fmt.Errorf("invalid payloadlength %d %d", nr, int(payloadlength))
	}

	return resp[0], resp[1:nr], nil
}
