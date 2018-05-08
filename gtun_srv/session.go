package main

import (
	"fmt"
	"net"
	"sync"
)

type DHCPPool struct {
	sync.Mutex
	ippool map[string]bool
}

func NewDHCPPool(prefix string) (pool *DHCPPool) {
	pool = &DHCPPool{}
	pool.ippool = make(map[string]bool)
	for i := 10; i < 250; i++ {
		// Force C class ip address
		ip := fmt.Sprintf("%s.%d", prefix, i)
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
