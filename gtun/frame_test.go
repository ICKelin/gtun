package gtun

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/songgao/water"
	"github.com/stretchr/testify/assert"
)

func TestFrameInvalid(t *testing.T) {
	type TestCase struct {
		input   []byte
		invalid bool
	}

	var units = []TestCase{
		TestCase{
			input:   []byte{},
			invalid: true,
		},
		TestCase{
			input:   []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x08, 0x00},
			invalid: false,
		},
		TestCase{
			input:   []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x08, 0x00, 0x04},
			invalid: false,
		},
	}

	for _, unit := range units {
		got := Frame(unit.input).Invalid()
		assert.Equal(t, unit.invalid, got, fmt.Sprintf("expected %v got %v", unit.invalid, got))
	}
}

func TestIPV4(t *testing.T) {
	type TestCase struct {
		input []byte
		ipv4  bool
	}

	var units = []TestCase{
		TestCase{
			input: []byte{},
			ipv4:  false,
		},
		TestCase{
			input: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x08, 0x00},
			ipv4:  true,
		},
		TestCase{
			input: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x08, 0x00, 0x04},
			ipv4:  true,
		},
	}

	for _, unit := range units {
		got := Frame(unit.input).IsIPV4()
		assert.Equal(t, unit.ipv4, got, fmt.Sprintf("%v expected %v got %v", unit.input, unit.ipv4, got))
	}
}

func TestDNAT(t *testing.T) {
	iface, err := water.New(water.Config{DeviceType: water.TUN})
	assert.Equal(t, nil, err)

	err = setdev(iface.Name())
	assert.Equal(t, nil, err)

	buff := make([]byte, 65536)
	go func() {
		addr, _ := net.ResolveUDPAddr("udp", ":80")
		listener, err := net.ListenUDP("udp", addr)
		assert.Equal(t, nil, err)
		buff := make([]byte, 1024)
		for {
			nr, _, err := listener.ReadFromUDP(buff)
			assert.Equal(t, nil, err)
			fmt.Println(string(buff[:nr]))
		}
	}()

	go func() {
		for {
			nr, err := iface.Read(buff)
			assert.Equal(t, nil, err)
			// fmt.Println("before:", buff[:nr])
			p := Packet(buff[:nr])
			np := p.DNAT(80, parseip("100.64.240.10"))
			// fmt.Println("after: ", np)
			iface.Write(np)
		}
	}()

	for {
		conn, err := net.Dial("udp", "100.64.240.11:80")
		if err == nil {
			conn.Write([]byte("hello world"))
			conn.Close()
		}
		time.Sleep(time.Second * 2)
	}
}

func setdev(name string) error {
	ip := "100.64.240.10"
	gw := "100.64.240.1"

	cmdlist := make([]*CMD, 0)
	cmdlist = append(cmdlist, &CMD{cmd: "ifconfig", args: []string{name, "up"}})

	args := strings.Split(fmt.Sprintf("%s %s %s", name, ip, ip), " ")
	cmdlist = append(cmdlist, &CMD{cmd: "ifconfig", args: args})

	args = strings.Split(fmt.Sprintf("add -net %s/24 %s", gw, ip), " ")
	cmdlist = append(cmdlist, &CMD{cmd: "route", args: args})

	for _, c := range cmdlist {
		_, err := exec.Command(c.cmd, c.args...).CombinedOutput()
		if err != nil {
			return err
		}
	}
	return nil
}

func TestSNAT(t *testing.T) {

}

func parseip(strip string) uint32 {
	s := strings.Split(strip, ".")
	i0, _ := strconv.Atoi(s[0])
	i1, _ := strconv.Atoi(s[1])
	i2, _ := strconv.Atoi(s[2])
	i3, _ := strconv.Atoi(s[3])

	return uint32(i0<<24) + uint32(i1<<16) + uint32(i2<<8) + uint32(i3)
}
