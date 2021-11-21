package forward

import (
	"github.com/ICKelin/gtun/internal/logs"
	"github.com/ICKelin/gtun/transport"
	"io"
	"sync"
)

type Forward struct {
	listener   transport.Listener
	routeTable *RouteTable
	mempool    sync.Pool
}

func NewForward(listener transport.Listener, routeTable *RouteTable) *Forward {
	return &Forward{
		listener:   listener,
		routeTable: routeTable,
		mempool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 1024*4)
			},
		},
	}
}

func (f *Forward) Serve() error {
	for {
		conn, err := f.listener.Accept()
		if err != nil {
			logs.Error("accept local fail: %v", err)
			break
		}

		logs.Debug("accept new connection: %v", conn.RemoteAddr())
		go f.forward(conn)
	}

	return nil
}

func (f *Forward) forward(conn transport.Conn) {
	defer conn.Close()
	entry, err := f.routeTable.Route()
	if err != nil {
		logs.Error("route fail: %v", err)
		return
	}

	defer entry.conn.Close()
	logs.Debug("open a new connection to next hop:%v", entry.conn.RemoteAddr())

	for {
		stream, err := conn.AcceptStream()
		if err != nil {
			logs.Error("accept stream fail: %v", err)
			break
		}

		logs.Debug("accept stream: %v", conn.RemoteAddr())
		dst, err := entry.conn.OpenStream()
		if err != nil {
			logs.Error("open next hop stream fail: %v", err)
			return
		}

		go f.handleStream(dst, stream)
	}
}

func (f *Forward) handleStream(dst, src transport.Stream) {
	defer dst.Close()
	defer src.Close()

	go func() {
		obj := f.mempool.Get()
		defer f.mempool.Put(obj)
		buf := obj.([]byte)
		io.CopyBuffer(dst, src, buf)
	}()

	obj := f.mempool.Get()
	defer f.mempool.Put(obj)
	buf := obj.([]byte)
	io.CopyBuffer(src, dst, buf)
}
