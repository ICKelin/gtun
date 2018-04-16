package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ICKelin/glog"
	"github.com/ICKelin/gtun/common"
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
	if *pcloud {
		// go CloudMode("gtun", "10.10.253.1", ":9621")
	} else {
		go LANMode(":9621")
	}

	sig := make(chan os.Signal, 3)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGABRT, syscall.SIGHUP)
	<-sig
}

func LANMode(addr string) {
	laddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		glog.ERROR(err)
		return
	}

	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		glog.ERROR(err)
		return
	}

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			glog.ERROR(err)
			break
		}

		conn.SetKeepAlive(true)
		conn.SetKeepAlivePeriod(time.Second * 30)
		go HandleClient(conn)
	}
}

func HandleClient(conn net.Conn) {
	defer conn.Close()
	accessip, err := Authorize(conn)
	if err != nil {
		glog.ERROR("authorize fail", err)
		return
	}

	defer glog.INFO("disconnect from", conn.RemoteAddr().String(), "his ip is ", accessip)
	glog.INFO("accept gtun client from", conn.RemoteAddr().String(), "assign ip", accessip)

	defer dhcppool.RecycleIP(accessip)

	clientpool.Add(accessip, conn)
	defer clientpool.Del(accessip)

	for {
		conn.SetReadDeadline(time.Now().Add(time.Minute * 30))
		cmd, pkt, err := common.Decode(conn)
		conn.SetReadDeadline(time.Time{})
		if err != nil {
			if err != io.EOF {
				glog.ERROR(err)
			}
			break
		}

		switch cmd {
		case common.C2S_HEARTBEAT:
			bytes := common.Encode(common.S2C_HEARTBEAT, nil)
			conn.Write(bytes)

		case common.C2C_DATA:
			// TODO: remove
			dst := fmt.Sprintf("%d.%d.%d.%d", pkt[16], pkt[17], pkt[18], pkt[19])

			glog.DEBUG("write to peer", dst)

			c := clientpool.Get(dst)
			if c != nil {
				c.SetWriteDeadline(time.Now().Add(time.Second * 10))
				bytes := common.Encode(common.C2C_DATA, pkt)
				_, err = c.Write(bytes)
				c.SetWriteDeadline(time.Time{})
				if err != nil {
					glog.ERROR("write to peer ", c.RemoteAddr().String(), dst, err)
				}
			} else {
				glog.ERROR(dst, "offline")
			}
		default:
			glog.INFO("unimplement cmd", cmd)

		}
	}
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
	} else {
		if accessip == "" {
			accessip, err = dhcppool.SelectIP()
			if err != nil {
				return "", err
			}
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
