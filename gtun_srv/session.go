package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
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

type UserInfo struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Keycode   string `json:"keycode"`
	ExpiredIn int64  `json:"expired_in"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

type UserPool struct {
	sync.Mutex
	Path  string               `json:"-"`
	Users map[string]*UserInfo `json:"users"`
}

func NewUserPool(path string) *UserPool {
	p := &UserPool{
		Path:  path,
		Users: make(map[string]*UserInfo),
	}
	return p
}

func (this *UserPool) Add(user string) {
	this.Lock()
	defer this.Unlock()
	u := &UserInfo{
		Keycode: user,
	}
	this.Users[user] = u
}

func (this *UserPool) Get(user string) *UserInfo {
	this.Lock()
	defer this.Unlock()
	return this.Users[user]
}

func (this *UserPool) Del(user string) string {
	this.Lock()
	defer this.Unlock()
	delete(this.Users, user)
	this.Save()
}

func (this *UserPool) Save() error {
	fp, err := os.Open(this.Path)
	if err != nil {
		return err
	}
	defer fp.Close()

	bytes, err := json.MarshalIndent(this, "", "\t")
	if err != nil {
		return err
	}

	_, err = fp.Write(bytes)
	return err
}

func (this *UserPool) List() (string, error) {
	this.Lock()
	defer this.Unlock()
	bytes, err := json.Marshal(this)
	return string(bytes), err
}
