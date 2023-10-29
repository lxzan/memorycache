package memorycache

import (
	"github.com/lxzan/memorycache/internal/heap"
	"github.com/lxzan/memorycache/internal/types"
	"hash/maphash"
	"math"
	"strings"
	"sync"
	"time"
)

type MemoryCache struct {
	config  *types.Config
	storage []*bucket
	seed    maphash.Seed
}

// New 创建缓存数据库实例
// Creating a Cached Database Instance
func New(options ...Option) *MemoryCache {
	var config = &types.Config{}
	options = append(options, withInitialize())
	for _, fn := range options {
		fn(config)
	}

	mc := &MemoryCache{
		config:  config,
		storage: make([]*bucket, config.BucketNum),
		seed:    maphash.MakeSeed(),
	}

	for i, _ := range mc.storage {
		mc.storage[i] = &bucket{
			Map:  make(map[string]*types.Element, config.InitialSize),
			Heap: heap.New(config.InitialSize),
		}
	}

	go func() {
		var ticker = time.NewTicker(config.Interval)
		defer ticker.Stop()
		for {
			<-ticker.C

			var now = time.Now().UnixMilli()
			for _, b := range mc.storage {
				b.expireTimeCheck(now, config.MaxKeysDeleted)
			}
		}
	}()

	return mc
}

func (c *MemoryCache) getBucket(key string) *bucket {
	var idx = maphash.String(c.seed, key) & uint64(c.config.BucketNum-1)
	return c.storage[idx]
}

// 获取过期时间, d<=0表示永不过期
func (c *MemoryCache) getExp(d time.Duration) int64 {
	if d <= 0 {
		return math.MaxInt
	}
	return time.Now().Add(d).UnixMilli()
}

// Clear 清空所有缓存
// clear all caches
func (c *MemoryCache) Clear() {
	for _, b := range c.storage {
		b.Lock()
		b.Heap = heap.New(c.config.InitialSize)
		b.Map = make(map[string]*types.Element, c.config.InitialSize)
		b.Unlock()
	}
}

// Set 设置键值和过期时间. exp<=0表示永不过期.
// Set the key value and expiration time. exp<=0 means never expire.
func (c *MemoryCache) Set(key string, value any, exp time.Duration) (replaced bool) {
	var b = c.getBucket(key)
	b.Lock()
	defer b.Unlock()

	var expireAt = c.getExp(exp)
	v, ok := b.Map[key]
	if ok {
		v.Value = value
		v.ExpireAt = expireAt
		b.Heap.Down(v.Index, b.Heap.Len())
		return true
	}

	var ele = &types.Element{Key: key, Value: value, ExpireAt: expireAt}
	b.Heap.Push(ele)
	b.Map[key] = ele
	if b.Heap.Len() > c.config.MaxCapacity {
		delete(b.Map, b.Heap.Pop().Key)
	}
	return false
}

// Get
func (c *MemoryCache) Get(key string) (any, bool) {
	var b = c.getBucket(key)
	b.Lock()
	v, exist := b.Map[key]
	b.Unlock()
	if !exist || v.Expired(time.Now().UnixMilli()) {
		return nil, false
	}
	return v.Value, true
}

// GetAndRefresh 获取. 如果存在, 刷新过期时间.
// Get a value. If it exists, refreshes the expiration time.
func (c *MemoryCache) GetAndRefresh(key string, exp time.Duration) (any, bool) {
	var b = c.getBucket(key)
	b.Lock()
	defer b.Unlock()

	v, exist := b.Map[key]
	if !exist || v.Expired(time.Now().UnixMilli()) {
		return nil, false
	}

	v.ExpireAt = c.getExp(exp)
	b.Heap.Down(v.Index, b.Heap.Len())
	return v, true
}

// Delete
func (c *MemoryCache) Delete(key string) (deleted bool) {
	var b = c.getBucket(key)
	b.Lock()
	defer b.Unlock()

	v, ok := b.Map[key]
	if !ok {
		return false
	}

	b.Heap.Delete(v.Index)
	delete(b.Map, key)
	return true
}

// Keys 获取前缀匹配的key
// Get prefix matching key
func (c *MemoryCache) Keys(prefix string) []string {
	var arr = make([]string, 0)
	var now = time.Now().UnixMilli()
	for _, b := range c.storage {
		b.Lock()
		for _, v := range b.Heap.Data {
			if !v.Expired(now) && strings.HasPrefix(v.Key, prefix) {
				arr = append(arr, v.Key)
			}
		}
		b.Unlock()
	}
	return arr
}

// Len 获取当前元素数量
// Get the number of elements
func (c *MemoryCache) Len() int {
	var num = 0
	for _, b := range c.storage {
		b.Lock()
		num += b.Heap.Len()
		b.Unlock()
	}
	return num
}

type bucket struct {
	sync.Mutex
	Map  map[string]*types.Element
	Heap *heap.Heap
}

// 过期时间检查
func (c *bucket) expireTimeCheck(now int64, num int) {
	c.Lock()
	defer c.Unlock()

	for c.Heap.Len() > 0 && c.Heap.Front().Expired(now) && num > 0 {
		delete(c.Map, c.Heap.Pop().Key)
		num--
	}
}
