package gtund

import (
	"fmt"
	"sync"

	"github.com/ICKelin/gtun/common"
	"github.com/ICKelin/gtun/pkg/logs"
)

type Forward struct {
	table *sync.Map
}

func NewForward() *Forward {
	forwardTable := &Forward{
		table: new(sync.Map),
	}
	return forwardTable
}

func (forward *Forward) Add(cip string, sndbuf chan []byte) {
	forward.table.Store(cip, sndbuf)
}

func (forward *Forward) Get(cip string) chan []byte {
	val, ok := forward.table.Load(cip)
	if ok {
		return val.(chan []byte)
	}
	return nil
}

func (forward *Forward) Del(cip string) {
	forward.table.Delete(cip)
}

func (forward *Forward) Broadcast(buff []byte) {
	forward.table.Range(func(key, val interface{}) bool {
		sndbuf, ok := val.(chan []byte)
		if ok {
			bytes, _ := common.Encode(common.C2C_DATA, buff)
			sndbuf <- bytes
		}
		return true
	})
}

func (forward *Forward) Peer(dst string, buff []byte) error {
	sndbuf := forward.Get(dst)
	if sndbuf == nil {
		return fmt.Errorf("%s offline", dst)
	}

	bytes, err := common.Encode(common.C2C_DATA, buff)
	if err != nil {
		return err
	}

	select {
	case sndbuf <- bytes:
	default:
	}
	logs.Debug("send dst %s bytes size: %d", dst, len(bytes))
	return nil
}
