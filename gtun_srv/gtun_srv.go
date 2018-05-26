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
	"strings"
	"syscall"
	"time"

	"github.com/ICKelin/glog"
	"github.com/ICKelin/gtun/common"
	"github.com/ICKelin/gtun/reverse"
	"github.com/songgao/water"
)

var (
	pkey        = flag.String("k", "gtun_authorize", "client authorize key")
	pgateway    = flag.String("g", "192.168.253.1", "local tun device ip")
	pladdr      = flag.String("l", ":9621", "local listen address")
	proute      = flag.String("r", "", "router rules url")
	pnameserver = flag.String("n", "", "nameservers for gtun_cli")
	preverse    = flag.String("p", "./reverse", "reverse proxy policy path")
	phelp       = flag.Bool("h", false, "print usage")
	ptap        = flag.Bool("t", false, "tap device")

	dhcppool       *DHCPPool
	clientpool     = NewClientPool()
	gRoute         = make([]string, 0)
	gNameserver    = make([]string, 0)
	gReversePolicy = make([]*ReversePolicy, 0)
)

func main() {
	flag.Parse()

	if *proute != "" {
		err := LoadRules(*proute)
		if err != nil {
			glog.WARM("load rules fail: ", err)
		}
	}

	if *pnameserver != "" {
		gNameserver = strings.Split(*pnameserver, ",")
	}

	if *pgateway == "" {
		glog.ERROR("gateway MUST NOT be empty")
		return
	}

	sp := strings.Split(*pgateway, ".")
	if len(sp) != 4 {
		glog.ERROR("ip address format fail", *pgateway)
		return
	}

	prefix := fmt.Sprintf("%s.%s.%s", sp[0], sp[1], sp[2])
	dhcppool = NewDHCPPool(prefix)

	// 2018.05.03
	// Support for reverse proxy
	if *preverse != "" {
		err := LoadReversePolicy(*preverse)
		if err != nil {
			glog.WARM("load reverse policy fail:", err)
		} else {
			for _, r := range gReversePolicy {
				go reverse.Proxy(r.Proto, r.From, r.To)
			}
		}
	}

	GtunServe(*pgateway, *pladdr)
}

type GtunClientContext struct {
	conn    net.Conn
	payload []byte
}

type ReversePolicy struct {
	Proto string `json:"proto"`
	From  string `json:"from"`
	To    string `to:"json"`
}

// Purpose:
//			Loading ip/cidr from file and deploy to gtun_cli
//			It seems like deploy router table to client, tell
//			the client to route these ips/cidrs
//			THERE IS NOT IP VALIDATE
func LoadRules(rfile string) error {
	fp, err := os.Open(rfile)
	if err != nil {
		return err
	}

	linecount := 0
	maxbytes := 0xff00
	curbytes := 0
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
		// There is no plan to fix this "feature"
		if linecount > 20 {
			gRoute = []string{}
			return fmt.Errorf("rules set max record set to 20, suggest using url instead of rule file")
		}

		// 2018.04.20 check max bytes
		// since the protocol header set 2bytes for pkt header
		// once overflow, cli json decode fail
		curbytes += len(bline)
		if curbytes > maxbytes {
			gRoute = []string{}
			return fmt.Errorf("rule set max bytes 0xff00")
		}

		// ignore comment
		if len(line) > 0 && line[0] == '#' {
			continue
		}

		gRoute = append(gRoute, line)
	}

	return nil
}

// 2018.05.03
// Purpose:
//			Loading reverse policy from path
//			The format of policy is from->to (example: :58422->192.168.8.10:8000)
//
// 2018.05.20
//			The format of policy change to: proto from->to to support udp reverse proxy
//			(example: tcp :58422->192.168.8.10:8000)
//			(example: udp :53->192.168.8.10:53)
//
func LoadReversePolicy(path string) error {
	fp, err := os.Open(path)
	if err != nil {
		return err
	}

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
		if len(line) > 0 && line[0] == '#' {
			continue
		}

		sp := strings.Split(line, "->")
		if len(sp) != 3 {
			continue
		}

		reversePolicy := &ReversePolicy{
			Proto: sp[0],
			From:  sp[1],
			To:    sp[2],
		}

		gReversePolicy = append(gReversePolicy, reversePolicy)
	}

	return nil
}

