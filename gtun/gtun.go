/*

MIT License

Copyright (c) 2018 ICKelin

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

*/

/*
	DESCRIPTION:
				This program is a gtun client for game/ip accelator.

	Author: ICKelin
*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ICKelin/glog"
	"github.com/ICKelin/gtun/common"
	"github.com/songgao/water"
)

var (
	psrv   = flag.String("s", "120.25.214.63:9621", "srv address")
	pkey   = flag.String("key", "gtun_authorize", "client authorize key")
	pdebug = flag.Bool("debug", false, "debug mode")
	ptap   = flag.Bool("tap", false, "tap mode")
)

func main() {
	flag.Parse()

	if *pdebug {
		glog.Init("gtun", glog.PRIORITY_DEBUG, "./", glog.OPT_DATE, 1024*10)
	} else {
		glog.Init("gtun", glog.PRIORITY_INFO, "./", glog.OPT_DATE, 1024*10)
	}

	if err := Login(); err != nil {
		glog.ERROR("Login fail")
		return
	}

	// 2018.04.25
	// Force using tun in !windows
	// Force using tap in windows
	ifce, err := NewIfce(*ptap)
	if err != nil {
		glog.ERROR(err)
		return
	}

	gtun := &GtunContext{
		srv:      *psrv,
		key:      *pkey,
		ldev:     ifce.Name(),
		sndqueue: make(chan []byte),
		rcvqueue: make(chan []byte),
		dhcpip:   "",
	}

	err = gtun.ConServer()
	if err != nil {
		glog.ERROR(err)
		return
	}

	err = SetDeviceIP(gtun)
	if err != nil {
		glog.ERROR(err)
		return
	}

	InsertRoute(gtun)

	go Heartbeat(gtun)
	go IfaceRead(ifce, gtun)

	go func() {
		for {
			wg := &sync.WaitGroup{}
			wg.Add(2)

			go Snd(ifce, gtun, wg)
			go Rcv(ifce, gtun, wg)

			wg.Wait()

			// reconnect
			for {
				err = gtun.ConServer()
				if err != nil {
					glog.ERROR(err)
					time.Sleep(time.Second * 3)
					continue
				}
				break
			}
			glog.INFO("reconnect success")
		}
	}()

	sig := make(chan os.Signal, 3)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGABRT, syscall.SIGHUP)
	<-sig
}

type GtunContext struct {
	sync.Mutex
	conn       net.Conn
	srv        string
	key        string
	dhcpip     string
	ldev       string
	route      []string // route info from gtun_srv
	nameserver []string // nameserver from gtun_srv UNUSED
	gateway    string   // gateway from gtun_srv
	sndqueue   chan []byte
	rcvqueue   chan []byte
}

func (this *GtunContext) ConServer() (err error) {
	conn, err := ConServer(this.srv)
	if err != nil {
		return fmt.Errorf("connect server %s", err.Error())
	}

	s2c, err := Authorize(conn, this.dhcpip, this.key)
	if err != nil {
		return fmt.Errorf("authorize fail %s", err.Error())
	}

	this.dhcpip = s2c.AccessIP
	this.conn = conn
	this.route = s2c.RouteRule
	this.nameserver = s2c.Nameservers
	this.gateway = s2c.Gateway
	return nil
}

func Rcv(ifce *water.Interface, gtun *GtunContext, wg *sync.WaitGroup) {
	defer wg.Done()
	defer gtun.conn.Close()

	for {
		cmd, pkt, err := common.Decode(gtun.conn)
		if err != nil {
			glog.INFO(err)
			break
		}
		switch cmd {
		case common.S2C_HEARTBEAT:
			glog.DEBUG("heartbeat from srv")

		case common.C2C_DATA:
			_, err := ifce.Write(pkt)
			if err != nil {
				glog.ERROR(err)
			}

		default:
			glog.INFO("unimplement cmd", int(cmd), pkt)
		}
	}
}

func Snd(ifce *water.Interface, gtun *GtunContext, wg *sync.WaitGroup) {
	defer wg.Done()
	defer gtun.conn.Close()

	for {
		pkt := <-gtun.sndqueue
		gtun.conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
		_, err := gtun.conn.Write(pkt)
		gtun.conn.SetWriteDeadline(time.Time{})
		if err != nil {
			glog.ERROR(err)
			break
		}
	}
}

func Heartbeat(gtun *GtunContext) {
	for {
		select {
		case <-time.After(time.Second * 3):
			bytes, _ := common.Encode(common.C2S_HEARTBEAT, nil)
			gtun.sndqueue <- bytes
		}
	}
}

func IfaceRead(ifce *water.Interface, gtun *GtunContext) {
	packet := make([]byte, 65536)
	for {
		n, err := ifce.Read(packet)
		if err != nil {
			glog.ERROR(err)
			break
		}

		bytes, _ := common.Encode(common.C2C_DATA, packet[:n])
		gtun.sndqueue <- bytes
	}
}

func ConServer(srv string) (conn net.Conn, err error) {
	srvaddr, err := net.ResolveTCPAddr("tcp", srv)
	if err != nil {
		return nil, err
	}

	for {
		tcp, err := net.DialTCP("tcp", nil, srvaddr)
		if err != nil {
			glog.ERROR(err)
			time.Sleep(time.Second * 3)
			continue
		}
		tcp.SetKeepAlive(true)
		tcp.SetKeepAlivePeriod(time.Second * 30)

		return tcp, nil
	}
}

func SetDeviceIP(gtun *GtunContext) (err error) {
	type CMD struct {
		cmd  string
		args []string
	}

	cmdlist := make([]*CMD, 0)

	switch runtime.GOOS {
	case "linux":
		cmdlist = append(cmdlist, &CMD{cmd: "ifconfig", args: []string{gtun.ldev, "up"}})
		args := strings.Split(fmt.Sprintf("addr add %s/24 dev %s", gtun.dhcpip, gtun.ldev), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "ip", args: args})

	case "darwin":
		cmdlist = append(cmdlist, &CMD{cmd: "ifconfig", args: []string{gtun.ldev, "up"}})

		args := strings.Split(fmt.Sprintf("%s %s %s", gtun.ldev, gtun.dhcpip, gtun.dhcpip), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "ifconfig", args: args})

		args = strings.Split(fmt.Sprintf("add -net %s/24 %s", gtun.gateway, gtun.dhcpip), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "route", args: args})

	case "windows":
		args := strings.Split(fmt.Sprintf("interface ip set address name=\"%s\" addr=%s source=static mask=255.255.255.0 gateway=%s", gtun.ldev, gtun.dhcpip, gtun.gateway), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "netsh", args: args})
	}

	for _, c := range cmdlist {
		output, err := exec.Command(c.cmd, c.args...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("run %s error %s", c, string(output))
		}
	}

	return nil
}

func Login() error {
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

	buff, _ := common.Encode(common.C2S_AUTHORIZE, payload)

	_, err = conn.Write(buff)
	if err != nil {
		return nil, err
	}

	cmd, resp, err := common.Decode(conn)
	if err != nil {
		return nil, err
	}

	if cmd != common.S2C_AUTHORIZE {
		err = fmt.Errorf("invalid authorize cmd %d", cmd)
		return nil, err
	}

	s2cauthorize = &common.S2CAuthorize{}
	err = json.Unmarshal(resp, s2cauthorize)
	if err != nil {
		return nil, err
	}

	return s2cauthorize, nil
}

func InsertRoute(gtun *GtunContext) {
	// Windows platform route add need iface index args.
	ifceIndex := -1
	ifce, err := net.InterfaceByName(gtun.ldev)
	if err != nil {
		if runtime.GOOS == "windows" {
			return
		}
	} else {
		ifceIndex = ifce.Index
	}
	for _, address := range gtun.route {
		insertRoute(address, gtun.ldev, gtun.dhcpip, gtun.gateway, ifceIndex)
	}
}

func insertRoute(address, device, tunip, gateway string, ifceIndex int) {
	type CMD struct {
		cmd  string
		args []string
	}

	cmd := &CMD{}

	switch runtime.GOOS {
	case "linux":
		args := strings.Split(fmt.Sprintf("ro add %s dev %s", address, device), " ")
		cmd = &CMD{cmd: "ip", args: args}

	case "darwin":
		args := strings.Split(fmt.Sprintf("add -net %s %s", address, tunip), " ")
		cmd = &CMD{cmd: "route", args: args}

	case "windows":
		args := strings.Split(fmt.Sprintf("add %s %s if %d", address, gateway, ifceIndex), " ")
		cmd = &CMD{cmd: "route", args: args}

	default:
		return
	}

	output, err := exec.Command(cmd.cmd, cmd.args...).CombinedOutput()
	if err != nil {
		glog.DEBUG("add", address, "fail:", string(output))
	}
	glog.DEBUG(cmd, "output", string(output))
}
