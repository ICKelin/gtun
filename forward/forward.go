package forward

import (
	"github.com/ICKelin/gtun/internal/logs"
	"io"
	"net"
	"sync"
)

type Forward struct {
	addr       string
	routeTable *RouteTable
	mempool    sync.Pool
}

func NewForward(addr string, routeTable *RouteTable) *Forward {
	return &Forward{
		addr:       addr,
		routeTable: routeTable,
		mempool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 1024*4)
			},
		},
	}
}

func (f *Forward) ListenAndServe() error {
	listener, err := net.Listen("tcp", f.addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			logs.Debug("accept fail: %v", err)
			break
		}

		logs.Debug("accept new connection: %v", conn.RemoteAddr())
		go f.forward(conn)
	}

	return nil
}

func (f *Forward) forward(conn net.Conn) {
	defer conn.Close()
	entry, err := f.routeTable.Route()
	if err != nil {
		logs.Error("route fail: %v", err)
		return
	}
	logs.Debug("prev hop %v next hop:%v", conn.RemoteAddr(), entry.conn.RemoteAddr())

	stream, err := entry.conn.OpenStream()
	if err != nil {
		logs.Error("open next hop stream fail: %v", err)
		return
	}
	defer stream.Close()

	go func() {
		obj := f.mempool.Get()
		defer f.mempool.Put(obj)
		buf := obj.([]byte)
		io.CopyBuffer(conn, stream, buf)
	}()

	obj := f.mempool.Get()
	defer f.mempool.Put(obj)
	buf := obj.([]byte)
	io.CopyBuffer(stream, conn, buf)
}
