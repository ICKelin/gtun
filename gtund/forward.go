package gtund

import (
	"net"
	"sync"
)

type ForwardConfig struct {
}
type Forward struct {
	sync.Mutex
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
