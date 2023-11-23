package memorycache

import (
	"context"
	"hash/maphash"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lxzan/memorycache/internal/utils"
)

type MemoryCache struct {
	config    *config
	storage   []*bucket
	seed      maphash.Seed
	timestamp atomic.Int64
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	once      sync.Once
}

// New 创建缓存数据库实例
// Creating a Cached Database Instance
func New(options ...Option) *MemoryCache {
	var c = &config{}
	options = append(options, withInitialize())
	for _, fn := range options {
		fn(c)
	}

	mc := &MemoryCache{
		config:  c,
		storage: make([]*bucket, c.BucketNum),
		seed:    maphash.MakeSeed(),
		wg:      sync.WaitGroup{},
		once:    sync.Once{},
	}
	mc.wg.Add(2)
	mc.ctx, mc.cancel = context.WithCancel(context.Background())
	mc.timestamp.Store(time.Now().UnixMilli())

	for i, _ := range mc.storage {
		mc.storage[i] = &bucket{
			Map:  make(map[string]*Element, c.InitialSize),
			Heap: newHeap(c.InitialSize),
		}
	}

	go func() {
		var d0 = c.MaxInterval
		var ticker = time.NewTicker(d0)
		defer ticker.Stop()

		for {
			select {
			case <-mc.ctx.Done():
				mc.wg.Done()
				return
			case now := <-ticker.C:
				var sum = 0
				for _, b := range mc.storage {
					sum += b.ExpireCheck(now.UnixMilli(), c.MaxKeysDeleted)
				}

				// 删除数量超过阈值, 缩小时间间隔
				if d1 := utils.SelectValue(sum > c.BucketNum*c.MaxKeysDeleted*7/10, c.MinInterval, c.MaxInterval); d1 != d0 {
					d0 = d1
					ticker.Reset(d0)
				}
			}
		}
	}()

	// 每秒更新一次时间戳
	go func() {
		var ticker = time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-mc.ctx.Done():
				mc.wg.Done()
				return
			case now := <-ticker.C:
				mc.timestamp.Store(now.UnixMilli())
			}
		}
	}()

	return mc
}

func (c *MemoryCache) Stop() {
	c.once.Do(func() {
		c.cancel()
		c.wg.Wait()
	})
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
	return c.timestamp.Load() + d.Milliseconds()
}

// 查找数据. 如果存在且超时, 删除并返回false
func (c *MemoryCache) fetch(b *bucket, key string) (*Element, bool) {
	v, exist := b.Map[key]
	if !exist {
		return nil, false
	}

	if v.expired(c.timestamp.Load()) {
		b.Heap.Delete(v.index)
		delete(b.Map, key)
		v.cb(v, ReasonExpired)
		return nil, false
	}

	return v, true
}

// 检查容量溢出
func (c *MemoryCache) overflow(b *bucket) {
	if b.Heap.Len() > c.config.MaxCapacity {
		head := b.Heap.Pop()
		delete(b.Map, head.Key)
		head.cb(head, ReasonOverflow)
	}
}

// Clear 清空所有缓存
// clear all caches
func (c *MemoryCache) Clear() {
	for _, b := range c.storage {
		b.Lock()
		b.Heap = newHeap(c.config.InitialSize)
		b.Map = make(map[string]*Element, c.config.InitialSize)
		b.Unlock()
	}
}

// Set 设置键值和过期时间. exp<=0表示永不过期.
// Set the key value and expiration time. exp<=0 means never expire.
func (c *MemoryCache) Set(key string, value any, exp time.Duration) (replaced bool) {
	return c.SetWithCallback(key, value, exp, emptyCallback)
}

// SetWithCallback 设置键值, 过期时间和回调函数. 容量溢出和过期都会触发回调.
// Set the key value, expiration time and callback function. The callback is triggered by both capacity overflow and expiration.
func (c *MemoryCache) SetWithCallback(key string, value any, exp time.Duration, cb CallbackFunc) (replaced bool) {
	var b = c.getBucket(key)
	b.Lock()
	defer b.Unlock()

	var expireAt = c.getExp(exp)
	v, ok := c.fetch(b, key)
	if ok {
		v.Value = value
		v.cb = cb
		b.Heap.UpdateTTL(v, expireAt)
		return true
	}

	var ele = &Element{Key: key, Value: value, ExpireAt: expireAt, cb: cb}
	b.Heap.Push(ele)
	b.Map[key] = ele
	c.overflow(b)
	return false
}

// Get
func (c *MemoryCache) Get(key string) (v any, exist bool) {
	var b = c.getBucket(key)
	b.Lock()
	defer b.Unlock()
	result, ok := c.fetch(c.getBucket(key), key)
	if !ok {
		return nil, false
	}
	return result.Value, true
}

// GetWithTTL 获取. 如果存在, 刷新过期时间.
// Get a value. If it exists, refreshes the expiration time.
func (c *MemoryCache) GetWithTTL(key string, exp time.Duration) (v any, exist bool) {
	var b = c.getBucket(key)
	b.Lock()
	defer b.Unlock()

	result, ok := c.fetch(b, key)
	if !ok {
		return nil, false
	}

	b.Heap.UpdateTTL(result, c.getExp(exp))
	return result.Value, true
}

// GetOrCreate 如果存在, 刷新过期时间. 如果不存在, 创建一个新的.
// Get or create a value. If it exists, refreshes the expiration time. If it does not exist, creates a new one.
func (c *MemoryCache) GetOrCreate(key string, value any, exp time.Duration) (v any, exist bool) {
	return c.GetOrCreateWithCallback(key, value, exp, emptyCallback)
}

// GetOrCreateWithCallback 如果存在, 刷新过期时间. 如果不存在, 创建一个新的.
// Get or create a value with CallbackFunc. If it exists, refreshes the expiration time. If it does not exist, creates a new one.
func (c *MemoryCache) GetOrCreateWithCallback(key string, value any, exp time.Duration, cb CallbackFunc) (v any, exist bool) {
	var b = c.getBucket(key)
	b.Lock()
	defer b.Unlock()

	expireAt := c.getExp(exp)
	result, ok := c.fetch(b, key)
	if ok {
		result.ExpireAt = expireAt
		b.Heap.Down(result.index, b.Heap.Len())
		return result.Value, true
	}

	var ele = &Element{Key: key, Value: value, ExpireAt: expireAt, cb: cb}
	b.Heap.Push(ele)
	b.Map[key] = ele
	c.overflow(b)
	return value, false
}

// Delete
func (c *MemoryCache) Delete(key string) (deleted bool) {
	var b = c.getBucket(key)
	b.Lock()
	defer b.Unlock()

	v, ok := c.fetch(b, key)
	if !ok {
		return false
	}

	b.Heap.Delete(v.index)
	delete(b.Map, key)
	v.cb(v, ReasonDeleted)
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
			if !v.expired(now) && strings.HasPrefix(v.Key, prefix) {
				arr = append(arr, v.Key)
			}
		}
		b.Unlock()
	}
	return arr
}

// Len 获取当前元素数量
// Get the number of Elements
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
	Map  map[string]*Element
	Heap *heap
}

// 过期时间检查
func (c *bucket) ExpireCheck(now int64, num int) int {
	c.Lock()
	defer c.Unlock()

	var sum = 0
	for c.Heap.Len() > 0 && c.Heap.Front().expired(now) && sum < num {
		head := c.Heap.Pop()
		delete(c.Map, head.Key)
		sum++
		head.cb(head, ReasonExpired)
	}
	return sum
}
