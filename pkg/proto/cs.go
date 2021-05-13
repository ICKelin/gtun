package proto

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
)

const (
	CmdAuth = iota
	CmdHeartbeat
	CmdData
)

type S2CHeartbeat struct {
}

type C2SHeartbeat struct{}

type C2SAuth struct {
	Key     string        `json:"key" yaml:"key"`
	Domain  string        `json:"domain" yaml:"domain"`
	Forward []ForwardItem `json:"forwards" yaml:"forwards"`
}

type ForwardItem struct {
	// forward protocol. eg: tcp, udp, https, http
	Protocol string `json:"protocol" yaml:"protocol"`

	// forward ports
	// key is the port opennotrd listen
	// value is local port
	Ports map[int]string `json:"ports" yaml:"ports"`

	// local ip, default is 127.0.0.1
	// the traffic will be forward to $LocalIP:$LocalPort
	// for example: 127.0.0.1:8080. 192.168.31.65:8080
	LocalIP string `json:"localIP" yaml:"localIP"`

	// raw config pass to server
	RawConfig string `json:"rawConfig" yaml:"rawConfig"`
}

type S2CAuth struct {
	Domain string `json:"domain"` // 分配域名
	Vip    string `json:"vip"`    // 分配虚拟ip地址
}

type ProxyProtocol struct {
	Protocol string `json:"protocol"`
	SrcIP    string `json:"sip"`
	SrcPort  string `json:"sport"`
	DstIP    string `json:"dip"`
	DstPort  string `json:"dport"`
}

// 1字节版本
// 1字节命令
// 2字节长度
type Header [4]byte

func (h Header) Version() int {
	return int(h[0])
}

func (h Header) Cmd() int {
	return int(h[1])
}

func (h Header) Bodylen() int {
	return (int(h[2]) << 8) + int(h[3])
}

func Read(conn net.Conn) (Header, []byte, error) {
	h := Header{}
	_, err := io.ReadFull(conn, h[:])
	if err != nil {
		return h, nil, err
	}

	bodylen := h.Bodylen()
	if bodylen <= 0 {
		return h, nil, nil
	}

	body := make([]byte, bodylen)
	_, err = io.ReadFull(conn, body)
	if err != nil {
		return h, nil, err
	}

	return h, body, nil
}

func Write(conn net.Conn, cmd int, body []byte) error {
	bodylen := make([]byte, 2)
	binary.BigEndian.PutUint16(bodylen, uint16(len(body)))

	hdr := []byte{0x01, byte(cmd)}
	hdr = append(hdr, bodylen...)

	writebody := make([]byte, 0)
	writebody = append(writebody, hdr...)
	writebody = append(writebody, body...)
	_, err := conn.Write(writebody)
	return err
}

func WriteJSON(conn net.Conn, cmd int, obj interface{}) error {
	body, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return Write(conn, cmd, body)
}

func ReadJSON(conn net.Conn, obj interface{}) error {
	_, body, err := Read(conn)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, obj)
}
