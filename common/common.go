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

type C2SAuthorize struct {
	AccessIP string `json:"access_ip"`
	Key      string `json:"key"`
}

type S2CAuthorize struct {
	Status         string   `json:"status"`
	AccessIP       string   `json:"access_ip"`
	RouteRule      []string `json:"route_rule"`
	Nameservers    []string `json:"nameservers"`
	Gateway        string   `json:"gateway"`
	RouteScriptUrl string   `json:"route_script_url"`
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

const (
	C2C_DATA = byte(0x00)

	C2S_DATA = byte(0x01)
	S2C_DATA = byte(0x02)

	C2S_HEARTBEAT = byte(0x03)
	S2C_HEARTBEAT = byte(0x04)

	C2S_AUTHORIZE = byte(0x05)
	S2C_AUTHORIZE = byte(0x06)
)
