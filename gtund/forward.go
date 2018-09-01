package gtund

import (
	"fmt"
	"net"
	"sync"

	"github.com/ICKelin/gtun/common"
)

type ForwardConfig struct {
}
type Forward struct {
	table *sync.Map
}

func NewForward() *Forward {
	forwardTable := &Forward{
		table: new(sync.Map),
	}
	return forwardTable
}

func (forward *Forward) Add(cip string, conn net.Conn) {
	forward.table.Store(cip, conn)
}

func (forward *Forward) Get(cip string) (conn net.Conn) {
	val, ok := forward.table.Load(cip)
	if ok {
		return val.(net.Conn)
	}
	return nil
}

func (forward *Forward) Del(cip string) {
	forward.table.Delete(cip)
}

func (forward *Forward) Broadcast(sndqueue chan *GtunClientContext, buff []byte) {
	forward.table.Range(func(key, val interface{}) bool {
		conn, ok := val.(net.Conn)
		if ok {
			bytes, _ := common.Encode(common.C2C_DATA, buff)
			sndqueue <- &GtunClientContext{conn: conn, payload: bytes}
		}
		return true
	})
}

func (forward *Forward) Peer(sndqueue chan *GtunClientContext, dst string, buff []byte) error {
	c := forward.Get(dst)
	if c == nil {
		return fmt.Errorf("%s offline", dst)
	}

	bytes, err := common.Encode(common.C2C_DATA, buff)
	if err != nil {
		return err
	}
	sndqueue <- &GtunClientContext{conn: c, payload: bytes}

	return nil
}
