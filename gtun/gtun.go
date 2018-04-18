package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
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
	pdev   = flag.String("dev", "gtun", "local tun device name")
	pkey   = flag.String("key", "gtun_authorize", "client authorize key")
	pdebug = flag.Bool("debug", false, "debug mode")
)

type GtunContext struct {
	sync.Mutex
	conn     net.Conn
	srv      string
	key      string
	dhcpip   string
	ldev     string
	sndqueue chan []byte
	rcvqueue chan []byte
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

	this.Lock()
	this.dhcpip = s2c.AccessIP
	this.conn = conn
	this.Unlock()

	return nil
}

func main() {
	flag.Parse()
	if *pdebug {
		glog.Init("gtun", glog.PRIORITY_DEBUG, "./", glog.OPT_DATE, 1024*10)
	} else {
		glog.Init("gtun", glog.PRIORITY_INFO, "./", glog.OPT_DATE, 1024*10)
	}

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

	err = SetTunIP(gtun.ldev, gtun.dhcpip)
	if err != nil {
		glog.ERROR(err)
		return
	}

	go Heartbeat(gtun)
	go IfaceRead(ifce, gtun)

	for {
		wg := &sync.WaitGroup{}
		wg.Add(2)

		go Snd(ifce, gtun, wg)
		go Rcv(ifce, gtun, wg)

		wg.Wait()
		glog.INFO("reconnect")

		err = gtun.ConServer()
		if err != nil && err != io.EOF {
			glog.ERROR(err)
			break
		}

	}

	sig := make(chan os.Signal, 3)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGABRT, syscall.SIGHUP)
	<-sig
}

func Rcv(ifce *water.Interface, gtun *GtunContext, wg *sync.WaitGroup) {
	defer wg.Done()
	defer gtun.conn.Close()

	for {
		cmd, pkt, err := common.Decode(gtun.conn)
		if err != nil {
			glog.INFO(cmd, pkt, err)
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
			bytes := common.Encode(common.C2S_HEARTBEAT, nil)
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
		bytes := common.Encode(common.C2C_DATA, packet[:n])
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

func Authorize(conn net.Conn, accessIP, key string) (s2cauthorize *common.S2CAuthorize, err error) {
	c2sauthorize := &common.C2SAuthorize{
		AccessIP: accessIP,
		Key:      key,
	}

	payload, err := json.Marshal(c2sauthorize)
	if err != nil {
		return nil, err
	}

	buff := common.Encode(common.C2S_AUTHORIZE, payload)

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
