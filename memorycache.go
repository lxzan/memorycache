package memorycache

import (
	"github.com/lxzan/memorycache/internal/heap"
	"github.com/lxzan/memorycache/internal/utils"
	"time"
)

type (
	Config struct {
		TTLCheckInterval int    // second
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

func (c *Config) initialize() *Config {
	if c.Segment <= 0 {
		c.Segment = 16
	}
	if c.TTLCheckInterval == 0 {
		c.TTLCheckInterval = 30
	}
	return c
}

func (c *Config) checkSegment() {
	var segment = c.Segment
	var msg = "segment=2^n"
	for segment > 1 {
		if segment%2 != 0 {
			panic(msg)
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

func (c *MemoryCache) valid(ts int64) bool {
	return ts == -1 || ts > utils.Timestamp()
}

func (c *MemoryCache) getExpireTimestamp(ttl ...time.Duration) int64 {
	if len(ttl) == 0 || ttl[0] <= 0 {
		return -1
	}
	return time.Now().Add(ttl[0]).UnixNano() / 1000000
}

// ttl: -1 means forever
func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	var ele = element{
		Value:    value,
		ExpireAt: c.getExpireTimestamp(ttl),
	}

	var bucket = c.storage.getBucket(key)
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

func (c *MemoryCache) Get(key string) (interface{}, bool) {
	var bucket = c.storage.getBucket(key)
	bucket.RLock()
	defer bucket.RUnlock()
	result, exist := bucket.data[key]
	if !exist || !c.valid(result.ExpireAt) {
		return nil, false
	}
	return result.Value, true
}

func (c *MemoryCache) Delete(key string) {
	var bucket = c.storage.getBucket(key)
	bucket.Lock()
	delete(bucket.data, key)
	bucket.Unlock()
}

func (c *MemoryCache) Expire(key string, ttl time.Duration) {
	var bucket = c.storage.getBucket(key)
	bucket.Lock()
	if result, exist := bucket.data[key]; exist && c.valid(result.ExpireAt) {
		result.ExpireAt = c.getExpireTimestamp(ttl)
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

func (c *MemoryCache) Keys() []string {
	var arr = make([]string, 0)
	for i, _ := range c.storage.buckets {
		var bucket = &c.storage.buckets[i]
		bucket.RLock()
		for k, v := range bucket.data {
			if c.valid(v.ExpireAt) {
				arr = append(arr, k)
			}
		}
		bucket.RUnlock()
	}
	return arr
}

func (c *MemoryCache) Len() int {
	var num = 0
	for i, _ := range c.storage.buckets {
		var bucket = &c.storage.buckets[i]
		bucket.RLock()
		for _, v := range bucket.data {
			if c.valid(v.ExpireAt) {
				num++
			}
		}
		bucket.RUnlock()
	}
	return num
}
