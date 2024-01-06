// package main

// import "sync"

// // import (
// // 	"github.com/allegro/bigcache/v3"
// // )

// // type DataBase struct {
// // 	key        string
// // 	genertaion int
// // 	cache      *bigcache.BigCache
// // }

// // func NewDataBase(cache *bigcache.BigCache) DataBase {
// // 	return DataBase{
// // 		cache: cache,
// // 	}
// // }

// // func (d *DataBase) SetValue(key string, value []byte) {
// // 	d.genertaion = d.genertaion + 1
// // 	d.cache.Set(key, value)
// // 	d.key = key
// // }

// // func (d *DataBase) GetValue(key string) ([]byte, error) {
// // 	return d.cache.Get(key)
// // }

// //	func (d *DataBase) notifyValue() bool {
// //		return false
// //	}
// const MembersToNotify = 2

// type oneAndOnlyNumber struct {
// 	num        int
// 	generation int
// 	numMutex   sync.RWMutex
// }

// func InitTheNumber(val int) *oneAndOnlyNumber {
// 	return &oneAndOnlyNumber{
// 		num: val,
// 	}
// }

// func (n *oneAndOnlyNumber) setValue(newVal int) {
// 	n.numMutex.Lock()
// 	defer n.numMutex.Unlock()
// 	n.num = newVal
// 	n.generation = n.generation + 1
// }

// func (n *oneAndOnlyNumber) getValue() (int, int) {
// 	n.numMutex.RLock()
// 	defer n.numMutex.RUnlock()
// 	return n.num, n.generation
// }

// func (n *oneAndOnlyNumber) notifyValue(curVal int, curGeneration int) bool {
// 	if curGeneration > n.generation {
// 		n.numMutex.Lock()
// 		defer n.numMutex.Unlock()
// 		n.generation = curGeneration
// 		n.num = curVal
// 		return true
// 	}
// 	return false
// }
