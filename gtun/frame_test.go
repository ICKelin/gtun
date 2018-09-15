package gtun

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/songgao/water"
)

func TestDNAT(t *testing.T) {
	ip := "192.168.11.10"     // tun device ip
	gw := "192.168.11.1"      // gateway
	target := "120.25.214.63" // test origin target
	local := "10.10.201.134"  // dnat to
	tmpip := "192.168.11.11"

	localbyte := [4]byte{byte(parseip(local) >> 24), byte(parseip(local) >> 16), byte(parseip(local) >> 8), byte(parseip(local))}
	iface, err := water.New(water.Config{DeviceType: water.TUN})
	if err != nil {
		log.Fatal(err)
	}

	err = setdev(ip, gw, iface.Name(), target)

	// test server,tcp and udp
	go func() {
		addr, _ := net.ResolveUDPAddr("udp", ":58422")
		udpsrv, err := net.ListenUDP("udp", addr)
		if err != nil {
			log.Fatal(err)
		}
		buff := make([]byte, 1024)

		listener, err := net.Listen("tcp", ":58423")
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			for {
				nr, addr, err := udpsrv.ReadFromUDP(buff)
				if err != nil {
					log.Fatal(err)
					continue
				}
				fmt.Println("echo req:", string(buff[:nr]))
				udpsrv.WriteToUDP([]byte("pong"), addr)
			}
		}()

		go func() {
			for {
				conn, err := listener.Accept()
				if err != nil {
					log.Fatal(err)
				}

				nr, err := conn.Read(buff)
				if err != nil {
					conn.Close()
					continue
				}

				fmt.Println("tcp rcv: ", string(buff[:nr]))
				conn.Write([]byte("pong"))
				conn.Close()
			}
		}()

	}()

	// iface dnat
	go func() {
		fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
		if err != nil || fd < 0 {
			log.Fatal(err)
		}
		defer syscall.Close(fd)

		buff := make([]byte, 65536)
		for {
			nr, err := iface.Read(buff)
			if err != nil {
				log.Fatal(err)
			}

			p := Packet(buff[:nr])

			if p.Dst() == tmpip {
				np := p.SNAT(80, parseip(target))
				np = p.DNAT(80, parseip(local))

				tcp := np[np.Length():]
				port := (int(tcp[2]) << 8) + int(tcp[3])
				if np.protocol() == 17 {
					syscall.Sendto(fd, np, 0, &syscall.SockaddrInet4{Port: port, Addr: localbyte})
				}

				if np.protocol() == 6 {
					tcp := np[np.Length():]
					port := (int(tcp[2]) << 8) + int(tcp[3])
					syscall.Sendto(fd, np, 0, &syscall.SockaddrInet4{Port: port, Addr: localbyte})
				}
			} else {
				np := p.DNAT(80, parseip(local))
				np = p.SNAT(80, parseip(tmpip))
				_, err = iface.Write(np)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}()

	// test client
	for {
		udpbuff := make([]byte, 1024)
		go func() {

		}()

		conn, err := net.Dial("udp", "120.25.214.63:58422")
		if err != nil {
			t.Error(err)
		}

		udpconn := conn.(*net.UDPConn)
		for {
			udpconn.Write([]byte("ping"))
			conn.SetReadDeadline(time.Now().Add(time.Second * 2))
			nr, err := udpconn.Read(udpbuff)
			conn.SetReadDeadline(time.Time{})
			if err != nil {
				t.Error(err)
				continue
			}
			fmt.Println(string(udpbuff[:nr]))
			time.Sleep(time.Second * 2)
		}
	}
}

func setdev(ip, gw, name string, target string) error {
	cmdlist := make([]*CMD, 0)

	switch runtime.GOOS {
	case "linux":
		cmdlist = append(cmdlist, &CMD{cmd: "ifconfig", args: []string{name, "up"}})

		args := strings.Split(fmt.Sprintf("%s %s %s", name, ip, ip), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "ifconfig", args: args})

		args = strings.Split(fmt.Sprintf("addr add %s/24 dev %s", ip, name), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "ip", args: args})

		args = strings.Split(fmt.Sprintf("ro add %s dev %s", target, name), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "ip", args: args})

	case "darwin":
		cmdlist = append(cmdlist, &CMD{cmd: "ifconfig", args: []string{name, "up"}})

		args := strings.Split(fmt.Sprintf("%s %s %s", name, ip, ip), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "ifconfig", args: args})

		args = strings.Split(fmt.Sprintf("add -net %s/24 %s", gw, ip), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "route", args: args})

		args = strings.Split(fmt.Sprintf("add -net %s %s", target, ip), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "route", args: args})

	}
	for _, c := range cmdlist {
		_, err := exec.Command(c.cmd, c.args...).CombinedOutput()
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
	return nil
}

func parseip(strip string) uint32 {
	s := strings.Split(strip, ".")
	i0, _ := strconv.Atoi(s[0])
	i1, _ := strconv.Atoi(s[1])
	i2, _ := strconv.Atoi(s[2])
	i3, _ := strconv.Atoi(s[3])

	return uint32(i0<<24) + uint32(i1<<16) + uint32(i2<<8) + uint32(i3)
}
