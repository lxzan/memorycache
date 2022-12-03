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
			h: make([]heap.Element, 0),
			m: make(map[string]element),
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
	m map[string]element
	h heap.Heap
}

// 过期时间检查
func (c *bucket) expireCheck(d time.Duration) {
	var ticker = time.NewTicker(d)
	defer ticker.Stop()
	for {
		<-ticker.C

		c.Lock()
		var ts = utils.Timestamp()
		for c.h.Len() > 0 {
			if c.h[0].ExpireAt > ts {
				break
			}

			var ele0 = c.h.Pop()
			ele1, exist := c.m[ele0.Key]
			if !exist || ele0.ExpireAt != ele1.ExpireAt {
				continue
			}
			delete(c.m, ele0.Key)
		}
		c.Unlock()
	}
}
