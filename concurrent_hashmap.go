package memorycache

import (
	"github.com/lxzan/memorycache/internal/heap"
	"github.com/lxzan/memorycache/internal/utils"
	"math/rand"
	"sync"
	"time"
)

type concurrent_hashmap struct {
	cfg     Config
	buckets []bucket
}

func newConcurrentHashmap(cfg Config) *concurrent_hashmap {
	var m = &concurrent_hashmap{
		cfg:     cfg,
		buckets: make([]bucket, cfg.Segment),
	}
	for i, _ := range m.buckets {
		m.buckets[i] = bucket{
			ttl:  make([]heap.Element, 0),
			data: make(map[string]element),
		}
	}
	for i, _ := range m.buckets {
		var d = cfg.TTLCheckInterval + rand.Intn(cfg.TTLCheckInterval)
		go m.buckets[i].expireCheck(time.Duration(d) * time.Second)
	}
	return m
}

func (c *concurrent_hashmap) getBucket(key string) *bucket {
	var idx = utils.NewFnv32([]byte(key)) & (c.cfg.Segment - 1)
	return &c.buckets[idx]
}

type bucket struct {
	sync.RWMutex
	data map[string]element
	ttl  heap.Heap
}

// 过期时间检查
func (c *bucket) expireCheck(d time.Duration) {
	var ticker = time.NewTicker(d)
	defer ticker.Stop()
	for {
		<-ticker.C

		c.Lock()
		var ts = utils.Timestamp()
		for c.ttl.Len() > 0 {
			if c.ttl[0].ExpireAt > ts {
				break
			}

			var ele0 = c.ttl.Pop()
			ele1, exist := c.data[ele0.Key]
			if !exist || ele1.ExpireAt > ts {
				continue
			}
			delete(c.data, ele0.Key)
		}
		c.Unlock()
	}
}
