package hashmap

import (
	"github.com/lxzan/memorycache/internal/heap"
	"github.com/lxzan/memorycache/internal/types"
	"github.com/lxzan/memorycache/internal/utils"
	"sync"
	"time"
)

type (
	ConcurrentMap struct {
		Segment uint32
		Buckets []*Bucket
	}

	Bucket struct {
		sync.RWMutex
		Map  map[string]types.Element
		Heap heap.Heap
	}
)

func NewConcurrentMap(segment uint32, interval time.Duration) *ConcurrentMap {
	var m = &ConcurrentMap{
		Segment: segment,
		Buckets: make([]*Bucket, segment),
	}
	for i, _ := range m.Buckets {
		m.Buckets[i] = &Bucket{
			Heap: make([]heap.Element, 0),
			Map:  make(map[string]types.Element),
		}
	}
	for i, _ := range m.Buckets {
		go m.Buckets[i].expireCheck(interval)
	}
	return m
}

func (c *ConcurrentMap) GetBucket(key string) *Bucket {
	var idx = utils.NewFnv32([]byte(key)) & (c.Segment - 1)
	return c.Buckets[idx]
}

// 过期时间检查
func (c *Bucket) expireCheck(d time.Duration) {
	var ticker = time.NewTicker(d)
	defer ticker.Stop()
	for {
		<-ticker.C

		c.Lock()
		var ts = utils.Timestamp()
		for c.Heap.Len() > 0 {
			if c.Heap[0].ExpireAt > ts {
				break
			}

			var ele0 = c.Heap.Pop()
			ele1, exist := c.Map[ele0.Key]
			if !exist || ele0.ExpireAt != ele1.ExpireAt {
				continue
			}
			delete(c.Map, ele0.Key)
		}
		c.Unlock()
	}
}
