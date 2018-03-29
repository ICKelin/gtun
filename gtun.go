package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
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

	tunip, err := GetTunIP(conn)
	if err != nil {
		glog.ERROR(err)
		return
	}

	err = SetTunIP(*pdev, tunip)
	if err != nil {
		glog.ERROR(err)
		return
	}

	go IfaceRead(ifce, conn)
	go IfaceWrite(ifce, conn)

	sig := make(chan os.Signal, 3)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGABRT, syscall.SIGHUP)
	<-sig
}

func IfaceRead(ifce *water.Interface, conn net.Conn) {
	packet := make([]byte, 65536)
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

func IfaceWrite(ifce *water.Interface, conn net.Conn) {
	packet := make([]byte, 65536)
	for {
		nr, err := conn.Read(packet)
		if err != nil {
			glog.ERROR(err)
			break
		}

		_, err = ifce.Write(packet[4:nr])
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

func ConServer(srv string) (conn net.Conn, err error) {
	conn, err = net.Dial("tcp", srv)
	if err != nil {
		return nil, err
	}
	return conn, err
}

func GetTunIP(conn net.Conn) (tunip string, err error) {
	plen := make([]byte, 4)
	nr, err := conn.Read(plen)
	if err != nil {
		return "", err
	}

	if nr != 4 {
		return "", fmt.Errorf("too short pkt")
	}

	payloadlength := binary.BigEndian.Uint32(plen)
	buff := make([]byte, int(payloadlength))

	nr, err = conn.Read(buff)
	if err != nil {
		return "", err
	}

	return string(buff[:nr]), nil
}

func SetTunIP(dev, tunip string) (err error) {
	uptun := fmt.Sprintf("ifconfig %s up", dev)
	setip := fmt.Sprintf("ip addr add %s/24 dev %s", tunip, dev)

	err = exec.Command("/bin/sh", "-c", uptun).Run()
	if err != nil {
		return fmt.Errorf("up %s error %s", dev, err.Error())
	}

	err = exec.Command("/bin/sh", "-c", setip).Run()
	if err != nil {
		return fmt.Errorf("up %s error %s", dev, err.Error())
	}

	return nil
}
