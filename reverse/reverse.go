package reverse

import (
	"net"
	"strings"
	"sync"

	"github.com/ICKelin/glog"
)

func Proxy(prot string, from, to string) {
	if strings.ToLower(prot) == "tcp" {
		ProxyTCP(from, to)
	}

	if strings.ToLower(prot) == "udp" {
		ProxyUDP(from, to)
	}
}

func ProxyTCP(from, to string) {
	listener, err := net.Listen("tcp", from)
	if err != nil {
		glog.ERROR(err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			glog.ERROR(err)
			break
		}

		go reverse("tcp", conn, to)
	}
}

func ProxyUDP(from, to string) {
	laddr, err := net.ResolveUDPAddr("udp", from)
	if err != nil {
		glog.ERROR(err)
		return
	}

	lconn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		glog.ERROR(err)
		return
	}

	reverse("udp", lconn, to)
}

func reverse(proto string, clientconn net.Conn, to string) {
	defer clientconn.Close()
	rconn, err := net.Dial(proto, to)
	if err != nil {
		glog.ERROR(err)
		return
	}
	defer rconn.Close()

	wg := &sync.WaitGroup{}
	wg.Add(2)

	// client->rconn
	go func() {
		defer wg.Done()
		Copy(rconn, clientconn)
	}()

	// rconn->client
	go func() {
		defer wg.Done()
		Copy(clientconn, rconn)
	}()

	wg.Wait()
}

func Copy(dst, src net.Conn) {
	buffer := make([]byte, 2048)
	for {
		nr, err := src.Read(buffer)
		if err != nil {
			glog.ERROR(err)
			break
		}

		_, err = dst.Write(buffer[:nr])
		if err != nil {
			glog.ERROR(err)
			break
		}
	}

}
