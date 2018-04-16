package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/ICKelin/glog"
	"github.com/ICKelin/gtun/common"
)

func Test_Authorize(t *testing.T) {
	go runserver()
	conn, err := ConServer("127.0.0.1:9621")
	if err != nil {
		t.Error(err)
		return
	}

	s2cauthorize, err := Authorize(conn, "", "gtun_authorize")
	if err != nil && err != io.EOF {
		t.Error(err)
		return
	}

	fmt.Println(s2cauthorize)
}

func runserver() {
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

		conn.SetKeepAlive(true)
		conn.SetKeepAlivePeriod(time.Second * 30)
		go handleclient(conn)
	}
}

func handleclient(conn net.Conn) {
	defer conn.Close()

	err := checkoauth(conn, "gtun_authorize")
	if err != nil && err != io.EOF {
		fmt.Println(err)
		return
	}

}

func checkoauth(conn net.Conn, key string) error {
	cmd, info, err := common.Decode(conn)
	if err != nil {
		return err
	}

	if cmd != common.C2S_AUTHORIZE {
		return fmt.Errorf("invalid authorize cmd")
	}

	c2sauthorize := &common.C2SAuthorize{}
	err = json.Unmarshal(info, &c2sauthorize)
	if err != nil {
		return err
	}

	s2cauthorize := &common.S2CAuthorize{}
	s2cauthorize.AccessIP = "10.10.253.2"
	s2cauthorize.Status = "authorize success"

	response, err := json.Marshal(s2cauthorize)
	if err != nil {
		return err
	}

	buff := common.Encode(common.S2C_AUTHORIZE, response)

	conn.Write(buff)
	return nil
}