func GtunServe(lip, listenAddr string) {
	cfg := water.Config{}

	if *ptap {
		cfg.DeviceType = water.TAP
	} else {
		cfg.DeviceType = water.TUN
	}

	ifce, err := water.New(cfg)
	if err != nil {
		glog.ERROR(err)
		return
	}

	err = SetDeviceIP(ifce.Name(), lip)
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
	go RespClient(sndqueue)

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			glog.ERROR(err)
			break
		}

		conn.SetKeepAlive(true)
		conn.SetKeepAlivePeriod(time.Second * 30)
		go HandleClient(ifce, conn, sndqueue)
	}
}

func HandleClient(ifce *water.Interface, conn net.Conn, sndqueue chan *GtunClientContext) {
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
			bytes, err := common.Encode(common.S2C_HEARTBEAT, nil)
			if err != nil {
				glog.ERROR(err)
				continue
			}

			sndqueue <- &GtunClientContext{conn: conn, payload: bytes}

		case common.C2C_DATA:
			_, err = ifce.Write(pkt)
			if err != nil {
				glog.ERROR(err)
			}

		default:
			glog.INFO("unimplement cmd", cmd, len(pkt))
		}
	}
}

// Purpose:
//			Read pkt/frame from tun/tap device and send back to gtun_cli
//			For tap device, I do not record MAC address table, STILL USE
//			IP ADDRESS AS SESSION KEY.
// Parameters:
//			ifce => device to read
//			sndqueue => send back queue
//
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

		ethOffset := 0

		if ifce.IsTAP() {
			if nr < 14 {
				glog.WARM("too short ethernet frame", nr)
				continue
			}

			// Not eq ip pkt, just broadcast it
			// This handle maybe dangerous
			if WhichProtocol(buff) != syscall.IPPROTO_IP {
				clientpool.Lock()
				for _, c := range clientpool.client {
					bytes, _ := common.Encode(common.C2C_DATA, buff[:nr])

					sndqueue <- &GtunClientContext{conn: c, payload: bytes}
				}
				clientpool.Unlock()
				continue
			}

			ethOffset = 14
		}

		if ifce.IsTUN() {
			if nr < 20 {
				glog.WARM("too short ippkt", nr)
				continue
			}
		}

		if nr < ethOffset+20 {
			glog.WARM("to short ippkt", nr, ethOffset+20)
			continue
		}

		// TODO ip version
		dst := ""
		if isIPV4(buff[ethOffset]) {
			dst = fmt.Sprintf("%d.%d.%d.%d", buff[ethOffset+16], buff[ethOffset+17], buff[ethOffset+18], buff[ethOffset+19])
		} else {
			glog.WARM("not support ipv6")
		}
		c := clientpool.Get(dst)
		if c != nil {
			bytes, err := common.Encode(common.C2C_DATA, buff[:nr])
			if err != nil {
				glog.ERROR(err)
				continue
			}

			sndqueue <- &GtunClientContext{conn: c, payload: bytes}
		} else {
			glog.ERROR(dst, "offline")
		}
	}
}

func RespClient(sndqueue chan *GtunClientContext) {
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

func SetDeviceIP(dev, tunip string) (err error) {
	type CMD struct {
		cmd  string
		args []string
	}

	cmdlist := make([]*CMD, 0)

	cmdlist = append(cmdlist, &CMD{cmd: "ifconfig", args: []string{dev, "up"}})

	args := strings.Split(fmt.Sprintf("addr add %s/24 dev %s", tunip, dev), " ")
	cmdlist = append(cmdlist, &CMD{cmd: "ip", args: args})

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
		AccessIP:    accessip,
		Status:      "authorize fail",
		RouteRule:   make([]string, 0),
		Nameservers: make([]string, 0),
		Gateway:     *pgateway,
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
		s2cauthorize.Nameservers = gNameserver
	}

	resp, err := json.Marshal(s2cauthorize)
	if err != nil {
		return "", err
	}

	buff, _ := common.Encode(common.S2C_AUTHORIZE, resp)
	_, err = conn.Write(buff)
	if err != nil {
		return "", err
	}

	return accessip, nil
}

func WhichProtocol(frame []byte) int {
	if len(frame) > 14 {
		return int(frame[12])<<8 + int(frame[13])
	}
	return -1
}

func isIPV4(vhl byte) bool {
	if (vhl >> 4) == 4 {
		return true
	}
	return false
}
