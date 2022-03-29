package memdb

import (
	"math/rand"
	"memdb/internal/heap"
	"memdb/internal/utils"
	"sync"
	"time"
	"unsafe"
)

const (
	expireCheckInterval = 10  // 过期时间检查间隔, 秒
	expireCheckNum      = 100 // 每次过期检查清除数据量
)

type concurrent_hashmap struct {
	segment uint32
	buckets []bucket
}

func newConcurrentHashmap(segment uint32) *concurrent_hashmap {
	var m = &concurrent_hashmap{
		segment: segment,
		buckets: make([]bucket, segment),
	}
	for i, _ := range m.buckets {
		m.buckets[i] = bucket{
			ttl:  make([]heap.Element, 0),
			data: make(map[string]element),
		}
	}
	for i, _ := range m.buckets {
		var d = expireCheckInterval + rand.Intn(expireCheckInterval/2)
		go m.buckets[i].expireCheck(time.Duration(d) * time.Second)
	}
	return m
}

func (self concurrent_hashmap) getBucket(key string) *bucket {
	var k = *(*[]byte)(unsafe.Pointer(&key))
	var idx = utils.NewFnv32(k) % self.segment
	return &self.buckets[idx]
}

type bucket struct {
	sync.RWMutex
	data map[string]element
	ttl  heap.Heap
}

// 过期时间检查
func (self *bucket) expireCheck(d time.Duration) {
	var ticker = time.NewTicker(d)
	defer ticker.Stop()
	for {
		<-ticker.C

		self.Lock()
		var num = 0
		var ts = time.Now().UnixMilli()
		for self.ttl.Len() > 0 {
			if num >= expireCheckNum || self.ttl[0].ExpireAt > ts {
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
