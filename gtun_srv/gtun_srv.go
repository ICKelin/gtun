package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/ICKelin/glog"
	"github.com/ICKelin/gtun/common"
	"github.com/songgao/water"
)

type DHCPPool struct {
	sync.Mutex
	ippool map[string]bool
}

func NewDHCPPool() (pool *DHCPPool) {
	pool = &DHCPPool{}
	pool.ippool = make(map[string]bool)
	for i := 10; i < 250; i++ {
		ip := fmt.Sprintf("10.10.253.%d", i)
		pool.ippool[ip] = false
	}
	return pool
}

func (this *DHCPPool) SelectIP() (ip string, err error) {
	this.Lock()
	defer this.Unlock()
	for ip, v := range this.ippool {
		if v == false {
			this.ippool[ip] = true
			return ip, nil
		}
	}
	return "", fmt.Errorf("not enough ip in pool")
}

func (this *DHCPPool) RecycleIP(ip string) {
	this.Lock()
	defer this.Unlock()
	this.ippool[ip] = false
}

func (this *DHCPPool) InUsed(ip string) bool {
	this.Lock()
	defer this.Unlock()
	return this.ippool[ip]
}

type ClientPool struct {
	sync.Mutex
	client map[string]net.Conn
}

func NewClientPool() (clientpool *ClientPool) {
	clientpool = &ClientPool{}
	clientpool.client = make(map[string]net.Conn)
	return clientpool
}

func (this *ClientPool) Add(cip string, conn net.Conn) {
	this.Lock()
	defer this.Unlock()
	this.client[cip] = conn
}

func (this *ClientPool) Get(cip string) (conn net.Conn) {
	this.Lock()
	defer this.Unlock()
	return this.client[cip]
}

func (this *ClientPool) Del(cip string) {
	this.Lock()
	defer this.Unlock()
	delete(this.client, cip)
}

var dhcppool = NewDHCPPool()
var clientpool = NewClientPool()

var (
	pkey   = flag.String("key", "gtun_authorize", "client authorize key")
	pcloud = flag.Bool("cloud", false, "cloud mode")
)

func main() {
	flag.Parse()
	CloudMode("gtun", "10.10.253.1", ":9621")
}

type GtunClientContext struct {
	conn    net.Conn
	payload []byte
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

		if nr < 25 {
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

	s2cauthorize := &common.S2CAuthorize{}
	if auth.Key != *pkey {
		s2cauthorize.AccessIP = ""
		s2cauthorize.Status = "authorize fail"
	} else if accessip == "" {
		accessip, err = dhcppool.SelectIP()
		if err != nil {
			return "", err
		}
		s2cauthorize.AccessIP = accessip
		s2cauthorize.Status = "authorize success"
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
