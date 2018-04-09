package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/ICKelin/glog"
	"github.com/ICKelin/gtun/common"
	"github.com/songgao/water"
)

var (
	psrv = flag.String("s", "120.25.214.63:9621", "srv address")
	pdev = flag.String("dev", "gtun", "local tun device name")
	pkey = flag.String("key", "gtun_authorize", "client authorize key")
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
	defer conn.Close()

	s2cauthorize, err := Authorize(conn, "", *pkey)
	if err != nil {
		glog.ERROR("authorize fail")
		return
	}

	glog.INFO("authorize success...")

	err = SetTunIP(*pdev, s2cauthorize.AccessIP)
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
		if er != nil {
			err = er
			break
		}

		left -= nw
	}
	return err
}

func ConServer(srv string) (conn net.Conn, err error) {
	srvaddr, err := net.ResolveTCPAddr("tcp", srv)
	if err != nil {
		return nil, err
	}

	tryCount := 0
	for {
		conn, err = net.DialTCP("tcp", nil, srvaddr)
		if err != nil {
			if tryCount > 10 {
				return nil, err
			}
			tryCount++
			time.Sleep(time.Second * 3)
			continue
		}
		return conn, err
	}
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

func Authorize(conn net.Conn, accessIP, key string) (s2cauthorize *common.S2CAuthorize, err error) {
	c2sauthorize := &common.C2SAuthorize{
		AccessIP: accessIP,
		Key:      key,
	}

	payload, err := json.Marshal(c2sauthorize)
	if err != nil {
		return nil, err
	}

	plen := make([]byte, 4)
	binary.BigEndian.PutUint32(plen, uint32(len(payload)))

	buff := make([]byte, 0)
	buff = append(buff, plen...)

	buff = append(buff, payload...)

	_, err = conn.Write(buff)
	if err != nil {
		return nil, err
	}

	_, err = conn.Read(plen)
	if err != nil {
		return nil, err
	}

	payloadlength := uint32(0)
	payloadlength = binary.BigEndian.Uint32(plen)
	resp := make([]byte, payloadlength)
	nr, err := conn.Read(resp)
	if err != nil {
		return nil, err
	}

	s2cauthorize = &common.S2CAuthorize{}
	err = json.Unmarshal(resp[:nr], s2cauthorize)
	if err != nil {
		return nil, err
	}

	return s2cauthorize, nil
}
