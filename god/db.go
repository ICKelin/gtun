package god

import (
	"time"

	"github.com/boltdb/bolt"
)

var (
	defaultPath  = "god.db"
	defaultBucke = "god"
)

type DBConfig struct {
	Path       string `json:"path"`
	BucketName string `json:"bucket_name"`
}

type DB struct {
	store   *bolt.DB
	buckets map[interface{}]*bolt.Bucket
}

func NewDB(path string) (*DB, error) {
	store, err := bolt.Open(path, 0600, &bolt.Options{Timeout: time.Second * 1})
	if err != nil {
		return nil, err
	}
	db := &DB{
		store: store,
	}

	return db, err
}

func (db *DB) NewBucket(bucket string) (*bolt.Bucket, error) {
	db.store.View(func(txt *bolt.Tx) error {
		db.buckets[bucket] = txt.Bucket([]byte(bucket))
		return nil
	})
	return db.buckets[bucket], nil
}

func (db *DB) Close() {
	db.store.Close()
}

func (db *DB) Get(bucket string, key []byte) {
	db.buckets[bucket].Get([]byte(key))
}

func (db *DB) Put(bucket string, key []byte, value []byte) {
	db.buckets[bucket].Put(key, value)
}

func (db *DB) Remove(bucket string, key []byte) {
	db.buckets[bucket].Delete(key)
}
