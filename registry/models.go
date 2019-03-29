package registry

import (
	"encoding/json"
	"fmt"
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

func (m *Models) RandomGetGtund(win bool) (*common.S2GRegister, error) {
	var res *common.S2GRegister
	m.cache.Range(func(key, val interface{}) bool {
		gtund, ok := val.(*common.S2GRegister)
		if ok {
			if gtund.IsWindows == win && gtund.Count < gtund.MaxClientCount {
				gtund.Count += 1
				res = gtund
				m.cache.Store(key, res)
				return false
			}
		}
		return true
	})

	if res != nil {
		return res, nil
	}

	return nil, fmt.Errorf("not available gtund node")
}

func (m *Models) Status() string {
	var res = make(map[string]interface{})
	m.cache.Range(func(key, value interface{}) bool {
		res[key.(string)] = value
		return true
	})

	bytes, _ := json.Marshal(res)
	return string(bytes)
}
