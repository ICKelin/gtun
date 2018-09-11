package gtun

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

var (
	sumoff    = 10
	sipoff    = 12
	dipoff    = 16
	protooff  = 9
	tcpsumoff = 36
	udpsumoff = 26
)

type Frame []byte
type Packet []byte

type psedoheader struct {
	sip    uint32
	dip    uint32
	zero   uint8
	proto  uint8
	length uint16
}

func (f Frame) Invalid() bool {
	return len(f) < 14
}

func (f Frame) IsIPV4() bool {
	if f.Invalid() {
		return false
	}
	proto := int(f[12])<<8 + int(f[13])
	return proto == 0x0800
}

func (p Packet) Invalid() bool {
	return len(p) < 20
}

func (p Packet) Copy() Packet {
	np := make([]byte, len(p))
	copy(np, p)
	return np
}

func (p Packet) Version() int {
	return int((p[0] >> 4))
}

func (p Packet) Dst() string {
	return fmt.Sprintf("%d.%d.%d.%d", p[dipoff], p[dipoff+1], p[dipoff+2], p[dipoff+3])
}

func (p Packet) DNAT(toport uint16, toip uint32) Packet {
	np := p.Copy()
	np[dipoff] = byte(toip >> 24)
	np[dipoff+1] = byte(toip >> 16)
	np[dipoff+2] = byte(toip >> 8)
	np[dipoff+3] = byte(toip)

	// checksum
	np.bzeroSum()
	np.sum()

	return np
}

func (p Packet) SNAT(fromport uint16, fromip uint32) Packet {
	np := p.Copy()
	np[sipoff] = byte(fromip >> 24)
	np[sipoff+1] = byte(fromip >> 16)
	np[sipoff+2] = byte(fromip >> 8)
	np[sipoff+3] = byte(fromip)
	return np
}

func (p Packet) bzeroSum() {
	p[sumoff] = 0
	p[sumoff+1] = 0
}

func (p Packet) sum() {
	sum := checksum(p[:20])
	p[sumoff] = byte(sum >> 8)
	p[sumoff+1] = byte(sum)

	psedo := psedoheader{
		sip:    (uint32(p[sipoff]) << 24) + (uint32(p[sipoff+1]) << 16) + (uint32(p[sipoff+2]) << 8) + uint32(p[sipoff+3]),
		dip:    (uint32(p[dipoff]) << 24) + (uint32(p[dipoff+1]) << 16) + (uint32(p[dipoff+2]) << 8) + uint32(p[dipoff+3]),
		zero:   0,
		proto:  p[protooff],
		length: uint16(len(p[20:])),
	}

	var b bytes.Buffer
	binary.Write(&b, binary.BigEndian, &psedo)

	if p[protooff] == 17 {
		p[udpsumoff] = 0
		p[udpsumoff+1] = 0
		bb := append(b.Bytes(), []byte(p)[20:]...)
		tsum := checksum(bb)
		p[udpsumoff] = byte(tsum >> 8)
		p[udpsumoff+1] = byte(tsum)
		return
	}

	if p[protooff] == 6 {
		p[tcpsumoff] = 0
		p[tcpsumoff+1] = 0
		bb := append(b.Bytes(), []byte(p)[20:]...)
		tsum := checksum(bb)
		p[tcpsumoff] = byte(tsum >> 8)
		p[tcpsumoff+1] = byte(tsum)
		return
	}
}

func (p Packet) sumvalue() uint16 {
	return uint16(p[sumoff])<<8 + uint16(p[sumoff+1])
}

func checksum(buf []byte) uint16 {
	sum := uint32(0)

	for ; len(buf) >= 2; buf = buf[2:] {
		sum += uint32(buf[0])<<8 | uint32(buf[1])
	}
	if len(buf) > 0 {
		sum += uint32(buf[0]) << 8
	}
	for sum > 0xffff {
		sum = (sum >> 16) + (sum & 0xffff)
	}
	csum := ^uint16(sum)
	/*
	 * From RFC 768:
	 * If the computed checksum is zero, it is transmitted as all ones (the
	 * equivalent in one's complement arithmetic). An all zero transmitted
	 * checksum value means that the transmitter generated no checksum (for
	 * debugging or for higher level protocols that don't care).
	 */
	if csum == 0 {
		csum = 0xffff
	}
	return csum
}
