package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"sync"

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
	pkey = flag.String("key", "gtun_authorize", "client authorize key")
)

func main() {
	flag.Parse()

	laddr, err := net.ResolveTCPAddr("tcp", ":9621")
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

		go HandleClient(conn)
	}
}

func HandleClient(conn net.Conn) {
	defer conn.Close()

	accessip, err := Authorize(conn)
	if err != nil {
		glog.ERROR(err)
		return
	}

	glog.INFO("accept gtun client from", conn.RemoteAddr().String(), "assign ip", accessip)
	defer dhcppool.RecycleIP(accessip)

	clientpool.Add(accessip, conn)
	defer clientpool.Del(accessip)

	buff := make([]byte, 65536)
	for {
		nr, err := conn.Read(buff)
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

		dst := fmt.Sprintf("%d.%d.%d.%d", buff[20], buff[21], buff[22], buff[23])

		c := clientpool.Get(dst)
		if c != nil {
			c.Write(buff[:nr])
		} else {
			glog.ERROR(dst, "offline")
		}
	}
}

func Authorize(conn net.Conn) (accessip string, err error) {
	payload, err := common.Decode(conn)
	if err != nil {
		return "", err
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
		if accessip == "" || dhcppool.InUsed(accessip) {
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

	buff := common.Encode(resp)
	_, err = conn.Write(buff)
	if err != nil {
		return "", err
	}

	return accessip, nil
}
