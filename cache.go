package memorycache

import (
	"context"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dolthub/maphash"
	"github.com/lxzan/memorycache/internal/containers"
	"github.com/lxzan/memorycache/internal/utils"
)

type MemoryCache[K comparable, V any] struct {
	conf      *config
	storage   []*bucket[K, V]
	hasher    maphash.Hasher[K]
	timestamp atomic.Int64
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	once      sync.Once
	callback  CallbackFunc[*Element[K, V]]
}

// New 创建缓存数据库实例
// Creating a Cached Database Instance
func New[K comparable, V any](options ...Option) *MemoryCache[K, V] {
	var conf = &config{TimeCacheEnabled: true}
	options = append(options, withInitialize())
	for _, fn := range options {
		fn(conf)
	}

	mc := &MemoryCache[K, V]{
		conf:    conf,
		storage: make([]*bucket[K, V], conf.BucketNum),
		hasher:  maphash.NewHasher[K](),
		wg:      sync.WaitGroup{},
		once:    sync.Once{},
	}
	mc.callback = func(entry *Element[K, V], reason Reason) {}
	mc.ctx, mc.cancel = context.WithCancel(context.Background())
	mc.timestamp.Store(time.Now().UnixMilli())

	for i, _ := range mc.storage {
		mc.storage[i] = &bucket[K, V]{
			MaxCapacity: conf.MaxCapacity,
			Map:         containers.NewMap[K, *Element[K, V]](conf.InitialSize, conf.SwissTable),
			Heap:        newHeap[K, V](conf.InitialSize),
			List:        new(queue[K, V]),
		}
	}

	go func() {
		var d0 = conf.MaxInterval
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
					sum += b.ExpireCheck(now.UnixMilli(), conf.MaxKeysDeleted)
				}

				// 删除数量超过阈值, 缩小时间间隔
				if d1 := utils.SelectValue(sum > conf.BucketNum*conf.MaxKeysDeleted*7/10, conf.MinInterval, conf.MaxInterval); d1 != d0 {
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

// Clear 清空缓存
// clear caches
func (c *MemoryCache[K, V]) Clear() {
	for _, b := range c.storage {
		b.Lock()
		b.Heap = newHeap[K, V](c.conf.InitialSize)
		b.Map = containers.NewMap[K, *Element[K, V]](c.conf.InitialSize, c.conf.SwissTable)
		b.List = new(queue[K, V])
		b.Unlock()
	}
}

func (c *MemoryCache[K, V]) Stop() {
	c.once.Do(func() {
		c.wg.Add(2)
		c.cancel()
		c.wg.Wait()
	})
}

func (c *MemoryCache[K, V]) getBucket(key K) *bucket[K, V] {
	var idx = c.hasher.Hash(key) & uint64(c.conf.BucketNum-1)
	return c.storage[idx]
}

func (c *MemoryCache[K, V]) getTimestamp() int64 {
	if c.conf.TimeCacheEnabled {
		return c.timestamp.Load()
	}
	return time.Now().UnixMilli()
}

// 获取过期时间, d<=0表示永不过期
func (c *MemoryCache[K, V]) getExp(d time.Duration) int64 {
	if d <= 0 {
		return math.MaxInt64
	}
	return c.getTimestamp() + d.Milliseconds()
}

// 查找数据. 如果存在且超时, 删除并返回false
func (c *MemoryCache[K, V]) fetch(b *bucket[K, V], key K) (*Element[K, V], bool) {
	ele, exist := b.Map.Get(key)
	if !exist {
		return nil, false
	}

	if ele.expired(c.getTimestamp()) {
		b.Delete(ele, ReasonExpired)
		return nil, false
	}

	return ele, true
}

// Set 设置键值和过期时间. exp<=0表示永不过期.
// Set the key value and expiration time. exp<=0 means never expire.
func (c *MemoryCache[K, V]) Set(key K, value V, exp time.Duration) (replaced bool) {
	return c.SetWithCallback(key, value, exp, c.callback)
}

// SetWithCallback 设置键值, 过期时间和回调函数. 容量溢出和过期都会触发回调.
// Set the key value, expiration time and callback function. The callback is triggered by both capacity overflow and expiration.
func (c *MemoryCache[K, V]) SetWithCallback(key K, value V, exp time.Duration, cb CallbackFunc[*Element[K, V]]) (replaced bool) {
	var b = c.getBucket(key)
	b.Lock()
	defer b.Unlock()

	var expireAt = c.getExp(exp)
	ele, ok := c.fetch(b, key)
	if ok {
		b.UpdateAll(ele, value, expireAt, cb)
		return true
	}

	b.Insert(key, value, expireAt, cb)
	return false
}

// Get
func (c *MemoryCache[K, V]) Get(key K) (v V, exist bool) {
	var b = c.getBucket(key)
	b.Lock()
	defer b.Unlock()
	ele, ok := c.fetch(c.getBucket(key), key)
	if !ok {
		return v, false
	}
	b.List.MoveToBack(ele)
	return ele.Value, true
}

// GetWithTTL 获取. 如果存在, 刷新过期时间.
// Get a value. If it exists, refreshes the expiration time.
func (c *MemoryCache[K, V]) GetWithTTL(key K, exp time.Duration) (v V, exist bool) {
	var b = c.getBucket(key)
	b.Lock()
	defer b.Unlock()

	ele, ok := c.fetch(b, key)
	if !ok {
		return v, false
	}

	b.UpdateTTL(ele, c.getExp(exp))
	return ele.Value, true
}

// GetOrCreate 如果存在, 刷新过期时间. 如果不存在, 创建一个新的.
// Get or create a value. If it exists, refreshes the expiration time. If it does not exist, creates a new one.
func (c *MemoryCache[K, V]) GetOrCreate(key K, value V, exp time.Duration) (v V, exist bool) {
	return c.GetOrCreateWithCallback(key, value, exp, c.callback)
}

// GetOrCreateWithCallback 如果存在, 刷新过期时间. 如果不存在, 创建一个新的.
// Get or create a value with CallbackFunc. If it exists, refreshes the expiration time. If it does not exist, creates a new one.
func (c *MemoryCache[K, V]) GetOrCreateWithCallback(key K, value V, exp time.Duration, cb CallbackFunc[*Element[K, V]]) (v V, exist bool) {
	var b = c.getBucket(key)
	b.Lock()
	defer b.Unlock()

	expireAt := c.getExp(exp)
	ele, ok := c.fetch(b, key)
	if ok {
		b.UpdateTTL(ele, expireAt)
		return ele.Value, true
	}

	b.Insert(key, value, expireAt, cb)
	return value, false
}

// Delete
func (c *MemoryCache[K, V]) Delete(key K) (deleted bool) {
	var b = c.getBucket(key)
	b.Lock()
	defer b.Unlock()

	ele, ok := c.fetch(b, key)
	if !ok {
		return false
	}

	b.Delete(ele, ReasonDeleted)
	return true
}

// Range
func (c *MemoryCache[K, V]) Range(f func(K, V) bool) {
	var now = time.Now().UnixMilli()
	for _, b := range c.storage {
		b.Lock()
		for _, ele := range b.Heap.Data {
			if ele.expired(now) {
				continue
			}
			if !f(ele.Key, ele.Value) {
				b.Unlock()
				return
			}
		}
		b.Unlock()
	}
}

// Len 获取当前元素数量
// Get the number of Elements
func (c *MemoryCache[K, V]) Len() int {
	var num = 0
	for _, b := range c.storage {
		b.Lock()
		num += b.Heap.Len()
		b.Unlock()
	}
	return num
}

type bucket[K comparable, V any] struct {
	sync.Mutex
	MaxCapacity int
	Map         containers.Map[K, *Element[K, V]]
	Heap        *heap[K, V]
	List        *queue[K, V]
}

// ExpireCheck 过期时间检查
func (c *bucket[K, V]) ExpireCheck(now int64, num int) int {
	c.Lock()
	defer c.Unlock()

	var sum = 0
	for c.Heap.Len() > 0 && c.Heap.Front().expired(now) && sum < num {
		c.Delete(c.Heap.Front(), ReasonExpired)
		sum++
	}
	return sum
}

func (c *bucket[K, V]) Delete(ele *Element[K, V], reason Reason) {
	c.List.Delete(ele)
	c.Heap.Delete(ele.index)
	c.Map.Delete(ele.Key)
	ele.cb(ele, reason)
}

func (c *bucket[K, V]) UpdateAll(ele *Element[K, V], value V, expireAt int64, cb CallbackFunc[*Element[K, V]]) {
	ele.Value = value
	ele.cb = cb
	c.UpdateTTL(ele, expireAt)
}

func (c *bucket[K, V]) UpdateTTL(ele *Element[K, V], expireAt int64) {
	c.Heap.UpdateTTL(ele, expireAt)
	c.List.MoveToBack(ele)
}

func (c *bucket[K, V]) Insert(key K, value V, expireAt int64, cb CallbackFunc[*Element[K, V]]) {
	if c.List.Len() >= c.MaxCapacity {
		c.Delete(c.List.Front(), ReasonEvicted)
	}

	var ele = &Element[K, V]{Key: key, Value: value, ExpireAt: expireAt, cb: cb}
	c.List.PushBack(ele)
	c.Heap.Push(ele)
	c.Map.Put(key, ele)
}
