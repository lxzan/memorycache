package memorycache

import (
	"context"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dolthub/maphash"
	"github.com/lxzan/dao/deque"
	"github.com/lxzan/memorycache/internal/containers"
	"github.com/lxzan/memorycache/internal/utils"
)

type MemoryCache[K comparable, V any] struct {
	conf      *config
	storage   []*bucket[K, V]
	hasher    utils.Hasher[K]
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
	var conf = &config{CachedTime: true}
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
		b := (&bucket[K, V]{conf: conf}).init()
		mc.storage[i] = b
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
					sum += b.Check(now.UnixMilli(), conf.DeleteLimits)
				}

				// 删除数量超过阈值, 缩小时间间隔
				if d1 := utils.SelectValue(sum > conf.BucketNum*conf.DeleteLimits*7/10, conf.MinInterval, conf.MaxInterval); d1 != d0 {
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
		b.init()
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

func (c *MemoryCache[K, V]) getTimestamp() int64 {
	if c.conf.CachedTime {
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

func (c *MemoryCache[K, V]) getBucket(key K) bucketWrapper[K, V] {
	var hashcode = c.hasher.Hash(key)
	var index = hashcode & uint64(c.conf.BucketNum-1)
	return bucketWrapper[K, V]{bucket: c.storage[index], hashcode: hashcode}
}

// 查找数据. 如果存在且超时, 删除并返回false
// @ele 查找结果
// @conflict 是否哈希冲突
// @exist 是否存在
func (c *MemoryCache[K, V]) fetch(b bucketWrapper[K, V], key K) (ele *Element[K, V], conflict, exist bool) {
	addr, ok := b.Map.Get(b.hashcode)
	if !ok {
		return nil, false, false
	}

	ele = b.List.Get(addr).Value()
	if ele.expired(c.getTimestamp()) {
		b.Delete(ele, ReasonExpired)
		return nil, false, false
	}

	return ele, key != ele.Key, true
}

// Set 设置键值和过期时间. exp<=0表示永不过期.
// Set the key value and expiration time. exp<=0 means never expire.
func (c *MemoryCache[K, V]) Set(key K, value V, exp time.Duration) (exist bool) {
	return c.SetWithCallback(key, value, exp, c.callback)
}

// SetWithCallback 设置键值, 过期时间和回调函数. 容量溢出和过期都会触发回调.
// Set the key value, expiration time and callback function. The callback is triggered by both capacity overflow and expiration.
func (c *MemoryCache[K, V]) SetWithCallback(key K, value V, exp time.Duration, cb CallbackFunc[*Element[K, V]]) (exist bool) {
	var b = c.getBucket(key)
	b.Lock()
	defer b.Unlock()

	var expireAt = c.getExp(exp)
	ele, conflict, ok := c.fetch(b, key)
	if conflict {
		ok = false
		b.Delete(ele, ReasonEvicted)
	}
	if ok {
		ele.Value, ele.cb = value, cb
		b.UpdateTTL(ele, expireAt)
		return true
	}

	ele = &Element[K, V]{Key: key, Value: value, ExpireAt: expireAt, cb: cb, hashcode: b.hashcode}
	b.Insert(ele)
	return false
}

// Get 查询缓存
// query cache
func (c *MemoryCache[K, V]) Get(key K) (v V, exist bool) {
	var b = c.getBucket(key)
	b.Lock()
	defer b.Unlock()

	ele, conflict, ok := c.fetch(b, key)
	if !ok || conflict {
		return v, false
	}

	b.List.MoveToBack(ele.addr)
	return ele.Value, true
}

// GetWithTTL 获取. 如果存在, 刷新过期时间.
// Get a value. If it exists, refreshes the expiration time.
func (c *MemoryCache[K, V]) GetWithTTL(key K, exp time.Duration) (v V, exist bool) {
	var b = c.getBucket(key)
	b.Lock()
	defer b.Unlock()

	ele, conflict, ok := c.fetch(b, key)
	if !ok || conflict {
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
	ele, conflict, ok := c.fetch(b, key)
	if conflict {
		ok = false
		b.Delete(ele, ReasonEvicted)
	}
	if ok {
		b.UpdateTTL(ele, expireAt)
		return ele.Value, true
	}

	ele = &Element[K, V]{Key: key, Value: value, ExpireAt: expireAt, cb: cb, hashcode: b.hashcode}
	b.Insert(ele)
	return value, false
}

// Delete 删除缓存
// delete cache
func (c *MemoryCache[K, V]) Delete(key K) (exist bool) {
	var b = c.getBucket(key)
	b.Lock()
	defer b.Unlock()

	ele, conflict, ok := c.fetch(b, key)
	if ok && !conflict {
		b.Delete(ele, ReasonDeleted)
		return true
	}

	return false
}

// Range 遍历缓存. 注意: 不要在回调函数里面操作 MemoryCache[K, V] 实例, 可能会造成死锁.
// Traverse the cache. Note: Do not manipulate MemoryCache[K, V] instances inside callback functions, as this may cause deadlocks.
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

// Len 快速获取当前缓存元素数量, 不做过期检查.
// Quickly gets the current number of cached elements, without checking for expiration.
func (c *MemoryCache[K, V]) Len() int {
	var num = 0
	for _, b := range c.storage {
		b.Lock()
		num += b.Heap.Len()
		b.Unlock()
	}
	return num
}

type (
	bucket[K comparable, V any] struct {
		sync.Mutex
		conf *config
		Map  containers.Map[uint64, deque.Pointer]
		Heap *heap[K, V]
		List *deque.Deque[*Element[K, V]]
	}

	bucketWrapper[K comparable, V any] struct {
		*bucket[K, V]
		hashcode uint64
	}
)

func (c *bucket[K, V]) init() *bucket[K, V] {
	c.Map = containers.NewMap[uint64, deque.Pointer](c.conf.BucketSize, c.conf.SwissTable)
	c.Heap = newHeap[K, V](c.conf.BucketSize)
	c.List = deque.New[*Element[K, V]](c.conf.BucketSize)
	return c
}

// Check 过期时间检查
func (c *bucket[K, V]) Check(now int64, num int) int {
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
	c.List.Remove(ele.addr)
	c.Heap.Delete(ele.index)
	c.Map.Delete(ele.hashcode)
	ele.cb(ele, reason)
}

func (c *bucket[K, V]) UpdateTTL(ele *Element[K, V], expireAt int64) {
	c.Heap.UpdateTTL(ele, expireAt)
	c.List.MoveToBack(ele.addr)
}

func (c *bucket[K, V]) Insert(ele *Element[K, V]) {
	if c.List.Len() >= c.conf.BucketCap {
		c.Delete(c.List.Front().Value(), ReasonEvicted)
	}

	ele.addr = c.List.PushBack(ele).Addr()
	c.Heap.Push(ele)
	c.Map.Put(ele.hashcode, ele.addr)
}
