package memdb

import (
	"memdb/internal/heap"
	"time"
)

type (
	MemDB struct {
		storage *concurrent_hashmap
	}

	element struct {
		Value    interface{}
		ExpireAt int64 // ms, -1 as forever
	}
)

func New() *MemDB {
	return &MemDB{
		storage: newConcurrentHashmap(16),
	}
}

func (self *MemDB) valid(ts int64) bool {
	return ts == -1 || ts > time.Now().UnixMilli()
}

func (self *MemDB) getExp(exp ...time.Duration) int64 {
	if len(exp) == 0 || exp[0] < 0 {
		return -1
	}
	return time.Now().Add(exp[0]).UnixMilli()
}

func (self *MemDB) Set(key string, value interface{}, exp ...time.Duration) {
	var ele = element{
		Value:    value,
		ExpireAt: self.getExp(exp...),
	}

	var bucket = self.storage.getBucket(key)
	bucket.Lock()
	bucket.data[key] = ele
	if ele.ExpireAt != -1 {
		bucket.ttl.Push(heap.Element{
			Key:      key,
			ExpireAt: ele.ExpireAt,
		})
	}
	bucket.Unlock()
}

func (self *MemDB) Get(key string) (interface{}, bool) {
	var bucket = self.storage.getBucket(key)
	bucket.RLock()
	defer bucket.RUnlock()
	result, exist := bucket.data[key]
	if !exist || !self.valid(result.ExpireAt) {
		return nil, false
	}
	return result.Value, true
}

func (self *MemDB) Delete(key string) {
	var bucket = self.storage.getBucket(key)
	bucket.Lock()
	delete(bucket.data, key)
	bucket.Unlock()
}

func (self *MemDB) Expire(key string, d time.Duration) {
	var bucket = self.storage.getBucket(key)
	bucket.Lock()
	if result, exist := bucket.data[key]; exist && self.valid(result.ExpireAt) {
		result.ExpireAt = self.getExp(d)
		bucket.data[key] = result
		if result.ExpireAt != -1 {
			bucket.ttl.Push(heap.Element{
				Key:      key,
				ExpireAt: result.ExpireAt,
			})
		}
	}
	bucket.Unlock()
}

func (self *MemDB) Keys() []string {
	var arr = make([]string, 0)
	for i, _ := range self.storage.buckets {
		var bucket = &self.storage.buckets[i]
		bucket.RLock()
		for k, v := range bucket.data {
			if self.valid(v.ExpireAt) {
				arr = append(arr, k)
			}
		}
		bucket.RUnlock()
	}
	return arr
}
