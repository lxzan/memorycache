package memorycache

import (
	"github.com/lxzan/memorycache/internal/hashmap"
	"github.com/lxzan/memorycache/internal/heap"
	"github.com/lxzan/memorycache/internal/types"
	"github.com/lxzan/memorycache/internal/utils"
	"time"
)

type (
	Config struct {
		TTLCheckInterval time.Duration
		Segment          uint32 // bucket segments, segment=2^n
	}

	MemoryCache struct {
		storage *hashmap.ConcurrentMap
	}
)

func (c *Config) initialize() *Config {
	if c.Segment <= 0 {
		c.Segment = 16
	}
	if c.TTLCheckInterval <= 0 {
		c.TTLCheckInterval = 30 * time.Second
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
		storage: hashmap.NewConcurrentMap(cfg.Segment, cfg.TTLCheckInterval),
	}
}

func (c *MemoryCache) valid(ts int64) bool {
	return ts <= 0 || ts > utils.Timestamp()
}

func (c *MemoryCache) getExpireTimestamp(expiration time.Duration) int64 {
	return time.Now().Add(expiration).UnixNano() / 1000000
}

// expiration: <=0表示永不过期
func (c *MemoryCache) Set(key string, value interface{}, expiration time.Duration) {
	var ele = types.Element{
		Value:    value,
		ExpireAt: c.getExpireTimestamp(expiration),
	}

	var bucket = c.storage.GetBucket(key)
	bucket.Lock()
	bucket.M[key] = ele
	if ele.ExpireAt != -1 {
		bucket.H.Push(heap.Element{
			Key:      key,
			ExpireAt: ele.ExpireAt,
		})
	}
	bucket.Unlock()
}

func (c *MemoryCache) Get(key string) (interface{}, bool) {
	var bucket = c.storage.GetBucket(key)
	bucket.RLock()
	defer bucket.RUnlock()
	result, exist := bucket.M[key]
	if !exist || !c.valid(result.ExpireAt) {
		return nil, false
	}
	return result.Value, true
}

func (c *MemoryCache) Delete(key string) {
	var bucket = c.storage.GetBucket(key)
	bucket.Lock()
	delete(bucket.M, key)
	bucket.Unlock()
}

func (c *MemoryCache) Expire(key string, expiration time.Duration) {
	var bucket = c.storage.GetBucket(key)
	bucket.Lock()
	if result, exist := bucket.M[key]; exist && c.valid(result.ExpireAt) {
		result.ExpireAt = c.getExpireTimestamp(expiration)
		bucket.M[key] = result
		if result.ExpireAt > 0 {
			bucket.H.Push(heap.Element{
				Key:      key,
				ExpireAt: result.ExpireAt,
			})
		}
	}
	bucket.Unlock()
}

func (c *MemoryCache) Keys() []string {
	var arr = make([]string, 0)
	for _, bucket := range c.storage.Buckets {
		bucket.RLock()
		for k, v := range bucket.M {
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
	for _, bucket := range c.storage.Buckets {
		bucket.RLock()
		for _, v := range bucket.M {
			if c.valid(v.ExpireAt) {
				num++
			}
		}
		bucket.RUnlock()
	}
	return num
}
