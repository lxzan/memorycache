package memorycache

import (
	"github.com/lxzan/memorycache/internal/hashmap"
	"github.com/lxzan/memorycache/internal/heap"
	"github.com/lxzan/memorycache/internal/types"
	"github.com/lxzan/memorycache/internal/utils"
	"time"
)

type MemoryCache struct {
	config  *Config
	storage *hashmap.ConcurrentMap
}

func New(options ...Option) *MemoryCache {
	var config = &Config{}
	options = append(options, withInitialize())
	for _, fn := range options {
		fn(config)
	}

	return &MemoryCache{
		config:  config,
		storage: hashmap.NewConcurrentMap(config.Segment, config.TTLCheckInterval),
	}
}

func (c *MemoryCache) valid(now, t int64) bool {
	return t <= 0 || t > now
}

func (c *MemoryCache) getExpireTimestamp(expiration time.Duration) int64 {
	return time.Now().Add(expiration).UnixNano() / 1000000
}

// Set
// 设置键值和过期时间
// expiration: <=0表示永不过期
// set the key value and expiration time
// expiration: <=0 means never expire
func (c *MemoryCache) Set(key string, value interface{}, expiration time.Duration) {
	var ele = types.Element{
		Value:    value,
		ExpireAt: c.getExpireTimestamp(expiration),
	}

	var bucket = c.storage.GetBucket(key)
	bucket.Lock()
	bucket.Map[key] = ele
	if ele.ExpireAt != -1 {
		bucket.Heap.Push(heap.Element{
			Key:      key,
			ExpireAt: ele.ExpireAt,
		})
	}
	bucket.Unlock()
}

// Get
// 根据键获取值
// get value by key
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	var bucket = c.storage.GetBucket(key)
	bucket.RLock()
	result, exist := bucket.Map[key]
	bucket.RUnlock()
	if !exist || !c.valid(utils.Timestamp(), result.ExpireAt) {
		return nil, false
	}
	return result.Value, true
}

// Delete
// 删除一个键
// delete value by key
func (c *MemoryCache) Delete(key string) {
	var bucket = c.storage.GetBucket(key)
	bucket.Lock()
	delete(bucket.Map, key)
	bucket.Unlock()
}

// 设置键的过期时间
func (c *MemoryCache) Expire(key string, expiration time.Duration) {
	var bucket = c.storage.GetBucket(key)
	bucket.Lock()
	if result, exist := bucket.Map[key]; exist && c.valid(utils.Timestamp(), result.ExpireAt) {
		result.ExpireAt = c.getExpireTimestamp(expiration)
		bucket.Map[key] = result
		if result.ExpireAt > 0 {
			bucket.Heap.Push(heap.Element{
				Key:      key,
				ExpireAt: result.ExpireAt,
			})
		}
	}
	bucket.Unlock()
}

// Keys
// 获取所有有效的键
// list valid keys
func (c *MemoryCache) Keys() []string {
	var arr = make([]string, 0)
	var now = utils.Timestamp()
	for _, bucket := range c.storage.Buckets {
		bucket.RLock()
		for k, v := range bucket.Map {
			if c.valid(now, v.ExpireAt) {
				arr = append(arr, k)
			}
		}
		bucket.RUnlock()
	}
	return arr
}

// Len
// 获取有效元素个数
// get the number of valid keys
func (c *MemoryCache) Len() int {
	var num = 0
	var now = utils.Timestamp()
	for _, bucket := range c.storage.Buckets {
		bucket.RLock()
		for _, v := range bucket.Map {
			if c.valid(now, v.ExpireAt) {
				num++
			}
		}
		bucket.RUnlock()
	}
	return num
}
