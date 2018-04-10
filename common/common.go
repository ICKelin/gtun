package common

import (
	"encoding/binary"
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

func Encode(payload []byte) []byte {
	buff := make([]byte, 0)

	plen := make([]byte, 4)
	binary.BigEndian.PutUint32(plen, uint32(len(payload)))

	buff = append(buff, plen...)
	buff = append(buff, payload...)

	return buff
}

func Decode(conn net.Conn) ([]byte, error) {
	plen := make([]byte, 4)
	_, err := conn.Read(plen)
	if err != nil {
		return nil, err
	}

	payloadlength := binary.BigEndian.Uint32(plen)
	resp := make([]byte, payloadlength)

	conn.SetReadDeadline(time.Now().Add(time.Second * 30))
	nr, err := conn.Read(resp)
	conn.SetReadDeadline(time.Time{})
	if err != nil {
		return nil, err
	}

	return resp[:nr], nil
}
