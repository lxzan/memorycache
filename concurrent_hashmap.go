package memorycache

import (
	"github.com/lxzan/memorycache/internal/heap"
	"github.com/lxzan/memorycache/internal/utils"
	"math/rand"
	"sync"
	"time"
	"unsafe"
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
			clear_count: cfg.ClearKeysCount,
			ttl:         make([]heap.Element, 0),
			data:        make(map[string]element),
		}
	}
	for i, _ := range m.buckets {
		var d = cfg.TTLCheckInterval + rand.Intn(cfg.TTLCheckInterval)
		go m.buckets[i].expireCheck(time.Duration(d) * time.Second)
	}
	return m
}

func (self concurrent_hashmap) getBucket(key *string) *bucket {
	var k = *(*[]byte)(unsafe.Pointer(key))
	var idx = utils.NewFnv32(k) & (self.cfg.Segment - 1)
	return &self.buckets[idx]
}

type bucket struct {
	sync.RWMutex
	data        map[string]element
	ttl         heap.Heap
	clear_count uint32
}

// 过期时间检查
func (self *bucket) expireCheck(d time.Duration) {
	var ticker = time.NewTicker(d)
	defer ticker.Stop()
	for {
		<-ticker.C

		self.Lock()
		var num uint32 = 0
		var ts = time.Now().UnixMilli()
		for self.ttl.Len() > 0 {
			if num >= self.clear_count || self.ttl[0].ExpireAt > ts {
				break
			}

			var ele0 = self.ttl.Pop()
			ele1, exist := self.data[ele0.Key]
			if !exist || ele1.ExpireAt > ts {
				continue
			}
			delete(self.data, ele0.Key)
			num++
		}
		self.Unlock()
	}
}
