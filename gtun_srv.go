package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/ICKelin/glog"
	"github.com/songgao/water"
)

var client = make(map[string]net.Conn)
var ippool = make(map[string]bool)

func init() {
	for i := 10; i < 250; i++ {
		ip := fmt.Sprintf("10.10.253.%d", i)
		ippool[ip] = false
	}
}

func main() {
	listener, err := net.Listen("tcp", ":9621")
	if err != nil {
		glog.ERROR(err)
		return
	}

	cfg := water.Config{
		DeviceType: water.TUN,
	}
	cfg.Name = "gtun_srv"
	ifce, err := water.New(cfg)

	if err != nil {
		glog.ERROR(err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			glog.ERROR(err)
			break
		}

		glog.INFO("accept gtun client")
		go HandleClient(conn, ifce)
	}
}

func HandleClient(conn net.Conn, ifce *water.Interface) {
	defer conn.Close()

	cip, err := SelectIP()
	if err != nil {
		glog.ERROR(err)
		return
	}

	defer RecycleIP(cip)

	client[cip] = conn
	defer delete(client, cip)

	if err := DHCP(conn, cip); err != nil {
		glog.ERROR("dhcp ip for client fail", err)
		return
	}

	buff := make([]byte, 65536)
	for {
		nr, err := conn.Read(buff)
		if err != nil {
			if err != io.EOF {
				glog.ERROR(err)
			}
			break
		}

		if nr < 25 {
			glog.ERROR("too short ippkt")
			continue
		}

		dst := fmt.Sprintf("%d.%d.%d.%d", buff[20], buff[21], buff[22], buff[23])

		c := client[dst]
		if c != nil {
			c.Write(buff[:nr])
		} else {
			_, err = ifce.Write(buff[4:nr])
			if err != nil {
				glog.ERROR(err)
			}
		}
	}
}

func DHCP(conn net.Conn, clientip string) (err error) {
	plen := make([]byte, 4)
	binary.BigEndian.PutUint32(plen, uint32(len(clientip)))

	payload := make([]byte, 0)
	payload = append(payload, plen...)
	payload = append(payload, []byte(clientip)...)

	_, err = conn.Write(payload)
	return err
}

func SelectIP() (ip string, err error) {
	for ip, v := range ippool {
		if v == false {
			ippool[ip] = true
			return ip, nil
		}
	}
	return "", fmt.Errorf("not enough ip in pool")
}

func RecycleIP(cip string) {
	ippool[cip] = false
}
