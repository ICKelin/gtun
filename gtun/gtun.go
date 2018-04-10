package main

import (
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

type GtunContext struct {
	conn   net.Conn
	srv    string
	key    string
	dhcpip string
}

func (this *GtunContext) ConServer() (err error) {
	conn, err := ConServer(this.srv)
	if err != nil {
		return err
	}

	s2c, err := Authorize(conn, this.dhcpip, this.key)
	if err != nil {
		return err
	}

	this.dhcpip = s2c.AccessIP
	this.conn = conn

	return nil
}

func (this *GtunContext) ReConServer() (err error) {
	return this.ConServer()
}

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

	gtun := &GtunContext{
		srv: *psrv,
		key: *pkey,
	}

	err = gtun.ConServer()
	if err != nil {
		glog.ERROR(err)
		return
	}

	err = SetTunIP(*pdev, gtun.dhcpip)
	if err != nil {
		glog.ERROR(err)
		return
	}

	go IfaceRead(ifce, gtun)
	go IfaceWrite(ifce, gtun)

	sig := make(chan os.Signal, 3)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGABRT, syscall.SIGHUP)
	<-sig
}

func IfaceRead(ifce *water.Interface, gtun *GtunContext) {
	packet := make([]byte, 65536)
	for {
		n, err := ifce.Read(packet)
		if err != nil {
			glog.ERROR(err)
			break
		}

		bytes := common.Encode(packet[:n])
		gtun.conn.SetWriteDeadline(time.Now().Add(time.Second * 30))
		_, err = gtun.conn.Write(bytes)
		gtun.conn.SetWriteDeadline(time.Time{})

		if err != nil {
			glog.INFO("reconnect since", err)
			gtun.ReConServer()
			continue
		}
	}
}

func IfaceWrite(ifce *water.Interface, gtun *GtunContext) {
	for {
		pkt, err := common.Decode(gtun.conn)
		if err != nil {
			glog.INFO("reconnect since", err)
			gtun.ReConServer()
			continue
		}

		_, err = ifce.Write(pkt)
		if err != nil {
			glog.ERROR(err)
		}
	}
}

func ConServer(srv string) (conn net.Conn, err error) {
	srvaddr, err := net.ResolveTCPAddr("tcp", srv)
	if err != nil {
		return nil, err
	}

	for {
		conn, err = net.DialTCP("tcp", nil, srvaddr)
		if err != nil {
			glog.ERROR(err)
			time.Sleep(time.Second * 3)
			continue
		}
		return conn, nil
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

	buff := common.Encode(payload)

	_, err = conn.Write(buff)
	if err != nil {
		return nil, err
	}

	resp, err := common.Decode(conn)
	if err != nil {
		return nil, err
	}

	s2cauthorize = &common.S2CAuthorize{}
	err = json.Unmarshal(resp, s2cauthorize)
	if err != nil {
		return nil, err
	}

	return s2cauthorize, nil
}
