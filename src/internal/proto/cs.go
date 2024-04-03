package proto

import (
	"encoding/binary"
	"encoding/json"
)

const (
	CmdAuth = iota
	CmdHeartbeat
	CmdData
)

type C2SAuth struct {
	Key string `json:"key"`
}

type S2CAuth struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ProxyProtocol struct {
	Protocol string `json:"protocol"`
	SrcIP    string `json:"sip"`
	SrcPort  string `json:"sport"`
	DstIP    string `json:"dip"`
	DstPort  string `json:"dport"`
}

func EncodeData(raw []byte) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(len(raw)))
	buf = append(buf, raw...)
	return buf
}

func EncodeProxyProtocol(protocol, sip, sport, dip, dport string) []byte {
	proxyProtocol := &ProxyProtocol{
		Protocol: protocol,
		SrcIP:    sip,
		SrcPort:  sport,
		DstIP:    dip,
		DstPort:  dport,
	}

	body, _ := json.Marshal(proxyProtocol)
	return EncodeData(body)
}
