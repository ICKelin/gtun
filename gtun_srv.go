package main

import (
	"io"
	"net"

	"github.com/ICKelin/glog"
)

var client = make([]net.Conn, 0)

func main() {
	listener, err := net.Listen("tcp", ":9621")
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

		client = append(client, conn)
		glog.INFO("accept gtun client")
		go HandleClient(conn)
	}
}

func HandleClient(conn net.Conn) {
	defer conn.Close()

	buff := make([]byte, 65536)
	for {
		nr, err := conn.Read(buff)
		if err != nil {
			if err != io.EOF {
				glog.ERROR(err)
			}
			break
		}

		// broadcast
		// TODO select peer
		for _, c := range client {
			if c.RemoteAddr().String() != conn.RemoteAddr().String() {
				c.Write(buff[:nr])
			}
		}
	}
}
