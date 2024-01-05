package main

import (
	"github.com/allegro/bigcache/v3"
)

type DataBase struct {
	key        string
	genertaion int
	cache      *bigcache.BigCache
}

func NewDataBase(cache *bigcache.BigCache) DataBase {
	return DataBase{
		cache: cache,
	}
}

func (d *DataBase) SetValue(key string, value []byte) {
	d.genertaion = d.genertaion + 1
	d.cache.Set(key, value)
	d.key = key
}

func (d *DataBase) GetValue(key string) ([]byte, error) {
	return d.cache.Get(key)
}

func (d *DataBase) notifyValue() bool {
	return false
}
