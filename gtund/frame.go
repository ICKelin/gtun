package gtund

import "fmt"

type Frame []byte
type Packet []byte

func (f Frame) Invalid() bool {
	return len(f) < 14
}

func (f Frame) IsIPV4() bool {
	proto := int(f[12])<<8 + int(f[13])
	return proto == 0x0800
}

func (p Packet) Invalid() bool {
	return len(p) < 20
}

func (p Packet) Version() int {
	return int((p[0] >> 4))
}

func (p Packet) Dst() string {
	return fmt.Sprintf("%d.%d.%d.%d", p[16], p[17], p[18], p[19])
}
