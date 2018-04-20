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
				This program is a gtun server for game/ip accelator.

	Author: ICKelin
*/

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/ICKelin/glog"
	"github.com/ICKelin/gtun/common"
	"github.com/songgao/water"
)

var (
	pkey     = flag.String("k", "gtun_authorize", "client authorize key")
	pgateway = flag.String("g", "192.168.253.1", "local tun device ip")
	pladdr   = flag.String("l", ":9621", "local listen address")
	proute   = flag.String("r", "", "router rules url")
	phelp    = flag.Bool("h", false, "print usage")

	dhcppool   = NewDHCPPool()
	clientpool = NewClientPool()
	gRoute     = make([]string, 0)
)

type GtunClientContext struct {
	conn    net.Conn
	payload []byte
}

func main() {
	flag.Parse()
	if *phelp {
		ShowUsage()
		return
	}

	if *proute != "" {
		err := LoadRules(*proute)
		if err != nil {
			glog.WARM("load rules fail: ", err)
		}

		glog.DEBUG("loaded", *proute, len(gRoute))
	}

	CloudMode("gtun", *pgateway, ":9621")
}

func ShowUsage() {
	flag.Usage()
}

func LoadRules(rfile string) error {
	fp, err := os.Open(rfile)
	if err != nil {
		return err
	}

	linecount := 0
	reader := bufio.NewReader(fp)
	for {
		bline, _, err := reader.ReadLine()
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}

		line := string(bline)
		linecount += 1

		// 2018.04.20 rule store max 20 rule record
		// There is no plan to fix this "feature" for CN
		if linecount > 20 {
			gRoute = []string{}
			return fmt.Errorf("rules set max record set to 20, suggest using url instead of rule file")
		}

		// ignore comment
		if len(line) > 0 && line[0] == '#' {
			continue
		}

		gRoute = append(gRoute, line)
	}

	return nil
}

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

	err = SetTunIP(ifce.Name(), lip)
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

	sndqueue := make(chan *GtunClientContext)
	go IfceRead(ifce, sndqueue)
	go RespGtunCloudClient(sndqueue)

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			glog.ERROR(err)
			break
		}

		conn.SetKeepAlive(true)
		conn.SetKeepAlivePeriod(time.Second * 30)
		go HandleCloudClient(ifce, conn, sndqueue)
	}
}

func HandleCloudClient(ifce *water.Interface, conn net.Conn, sndqueue chan *GtunClientContext) {
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

	for {
		cmd, pkt, err := common.Decode(conn)
		if err != nil {
			if err != io.EOF {
				glog.ERROR(err)
			}
			break
		}

		switch cmd {
		case common.C2S_HEARTBEAT:
			bytes := common.Encode(common.S2C_HEARTBEAT, nil)
			sndqueue <- &GtunClientContext{conn: conn, payload: bytes}

		case common.C2C_DATA:
			if len(pkt) < 20 {
				glog.ERROR("invalid ippkt length", len(pkt))
				break
			}

			_, err = ifce.Write(pkt)
			if err != nil {
				glog.ERROR(err)
			}

		default:
			glog.INFO("unimplement cmd", cmd, len(pkt))
		}
	}
}

func IfceRead(ifce *water.Interface, sndqueue chan *GtunClientContext) {
	buff := make([]byte, 65536)
	for {
		nr, err := ifce.Read(buff)
		if err != nil {
			if err != io.EOF {
				glog.ERROR(err)
			}
			continue
		}

		if nr < 20 {
			glog.ERROR("too short ippkt")
			continue
		}

		dst := fmt.Sprintf("%d.%d.%d.%d", buff[16], buff[17], buff[18], buff[19])
		c := clientpool.Get(dst)
		if c != nil {
			bytes := common.Encode(common.C2C_DATA, buff[:nr])
			sndqueue <- &GtunClientContext{conn: c, payload: bytes}
		} else {
			glog.ERROR(dst, "offline")
		}
	}
}

func RespGtunCloudClient(sndqueue chan *GtunClientContext) {
	for {
		ctx := <-sndqueue
		ctx.conn.SetWriteDeadline(time.Now().Add(time.Second * 30))
		nw, err := ctx.conn.Write(ctx.payload)
		ctx.conn.SetWriteDeadline(time.Time{})
		if err != nil {
			glog.ERROR(err)
		}

		if nw != len(ctx.payload) {
			glog.ERROR("write not full", nw, len(ctx.payload))
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

func Authorize(conn net.Conn) (accessip string, err error) {
	cmd, payload, err := common.Decode(conn)
	if err != nil {
		return "", err
	}

	if cmd != common.C2S_AUTHORIZE {
		return "", fmt.Errorf("invalid authhorize cmd %d", cmd)
	}

	auth := &common.C2SAuthorize{}
	err = json.Unmarshal(payload, &auth)
	if err != nil {
		return "", err
	}

	accessip = auth.AccessIP

	s2cauthorize := &common.S2CAuthorize{
		AccessIP:  accessip,
		Status:    "authorize fail",
		RouteRule: make([]string, 0),
	}

	if auth.Key == *pkey {
		s2cauthorize.Status = "authorize success"
		if accessip == "" {
			accessip, err = dhcppool.SelectIP()
			if err != nil {
				return "", err
			}
			s2cauthorize.AccessIP = accessip
		}
		s2cauthorize.RouteRule = gRoute
	}

	resp, err := json.Marshal(s2cauthorize)
	if err != nil {
		return "", err
	}

	buff := common.Encode(common.S2C_AUTHORIZE, resp)
	_, err = conn.Write(buff)
	if err != nil {
		return "", err
	}

	return accessip, nil
}
