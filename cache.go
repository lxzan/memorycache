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
	conf       *config
	storage    []*bucket[K, V]
	hasher     utils.Hasher[K]
	timestamp  atomic.Int64
	bucketMask uint64
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	once       sync.Once
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
		conf:       conf,
		storage:    make([]*bucket[K, V], conf.BucketNum),
		hasher:     maphash.NewHasher[K](),
		bucketMask: uint64(conf.BucketNum - 1),
		wg:         sync.WaitGroup{},
		once:       sync.Once{},
	}
	mc.ctx, mc.cancel = context.WithCancel(context.Background())
	mc.timestamp.Store(time.Now().UnixMilli())

	for i := range mc.storage {
		b := (&bucket[K, V]{conf: conf}).init()
		mc.storage[i] = b
	}

	mc.wg.Add(2)

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
				if sum > 0 {
					d1 := utils.SelectValue(sum > conf.BucketNum*conf.DeleteLimits*7/10, conf.MinInterval, conf.MaxInterval)
					if d1 != d0 {
						d0 = d1
						ticker.Reset(d0)
					}
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

// Clear 清空缓存. 对存活非过期元素触发 ReasonCleared 回调.
// Clear caches. Triggers ReasonCleared callback for alive non-expired elements.
func (c *MemoryCache[K, V]) Clear() {
	var now = c.getTimestamp()
	for _, b := range c.storage {
		b.Lock()
		b.List.Range(func(ele *Element[K, V]) bool {
			if !ele.expired(now) && ele.cb != nil {
				ele.cb(ele, ReasonCleared)
			}
			return true
		})
		b.init()
		b.Unlock()
	}
}

func (c *MemoryCache[K, V]) Stop() {
	c.once.Do(func() {
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

// Set 设置键值和过期时间. exp<=0表示永不过期.
// Set the key value and expiration time. exp<=0 means never expire.
func (c *MemoryCache[K, V]) Set(key K, value V, exp time.Duration) (exist bool) {
	hashcode := c.hasher.Hash(key)
	b := c.storage[hashcode&c.bucketMask]
	b.Lock()

	var expireAt = c.getExp(exp)
	addr, ok := b.Map.Get(hashcode)
	if ok {
		ele := b.List.Get(addr)
		if ele.expired(c.timestamp.Load()) {
			b.Delete(ele, ReasonExpired)
		} else if key != ele.Key {
			b.Delete(ele, ReasonEvicted)
		} else {
			ele.Value = value
			if ele.ExpireAt != expireAt {
				b.Heap.UpdateTTL(ele, expireAt)
				b.List.MoveToBack(ele.addr)
			}
			b.Unlock()
			return true
		}
	}

	ele := b.GetElement()
	ele.Key, ele.Value, ele.ExpireAt, ele.hashcode = key, value, expireAt, hashcode
	b.Insert(ele)
	b.Unlock()
	return false
}

// SetWithCallback 设置键值, 过期时间和回调函数. 容量溢出和过期都会触发回调.
// 注意: 不要在回调函数里面操作 MemoryCache 实例, 可能会造成死锁.
// Set the key value, expiration time and callback function. The callback is triggered by both capacity overflow and expiration.
// Note: Don't manipulate MemoryCache instances in callback functions, as this may cause deadlocks.
func (c *MemoryCache[K, V]) SetWithCallback(key K, value V, exp time.Duration, cb CallbackFunc[*Element[K, V]]) (exist bool) {
	hashcode := c.hasher.Hash(key)
	b := c.storage[hashcode&c.bucketMask]
	b.Lock()

	var expireAt = c.getExp(exp)
	addr, ok := b.Map.Get(hashcode)
	if ok {
		ele := b.List.Get(addr)
		if ele.expired(c.timestamp.Load()) {
			b.Delete(ele, ReasonExpired)
		} else if key != ele.Key {
			b.Delete(ele, ReasonEvicted)
		} else {
			ele.Value, ele.cb = value, cb
			if ele.ExpireAt != expireAt {
				b.Heap.UpdateTTL(ele, expireAt)
				b.List.MoveToBack(ele.addr)
			}
			b.Unlock()
			return true
		}
	}

	ele := b.GetElement()
	ele.Key, ele.Value, ele.ExpireAt, ele.hashcode, ele.cb = key, value, expireAt, hashcode, cb
	b.Insert(ele)
	b.Unlock()
	return false
}

// Get 查询缓存
// query cache
func (c *MemoryCache[K, V]) Get(key K) (v V, exist bool) {
	hashcode := c.hasher.Hash(key)
	b := c.storage[hashcode&c.bucketMask]
	b.Lock()

	addr, ok := b.Map.Get(hashcode)
	if !ok {
		b.Unlock()
		return v, false
	}
	ele := b.List.Get(addr)
	if ele.expired(c.timestamp.Load()) {
		b.Delete(ele, ReasonExpired)
		b.Unlock()
		return v, false
	}
	if key != ele.Key {
		b.Unlock()
		return v, false
	}

	if ele.next != 0 && ele.prev != 0 {
		list := b.List
		elements := list.elements
		prev := &elements[ele.prev]
		next := &elements[ele.next]
		prev.next = ele.next
		next.prev = ele.prev

		tail := &elements[list.tail]
		tail.next = ele.addr
		ele.prev = list.tail
		ele.next = 0
		list.tail = ele.addr

		b.Unlock()
		return ele.Value, true
	}

	b.List.MoveToBack(ele.addr)
	b.Unlock()
	return ele.Value, true
}

// GetTTL 获取 key 的过期时间
// Get the expiration time of the key.
func (c *MemoryCache[K, V]) GetTTL(key K) (time.Time, bool) {
	hashcode := c.hasher.Hash(key)
	b := c.storage[hashcode&c.bucketMask]
	b.Lock()

	addr, ok := b.Map.Get(hashcode)
	if !ok {
		b.Unlock()
		return time.Time{}, false
	}
	ele := b.List.Get(addr)
	if ele.expired(c.getTimestamp()) {
		b.Unlock()
		return time.Time{}, false
	}
	if key != ele.Key {
		b.Unlock()
		return time.Time{}, false
	}
	b.Unlock()
	if ele.ExpireAt == math.MaxInt64 {
		return time.Time{}, true
	}
	return time.UnixMilli(ele.ExpireAt), true
}

// GetWithTTL 获取. 如果存在, 刷新过期时间.
// Get a value. If it exists, refreshes the expiration time.
func (c *MemoryCache[K, V]) GetWithTTL(key K, exp time.Duration) (v V, exist bool) {
	hashcode := c.hasher.Hash(key)
	b := c.storage[hashcode&c.bucketMask]
	b.Lock()

	addr, ok := b.Map.Get(hashcode)
	if !ok {
		b.Unlock()
		return v, false
	}
	ele := b.List.Get(addr)
	if ele.expired(c.timestamp.Load()) {
		b.Delete(ele, ReasonExpired)
		b.Unlock()
		return v, false
	}
	if key != ele.Key {
		b.Unlock()
		return v, false
	}

	var expireAt = c.getExp(exp)
	if ele.ExpireAt != expireAt {
		b.Heap.UpdateTTL(ele, expireAt)
		b.List.MoveToBack(ele.addr)
	}
	b.Unlock()
	return ele.Value, true
}

// UpdateTTL 更新 key 的过期时间. d<=0 表示永不过期.
// Update the expiration time of the key. d<=0 means never expire.
func (c *MemoryCache[K, V]) UpdateTTL(key K, d time.Duration) bool {
	hashcode := c.hasher.Hash(key)
	b := c.storage[hashcode&c.bucketMask]
	b.Lock()

	addr, ok := b.Map.Get(hashcode)
	if !ok {
		b.Unlock()
		return false
	}
	ele := b.List.Get(addr)
	if ele.expired(c.timestamp.Load()) {
		b.Unlock()
		return false
	}
	if key != ele.Key {
		b.Unlock()
		return false
	}
	var expireAt = c.getExp(d)
	if ele.ExpireAt != expireAt {
		b.Heap.UpdateTTL(ele, expireAt)
		b.List.MoveToBack(ele.addr)
	}
	b.Unlock()
	return true
}

// GetOrCreate 如果存在, 返回已有值(忽略 value 参数)并刷新过期时间. 如果不存在, 创建一个新的.
// Get or create a value. If it exists, returns the existing value (ignoring the value parameter) and refreshes the expiration time. If it does not exist, creates a new one.
func (c *MemoryCache[K, V]) GetOrCreate(key K, value V, exp time.Duration) (v V, exist bool) {
	return c.GetOrCreateWithCallback(key, value, exp, nil)
}

// GetOrCreateWithCallback 如果存在, 返回已有值(忽略 value 和 cb 参数)并刷新过期时间. 如果不存在, 创建一个新的.
// 注意: 不要在回调函数里面操作 MemoryCache 实例, 可能会造成死锁.
// Get or create a value with CallbackFunc. If it exists, returns the existing value (ignoring the value and cb parameters) and refreshes the expiration time. If it does not exist, creates a new one.
// Note: Don't manipulate MemoryCache instances in callback functions, as this may cause deadlocks.
func (c *MemoryCache[K, V]) GetOrCreateWithCallback(key K, value V, exp time.Duration, cb CallbackFunc[*Element[K, V]]) (v V, exist bool) {
	hashcode := c.hasher.Hash(key)
	b := c.storage[hashcode&c.bucketMask]
	b.Lock()

	var expireAt = c.getExp(exp)
	addr, ok := b.Map.Get(hashcode)
	if ok {
		ele := b.List.Get(addr)
		if ele.expired(c.timestamp.Load()) {
			b.Delete(ele, ReasonExpired)
		} else if key != ele.Key {
			b.Delete(ele, ReasonEvicted)
		} else {
			if ele.ExpireAt != expireAt {
				b.Heap.UpdateTTL(ele, expireAt)
				b.List.MoveToBack(ele.addr)
			}
			b.Unlock()
			return ele.Value, true
		}
	}

	ele := b.GetElement()
	ele.Key, ele.Value, ele.ExpireAt, ele.hashcode, ele.cb = key, value, expireAt, hashcode, cb
	b.Insert(ele)
	b.Unlock()
	return value, false
}

// Delete 删除缓存
// delete cache
func (c *MemoryCache[K, V]) Delete(key K) (exist bool) {
	hashcode := c.hasher.Hash(key)
	b := c.storage[hashcode&c.bucketMask]
	b.Lock()

	addr, ok := b.Map.Get(hashcode)
	if !ok {
		b.Unlock()
		return false
	}
	ele := b.List.Get(addr)
	if ele.expired(c.timestamp.Load()) {
		b.Delete(ele, ReasonExpired)
		b.Unlock()
		return false
	}
	if key != ele.Key {
		b.Unlock()
		return false
	}

	b.Delete(ele, ReasonDeleted)
	b.Unlock()
	return true
}

// Range 遍历缓存
// 注意: 不要在回调函数里面操作 MemoryCache[K, V] 实例, 可能会造成死锁.
// Traverse the cache.
// Note: Do not manipulate MemoryCache[K, V] instances inside callback functions, as this may cause deadlocks.
func (c *MemoryCache[K, V]) Range(f func(K, V) bool) {
	var now = c.getTimestamp()
	for _, b := range c.storage {
		b.Lock()
		stopped := false
		b.List.Range(func(ele *Element[K, V]) bool {
			if ele.expired(now) {
				return true
			}
			if !f(ele.Key, ele.Value) {
				stopped = true
				return false
			}
			return true
		})
		b.Unlock()
		if stopped {
			return
		}
	}
}

// Len 快速获取当前缓存元素数量, 不做过期检查.
// 注意: 返回值可能包含已过期但未清除的元素.
// Quickly gets the current number of cached elements, without checking for expiration.
// Note: the returned count may include expired elements that have not been cleaned up yet.
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
	conf *config
	Map  containers.Map[uint64, pointer]
	Heap *heap[K, V]
	List *deque[K, V]
}

func (c *bucket[K, V]) init() *bucket[K, V] {
	c.Map = containers.NewMap[uint64, pointer](c.conf.BucketSize, c.conf.SwissTable)
	c.List = newDeque[K, V](c.conf.BucketSize)
	c.Heap = newHeap[K, V](c.List, c.conf.BucketSize)
	return c
}

// Check 过期时间检查
func (c *bucket[K, V]) Check(now int64, num int) int {
	c.Lock()
	defer c.Unlock()

	var sum = 0
	for c.Heap.Len() > 0 && sum < num {
		ele := c.Heap.Front()
		if ele == nil || !ele.expired(now) {
			break
		}
		c.Delete(ele, ReasonExpired)
		sum++
	}
	return sum
}

func (c *bucket[K, V]) Delete(ele *Element[K, V], reason Reason) {
	c.Heap.Delete(ele.index)
	c.Map.Delete(ele.hashcode)
	if ele.cb != nil {
		ele.cb(ele, reason)
	}
	c.List.Remove(ele.addr) // 必须最后删除List, 因为会清空*Element[K, V]数据
}

func (c *bucket[K, V]) UpdateTTL(ele *Element[K, V], expireAt int64) {
	if ele.ExpireAt == expireAt {
		return
	}
	c.Heap.UpdateTTL(ele, expireAt)
	c.List.MoveToBack(ele.addr)
}

func (c *bucket[K, V]) GetElement() *Element[K, V] {
	if c.List.Len() >= c.conf.BucketCap {
		c.Delete(c.List.Front(), ReasonEvicted)
	}
	return c.List.PushBack()
}

func (c *bucket[K, V]) Insert(ele *Element[K, V]) {
	c.Heap.Push(ele)
	c.Map.Put(ele.hashcode, ele.addr)
}
