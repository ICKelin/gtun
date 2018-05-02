package reverse

import (
	"net"
	"sync"

	"github.com/ICKelin/glog"
)

func Proxy(from, to string) {
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

		go reverse(conn, to)
	}
}

func reverse(clientconn net.Conn, to string) {
	defer clientconn.Close()
	rconn, err := net.Dial("tcp", to)
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
