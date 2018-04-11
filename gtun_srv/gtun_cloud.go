package main

import (
	"fmt"
	"io"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/ICKelin/glog"
	"github.com/ICKelin/gtun/common"
	"github.com/songgao/water"
)

func CloudMode(device, lip, listenAddr string) {
	cfg := water.Config{
		DeviceType: water.TUN,
	}

	cfg.Name = device
	ifce, err := water.New(cfg)
	if err != nil {
		glog.ERROR(err)
		return
	}

	err = SetTunIP(cfg.Name, lip)
	if err != nil {
		glog.ERROR(err)
		return
	}

	laddr, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		glog.ERROR(err)
		return
	}

	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		glog.ERROR(err)
		return
	}

	go IfceWrite(ifce)

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			glog.ERROR(err)
			break
		}

		conn.SetKeepAlive(true)
		conn.SetKeepAlivePeriod(time.Second * 30)
		go IfceRead(ifce, conn)
	}
}

func IfceRead(ifce *water.Interface, conn net.Conn) {
	defer conn.Close()

	accessip, err := Authorize(conn)
	if err != nil {
		glog.ERROR(err)
		return
	}

	clientpool.Add(accessip, conn)
	defer clientpool.Del(accessip)

	glog.INFO("accept cloud client from", conn.RemoteAddr().String(), "assign ip", accessip)
	defer dhcppool.RecycleIP(accessip)

	buff := make([]byte, 65536)
	for {
		nr, err := conn.Read(buff)
		if err != nil {
			glog.ERROR(err)
			break
		}

		_, err = ifce.Write(buff[4:nr])
		if err != nil {
			glog.ERROR(err)
		}
	}
}

func IfceWrite(ifce *water.Interface) {
	buff := make([]byte, 65536)
	for {
		nr, err := ifce.Read(buff)
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

		// TODO: remove
		dst := fmt.Sprintf("%d.%d.%d.%d", buff[16], buff[17], buff[18], buff[19])
		c := clientpool.Get(dst)
		if c != nil {
			c.SetWriteDeadline(time.Now().Add(time.Second * 10))
			bytes := common.Encode(buff[:nr])
			_, err = c.Write(bytes)
			c.SetWriteDeadline(time.Time{})
			if err != nil {
				glog.ERROR("write to peer ", c.RemoteAddr().String(), dst, err)
			}
		} else {
			glog.ERROR(dst, "offline")
		}
	}
}

func SetTunIP(dev, tunip string) (err error) {
	type CMD struct {
		cmd  string
		args []string
	}

	cmdlist := make([]*CMD, 0)

	cmdlist = append(cmdlist, &CMD{cmd: "ifconfig", args: []string{dev, "up"}})

	switch runtime.GOOS {
	case "linux":
		args := strings.Split(fmt.Sprintf("addr add %s/24 dev %s", tunip, dev), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "ip", args: args})

	case "darwin":
		args := strings.Split(fmt.Sprintf("%s %s %s", dev, tunip, tunip), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "ifconfig", args: args})

		sp := strings.Split(tunip, ".")

		gateway := fmt.Sprintf("%s.%s.%s.0/24", sp[0], sp[1], sp[2])

		args = strings.Split(fmt.Sprintf("add -net %s/24 %s", gateway, tunip), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "route", args: args})

	default:
	}

	for _, c := range cmdlist {
		output, err := exec.Command(c.cmd, c.args...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("run %s error %s", c, string(output))
		}
	}

	return nil
}
