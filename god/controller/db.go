package controller

import (
	"fmt"
	"sync"

	"github.com/ICKelin/gtun/common"
)

var (
	defaultPath     = "god.db"
	gDB         *DB = nil
)

type DB struct {
	sync.Mutex
	records *sync.Map
}

func NewDB() *DB {
	return &DB{
		records: new(sync.Map),
	}
}

func (db *DB) Get(key interface{}) (interface{}, bool) {
	return db.records.Load(key)
}

func (db *DB) Set(key, val interface{}) {
	db.records.Store(key, val)
}

func (db *DB) Del(key interface{}) {
	db.records.Delete(key)
}

func GetDB() *DB {
	if gDB == nil {
		gDB = NewDB()
	}
	return gDB
}

func (db *DB) GetAvailableGtund(isWindows bool) (*common.S2GRegister, error) {
	var result *common.S2GRegister = nil

	db.records.Range(func(key, val interface{}) bool {
		regInfo := val.(*common.S2GRegister)
		if regInfo.Count < regInfo.MaxClientCount && regInfo.IsWindows == isWindows {
			result = regInfo
			// TODO: test gtund availablility here
			return false
		}
		return true
	})
	if result == nil {
		return nil, fmt.Errorf("not available gtund to serve you")
	}
	return result, nil
}
