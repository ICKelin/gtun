package registry

import (
	"sync"

	"github.com/ICKelin/gtun/common"
)

type Models struct {
	cache *sync.Map
}

func NewModels() *Models {
	return &Models{
		cache: &sync.Map{},
	}
}

func (m *Models) NewGtund(addr string, reg *common.S2GRegister) {
	m.cache.Store(addr, reg)
}

func (m *Models) RemoveGtund(addr string) {
	m.cache.Delete(addr)
}

func (m *Models) UpdateRefCount(addr string, count int) {
	ele, ok := m.cache.Load(addr)
	if !ok {
		return
	}

	ele.(*common.S2GRegister).Count += count
	m.cache.Store(addr, ele)
}
