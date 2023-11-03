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

type Timer struct {
	now atomic.Int64
}

type MemoryCache struct {
	config    *config
	storage   []*bucket
	seed      maphash.Seed
	timer     Timer
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	closeOnce sync.Once
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
		config:    c,
		storage:   make([]*bucket, c.BucketNum),
		seed:      maphash.MakeSeed(),
		timer:     Timer{now: atomic.Int64{}},
		wg:        sync.WaitGroup{},
		closeOnce: sync.Once{},
	}

	mc.ctx, mc.cancel = context.WithCancel(context.Background())
	mc.timer.now.Store(time.Now().UnixMilli())

	for i, _ := range mc.storage {
		mc.storage[i] = &bucket{
			Map:  make(map[string]*Element, c.InitialSize),
			heap: newHeap(c.InitialSize),
		}
	}

	mc.wg.Add(2)

	go func() {
		var d0 = c.MaxInterval
		var ticker = time.NewTicker(d0)
		defer func() {
			ticker.Stop()
			mc.wg.Done()
		}()

		for {
			select {
			case <-mc.ctx.Done():
				return

			case <-ticker.C:
				var sum = 0
				var now = time.Now().UnixMilli()
				for _, b := range mc.storage {
					sum += b.expireTimeCheck(now, c.MaxKeysDeleted)
				}

				if d1 := utils.SelectValue(sum > c.BucketNum*c.MaxKeysDeleted*7/10, c.MinInterval, c.MaxInterval); d1 != d0 {
					d0 = d1
					ticker.Reset(d0)
				}
			}
		}
	}()

	go func() {
		var ticker = time.NewTicker(time.Second)

		defer func() {
			ticker.Stop()
			mc.wg.Done()
		}()

		for {
			select {
			case <-mc.ctx.Done():
				return

			case <-ticker.C:
				mc.timer.now.Store(time.Now().UnixMilli())
			}
		}
	}()

	return mc
}

func (c *MemoryCache) Close() {
	c.closeOnce.Do(func() {
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

	return c.timer.now.Load() + d.Milliseconds()
}

// Clear 清空所有缓存
// clear all caches
func (c *MemoryCache) Clear() {
	for _, b := range c.storage {
		b.Lock()
		b.heap = newHeap(c.config.InitialSize)
		b.Map = make(map[string]*Element, c.config.InitialSize)
		b.Unlock()
	}
}

// Set 设置键值和过期时间. exp<=0表示永不过期.
// Set the key value and expiration time. exp<=0 means never expire.
func (c *MemoryCache) Set(key string, value any, exp time.Duration) (replaced bool) {
	return c.SetWithCallback(key, value, exp, emptyCallbackFunc)
}

// SetWithCallback 设置键值, 过期时间和回调函数. 容量溢出和过期都会触发回调.
// Set the key value, expiration time and callback function. The callback is triggered by both capacity overflow and expiration.
func (c *MemoryCache) SetWithCallback(key string, value any, exp time.Duration, cb CallbackFunc) (replaced bool) {
	var b = c.getBucket(key)
	b.Lock()
	defer b.Unlock()

	var expireAt = c.getExp(exp)
	v, ok := b.Map[key]
	if ok {
		v.Value = value
		v.ExpireAt = expireAt
		v.cb = cb
		b.heap.Down(v.index, b.heap.Len())
		return true
	}

	var ele = &Element{Key: key, Value: value, ExpireAt: expireAt, cb: cb}
	b.heap.Push(ele)
	b.Map[key] = ele
	if b.heap.Len() > c.config.MaxCapacity {
		head := b.heap.Pop()
		delete(b.Map, head.Key)
		head.cb(head, ReasonOverflow)
	}
	return false
}

// Get
func (c *MemoryCache) Get(key string) (any, bool) {
	var b = c.getBucket(key)
	b.Lock()
	v, exist := b.Map[key]
	b.Unlock()
	if !exist || v.expired(c.timer.now.Load()) {
		return nil, false
	}
	return v.Value, true
}

// GetWithTTL 获取. 如果存在, 刷新过期时间.
// Get a value. If it exists, refreshes the expiration time.
func (c *MemoryCache) GetWithTTL(key string, exp time.Duration) (any, bool) {
	var b = c.getBucket(key)
	b.Lock()
	defer b.Unlock()

	v, exist := b.Map[key]
	if !exist || v.expired(c.timer.now.Load()) {
		return nil, false
	}

	v.ExpireAt = c.getExp(exp)
	b.heap.Down(v.index, b.heap.Len())
	return v.Value, true
}

// GetOrCreate 如果存在, 刷新过期时间. 如果不存在, 创建一个新的.
// Get or create a value. If it exists, refreshes the expiration time. If it does not exist, creates a new one.
func (c *MemoryCache) GetOrCreate(key string, value any, exp time.Duration) (any, bool) {
	return c.GetOrCreateWithCallback(key, value, exp, emptyCallbackFunc)
}

// GetOrCreate 如果存在, 刷新过期时间. 如果不存在, 创建一个新的.
// Get or create a value with CallbackFunc. If it exists, refreshes the expiration time. If it does not exist, creates a new one.
func (c *MemoryCache) GetOrCreateWithCallback(key string, value any, exp time.Duration, cb CallbackFunc) (any, bool) {
	var b = c.getBucket(key)
	b.Lock()
	defer b.Unlock()

	v, exist := b.Map[key]
	if !exist {
		expireAt := c.getExp(exp)
		ele := &Element{Key: key, Value: value, ExpireAt: expireAt, cb: cb}
		b.heap.Push(ele)
		b.Map[key] = ele
		if b.heap.Len() > c.config.MaxCapacity {
			head := b.heap.Pop()
			delete(b.Map, head.Key)
			head.cb(head, ReasonOverflow)
		}
		return value, true
	}

	if v.expired(c.timer.now.Load()) {
		return nil, false
	}

	v.ExpireAt = c.getExp(exp)
	b.heap.Down(v.index, b.heap.Len())
	return v.Value, true
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

	b.heap.Delete(v.index)
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
		for _, v := range b.heap.Data {
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
		num += b.heap.Len()
		b.Unlock()
	}
	return num
}

type bucket struct {
	sync.Mutex
	Map  map[string]*Element
	heap *heap
}

// 过期时间检查
func (c *bucket) expireTimeCheck(now int64, num int) int {
	c.Lock()
	defer c.Unlock()

	var sum = 0
	for c.heap.Len() > 0 && c.heap.Front().expired(now) && sum < num {
		head := c.heap.Pop()
		delete(c.Map, head.Key)
		sum++
		head.cb(head, ReasonExpired)
	}
	return sum
}
