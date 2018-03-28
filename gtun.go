package main

import (
	"encoding/binary"
	"flag"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/ICKelin/glog"
	"github.com/songgao/water"
)

var (
	psrv = flag.String("s", "120.25.214.63:9621", "srv address")
	pdev = flag.String("dev", "gtun", "local tun device name")
)

func main() {
	flag.Parse()

	cfg := water.Config{
		DeviceType: water.TUN,
	}
	cfg.Name = *pdev
	ifce, err := water.New(cfg)

	if err != nil {
		glog.ERROR(err)
		return
	}

	conn, err := ConServer(*psrv)
	if err != nil {
		glog.ERROR(err)
		return
	}

	go IfRead(ifce, conn)
	go IfWrite(ifce, conn)

	sig := make(chan os.Signal, 3)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGABRT, syscall.SIGHUP)
	<-sig
}

func ConServer(srv string) (conn net.Conn, err error) {
	conn, err = net.Dial("tcp", srv)
	if err != nil {
		return nil, err
	}
	return conn, err
}

func IfRead(ifce *water.Interface, conn net.Conn) {
	packet := make([]byte, 2000)
	for {
		n, err := ifce.Read(packet)
		if err != nil {
			glog.ERROR(err)
			break
		}

		err = ForwardSrv(conn, packet[:n])
		if err != nil {
			glog.ERROR(err)
		}
	}
}

func IfWrite(ifce *water.Interface, conn net.Conn) {
	packet := make([]byte, 2000)
	for {
		nr, err := conn.Read(packet)
		if err != nil {
			glog.ERROR(err)
			break
		}

		_, err = ifce.Write(packet[:nr])
		if err != nil {
			glog.ERROR(err)
		}
	}
}

func ForwardSrv(srvcon net.Conn, buff []byte) (err error) {
	output := make([]byte, 0)
	bsize := make([]byte, 4)
	binary.BigEndian.PutUint32(bsize, uint32(len(buff)))

	output = append(output, bsize...)
	output = append(output, buff...)

	left := len(output)
	for left > 0 {
		nw, er := srvcon.Write(output)
		if err != nil {
			err = er
		}

		left -= nw
	}

	return err
}
