package god

import (
	"sync"
)

var (
	defaultPath = "god.db"
)

type DB struct {
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
	return NewDB()
}
