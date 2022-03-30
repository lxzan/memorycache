package memorycache

import (
	"github.com/lxzan/memorycache/internal/heap"
	"time"
)

type (
	Config struct {
		TTLCheckInterval int    // second
		ClearPerCheck    uint32 // clear keys per check
		Segment          uint32 // bucket segments, segment=2^n
	}

	MemoryCache struct {
		storage *concurrent_hashmap
	}

	element struct {
		Value    interface{}
		ExpireAt int64 // ms, -1 as forever
	}
)

func (self *Config) initialize() *Config {
	if self.Segment <= 0 {
		self.Segment = 16
	}
	if self.TTLCheckInterval == 0 {
		self.TTLCheckInterval = 30
	}
	if self.ClearPerCheck == 0 {
		self.ClearPerCheck = 100
	}
	return self
}

func (self Config) checkSegment() {
	var segment = self.Segment
	var msg = "segment=2^n"
	if segment <= 1 {
		panic(msg)
	}
	for segment > 1 {
		if segment%2 != 0 {
			panic(segment)
		}
		segment /= 2
	}
}

func New(config ...Config) *MemoryCache {
	var cfg Config
	if len(config) > 0 {
		cfg = config[0]
	}
	cfg.initialize().checkSegment()
	return &MemoryCache{
		storage: newConcurrentHashmap(cfg),
	}
}

func (self *MemoryCache) valid(ts int64) bool {
	return ts == -1 || ts > time.Now().UnixMilli()
}

func (self *MemoryCache) getExpire(expire ...time.Duration) int64 {
	if len(expire) == 0 || expire[0] < 0 {
		return -1
	}
	return time.Now().Add(expire[0]).UnixMilli()
}

// empty exp means forever
func (self *MemoryCache) Set(key string, value interface{}, expire ...time.Duration) {
	var ele = element{
		Value:    value,
		ExpireAt: self.getExpire(expire...),
	}

	var bucket = self.storage.getBucket(&key)
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

func (self *MemoryCache) Get(key string) (interface{}, bool) {
	var bucket = self.storage.getBucket(&key)
	bucket.RLock()
	defer bucket.RUnlock()
	result, exist := bucket.data[key]
	if !exist || !self.valid(result.ExpireAt) {
		return nil, false
	}
	return result.Value, true
}

func (self *MemoryCache) Delete(key string) {
	var bucket = self.storage.getBucket(&key)
	bucket.Lock()
	delete(bucket.data, key)
	bucket.Unlock()
}

func (self *MemoryCache) Expire(key string, d time.Duration) {
	var bucket = self.storage.getBucket(&key)
	bucket.Lock()
	if result, exist := bucket.data[key]; exist && self.valid(result.ExpireAt) {
		result.ExpireAt = self.getExpire(d)
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

func (self *MemoryCache) Keys() []string {
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

func (self *MemoryCache) Len() int {
	var num = 0
	for i, _ := range self.storage.buckets {
		var bucket = &self.storage.buckets[i]
		bucket.RLock()
		for _, v := range bucket.data {
			if self.valid(v.ExpireAt) {
				num++
			}
		}
		bucket.RUnlock()
	}
	return num
}
