package benchmark

import (
	"testing"
	"time"

	"github.com/Yiling-J/theine-go"
	"github.com/dgraph-io/ristretto"
	"github.com/lxzan/memorycache"
	"github.com/lxzan/memorycache/internal/utils"
)

const (
	sharding   = 128
	capacity   = 10000
	benchcount = 1 << 20
)

var (
	benchkeys = make([]string, 0, benchcount)

	options = []memorycache.Option{
		memorycache.WithBucketNum(sharding),
		memorycache.WithBucketSize(capacity/10, capacity),
		memorycache.WithSwissTable(false),
		memorycache.WithLRU(true),
	}
)

func init() {
	for i := 0; i < benchcount; i++ {
		benchkeys = append(benchkeys, string(utils.AlphabetNumeric.Generate(16)))
	}
}

func getIndex(i int) int {
	return i & (len(benchkeys) - 1)
}

func BenchmarkMemoryCache_Set(b *testing.B) {
	var mc = memorycache.New[string, int](options...)
	b.RunParallel(func(pb *testing.PB) {
		var i = 0
		for pb.Next() {
			index := getIndex(i)
			i++
			mc.Set(benchkeys[index], 1, time.Hour)
		}
	})
}

func BenchmarkMemoryCache_Get(b *testing.B) {
	var mc = memorycache.New[string, int](options...)
	for i := 0; i < benchcount; i++ {
		mc.Set(benchkeys[i%benchcount], 1, time.Hour)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var i = 0
		for pb.Next() {
			index := getIndex(i)
			i++
			mc.Get(benchkeys[index])
		}
	})
}

func BenchmarkMemoryCache_SetAndGet(b *testing.B) {
	var mc = memorycache.New[string, int](options...)
	for i := 0; i < benchcount; i++ {
		mc.Set(benchkeys[i%benchcount], 1, time.Hour)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var i = 0
		for pb.Next() {
			index := getIndex(i)
			i++
			if index&7 == 0 {
				mc.Set(benchkeys[index], 1, time.Hour)
			} else {
				mc.Get(benchkeys[index])
			}
		}
	})
}

func BenchmarkRistretto_Set(b *testing.B) {
	var mc, _ = ristretto.NewCache(&ristretto.Config{
		NumCounters: capacity * sharding * 10, // number of keys to track frequency of (10M).
		MaxCost:     1 << 30,                  // maximum cost of cache (1GB).
		BufferItems: 64,                       // number of keys per Get buffer.
	})
	b.RunParallel(func(pb *testing.PB) {
		var i = 0
		for pb.Next() {
			index := getIndex(i)
			i++
			mc.SetWithTTL(benchkeys[index], 1, 1, time.Hour)
		}
	})
}

func BenchmarkRistretto_Get(b *testing.B) {
	var mc, _ = ristretto.NewCache(&ristretto.Config{
		NumCounters: capacity * sharding * 10, // number of keys to track frequency of (10M).
		MaxCost:     1 << 30,                  // maximum cost of cache (1GB).
		BufferItems: 64,                       // number of keys per Get buffer.
	})
	for i := 0; i < benchcount; i++ {
		mc.SetWithTTL(benchkeys[i%benchcount], 1, 1, time.Hour)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var i = 0
		for pb.Next() {
			index := getIndex(i)
			i++
			mc.Get(benchkeys[index])
		}
	})
}

func BenchmarkRistretto_SetAndGet(b *testing.B) {
	var mc, _ = ristretto.NewCache(&ristretto.Config{
		NumCounters: capacity * sharding * 10, // number of keys to track frequency of (10M).
		MaxCost:     1 << 30,                  // maximum cost of cache (1GB).
		BufferItems: 64,                       // number of keys per Get buffer.
	})
	for i := 0; i < benchcount; i++ {
		mc.SetWithTTL(benchkeys[i%benchcount], 1, 1, time.Hour)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var i = 0
		for pb.Next() {
			index := getIndex(i)
			i++
			if index&7 == 0 {
				mc.SetWithTTL(benchkeys[index], 1, 1, time.Hour)
			} else {
				mc.Get(benchkeys[index])
			}
		}
	})
}

func BenchmarkTheine_Set(b *testing.B) {
	mc, _ := theine.NewBuilder[string, int](sharding * capacity).Build()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			index := getIndex(i)
			i++
			mc.SetWithTTL(benchkeys[index], 1, 1, time.Hour)
		}
	})
}

func BenchmarkTheine_Get(b *testing.B) {
	mc, _ := theine.NewBuilder[string, int](sharding * capacity).Build()
	for i := 0; i < benchcount; i++ {
		mc.SetWithTTL(benchkeys[i%benchcount], 1, 1, time.Hour)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			index := getIndex(i)
			i++
			mc.Get(benchkeys[index])
		}
	})
}

func BenchmarkTheine_SetAndGet(b *testing.B) {
	mc, _ := theine.NewBuilder[string, int](sharding * capacity).Build()
	for i := 0; i < benchcount; i++ {
		mc.SetWithTTL(benchkeys[i%benchcount], 1, 1, time.Hour)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			index := getIndex(i)
			i++
			if index&7 == 0 {
				mc.SetWithTTL(benchkeys[index], 1, 1, time.Hour)
			} else {
				mc.Get(benchkeys[index])
			}
		}
	})
}
