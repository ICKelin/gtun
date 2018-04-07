package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/ICKelin/glog"
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

const (
	AUTHORIZE_FAIL    = 0x00
	AUTHORIZE_SUCCESS = 0x01
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

		glog.INFO("accept gtun client")
		go HandleClient(conn)
	}
}

func HandleClient(conn net.Conn) {
	defer conn.Close()

	err := Authorize(conn)
	if err != nil {
		glog.ERROR(err)
		return
	}

	cip, err := dhcppool.SelectIP()
	if err != nil {
		glog.ERROR(err)
		return
	}
	defer dhcppool.RecycleIP(cip)

	clientpool.Add(cip, conn)
	defer clientpool.Del(cip)

	if err := DHCP(conn, cip); err != nil {
		glog.ERROR("dhcp ip for client fail", err)
		return
	}

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

func Authorize(conn net.Conn) (err error) {
	plen := make([]byte, 4)
	nr, err := conn.Read(plen)
	if err != nil {
		return err
	}

	payloadlength := binary.BigEndian.Uint32(plen)
	authorizekey := make([]byte, payloadlength)
	nr, err = conn.Read(authorizekey)
	if err != nil {
		return err
	}

	if nr != int(payloadlength) {
		return fmt.Errorf("too short authorize key")
	}

	if string(authorizekey) == *pkey {
		conn.Write([]byte{AUTHORIZE_SUCCESS})
	} else {
		conn.Write([]byte{AUTHORIZE_FAIL})
	}

	return nil
}

func DHCP(conn net.Conn, clientip string) (err error) {
	plen := make([]byte, 4)
	binary.BigEndian.PutUint32(plen, uint32(len(clientip)))

	payload := make([]byte, 0)
	payload = append(payload, plen...)
	payload = append(payload, []byte(clientip)...)

	_, err = conn.Write(payload)
	return err
}
