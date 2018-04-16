package common

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

type C2SAuthorize struct {
	AccessIP string `json:"access_ip"`
	Key      string `json:"key"`
}
type S2CAuthorize struct {
	Status   string `json:"status"`
	AccessIP string `json:"access_ip"`
}

func Encode(cmd byte, payload []byte) []byte {
	buff := make([]byte, 0)

	plen := make([]byte, 4)
	binary.BigEndian.PutUint32(plen, uint32(len(payload))+1)

	buff = append(buff, plen...)
	buff = append(buff, cmd)
	buff = append(buff, payload...)

	return buff
}

func Decode(conn net.Conn) (byte, []byte, error) {
	plen := make([]byte, 4)
	conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	_, err := conn.Read(plen)
	conn.SetReadDeadline(time.Time{})
	if err != nil {
		return 0, nil, err
	}

	payloadlength := binary.BigEndian.Uint32(plen)
	resp := make([]byte, payloadlength)
	conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	nr, err := conn.Read(resp)
	conn.SetReadDeadline(time.Time{})
	if err != nil {
		return 0, nil, err
	}

	if nr < 1 {
		return 0, nil, fmt.Errorf("invalid pkt")
	}

	return resp[0], resp[1:nr], nil
}

const (
	C2C_DATA = byte(0x00)

	C2S_DATA = byte(0x01)
	S2C_DATA = byte(0x02)

	C2S_HEARTBEAT = byte(0x03)
	S2C_HEARTBEAT = byte(0x04)

	C2S_AUTHORIZE = byte(0x05)
	S2C_AUTHORIZE = byte(0x06)
)
