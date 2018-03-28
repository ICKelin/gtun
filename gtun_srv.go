package main

import (
	"encoding/binary"
	"io"
	"net"

	"github.com/ICKelin/glog"
	"github.com/songgao/water"
)

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

		go HandleClient(conn, ifce)
	}
}

func IfRead(ifce *water.Interface, conn net.Conn) {
	buff := make([]byte, 2048)
	for {
		nr, err := ifce.Read(buff)
		if err != nil {
			glog.ERROR(err)
			break
		}

		conn.Write(buff[:nr])
	}
}

func HandleClient(conn net.Conn, ifce *water.Interface) {
	defer conn.Close()

	go IfRead(ifce, conn)
	plen := make([]byte, 4)

	for {
		nr, err := conn.Read(plen)
		if err != nil {
			if err != io.EOF {
				glog.ERROR(err)
			}
			break
		}

		if nr != 4 {
			glog.ERROR("too short data", nr)
			continue
		}

		payloadLength := binary.BigEndian.Uint32(plen)

		buff := make([]byte, payloadLength)
		nr, err = conn.Read(buff)
		if err != nil {
			glog.ERROR(err)
			continue
		}

		if nr != int(payloadLength) {
			glog.ERROR("size no match ", nr, payloadLength)
			continue
		}

		_, err = ifce.Write(buff[:nr])
		if err != nil {
			glog.ERROR(err)
		}
	}
}
