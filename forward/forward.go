package forward

import (
	"github.com/ICKelin/gtun/internal/logs"
	"github.com/ICKelin/gtun/transport"
	"io"
	"sync"
)

type Forward struct {
	listener transport.Listener
	dialer   transport.Dialer
	mempool  sync.Pool
}

func NewForward(listener transport.Listener, dialer transport.Dialer) *Forward {
	return &Forward{
		listener: listener,
		dialer:   dialer,
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

		go f.forward(conn)
	}

	return nil
}

func (f *Forward) forward(conn transport.Conn) {
	defer conn.Close()

	// create a new connection to nexthop
	nexthopConn, err := f.dialer.Dial()
	if err != nil {
		logs.Error("dian next hop fail: %v", err)
		return
	}
	defer nexthopConn.Close()
	logs.Debug("open a new connection to nexthop")

	// TODO: verify connection
	for {
		stream, err := conn.AcceptStream()
		if err != nil {
			logs.Error("accept stream fail: %v", err)
			break
		}

		logs.Debug("accept stream: %v", conn.RemoteAddr())
		dst, err := nexthopConn.OpenStream()
		if err != nil {
			logs.Error("open nexthop stream fail: %v", err)
			return
		}

		go f.handleStream(dst, stream)
	}
}

func (f *Forward) handleStream(dst, src transport.Stream) {
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
