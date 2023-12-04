package benchmark

import (
	"testing"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/lxzan/memorycache"
	"github.com/lxzan/memorycache/internal/utils"
	"github.com/maypok86/otter"
)

const benchcount = 1000000

var benchkeys = make([]string, 0, benchcount)

func init() {
	for i := 0; i < benchcount; i++ {
		benchkeys = append(benchkeys, string(utils.AlphabetNumeric.Generate(16)))
	}
}

func BenchmarkMemoryCache_Set(b *testing.B) {
	var mc = memorycache.New(
		memorycache.WithBucketNum(128),
		memorycache.WithBucketSize(1000, 10000),
	)
	b.RunParallel(func(pb *testing.PB) {
		var i = 0
		for pb.Next() {
			index := i % benchcount
			i++
			mc.Set(benchkeys[index], 1, time.Hour)
		}
	})
}

func BenchmarkMemoryCache_Get(b *testing.B) {
	var mc = memorycache.New(
		memorycache.WithBucketNum(128),
		memorycache.WithBucketSize(1000, 10000),
	)
	for i := 0; i < benchcount; i++ {
		mc.Set(benchkeys[i%benchcount], 1, time.Hour)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var i = 0
		for pb.Next() {
			index := i % benchcount
			i++
			mc.Get(benchkeys[index])
		}
	})
}

func BenchmarkMemoryCache_SetAndGet(b *testing.B) {
	var mc = memorycache.New(
		memorycache.WithBucketNum(128),
		memorycache.WithBucketSize(1000, 10000),
	)
	for i := 0; i < benchcount; i++ {
		mc.Set(benchkeys[i%benchcount], 1, time.Hour)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var i = 0
		for pb.Next() {
			index := i % benchcount
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
		NumCounters: 10000 * 128 * 10, // number of keys to track frequency of (10M).
		MaxCost:     1 << 30,          // maximum cost of cache (1GB).
		BufferItems: 64,               // number of keys per Get buffer.
	})
	b.RunParallel(func(pb *testing.PB) {
		var i = 0
		for pb.Next() {
			index := i % benchcount
			i++
			mc.SetWithTTL(benchkeys[index], 1, 1, time.Hour)
		}
	})
}

func BenchmarkRistretto_Get(b *testing.B) {
	var mc, _ = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	for i := 0; i < benchcount; i++ {
		mc.SetWithTTL(benchkeys[i%benchcount], 1, 1, time.Hour)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var i = 0
		for pb.Next() {
			index := i % benchcount
			i++
			mc.Get(benchkeys[index])
		}
	})
}

func BenchmarkRistretto_SetAndGet(b *testing.B) {
	var mc, _ = ristretto.NewCache(&ristretto.Config{
		NumCounters: 10000 * 128 * 10, // number of keys to track frequency of (10M).
		MaxCost:     1 << 30,          // maximum cost of cache (1GB).
		BufferItems: 64,               // number of keys per Get buffer.
	})
	for i := 0; i < benchcount; i++ {
		mc.SetWithTTL(benchkeys[i%benchcount], 1, 1, time.Hour)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var i = 0
		for pb.Next() {
			index := i % benchcount
			i++
			if index&7 == 0 {
				mc.SetWithTTL(benchkeys[index], 1, 1, time.Hour)
			} else {
				mc.Get(benchkeys[index])
			}
		}
	})
}

func BenchmarkOtter_Set(b *testing.B) {
	var mc, _ = otter.MustBuilder[string, int](10000 * 128).Build()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			index := i % benchcount
			i++
			mc.SetWithTTL(benchkeys[index], 1, time.Hour)
		}
	})
}

func BenchmarkOtter_Get(b *testing.B) {
	mc, _ := otter.MustBuilder[string, int](10000 * 128).Build()
	for i := 0; i < benchcount; i++ {
		mc.SetWithTTL(benchkeys[i%benchcount], 1, time.Hour)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			index := i % benchcount
			i++
			if index&7 == 0 {
				mc.SetWithTTL(benchkeys[index], 1, time.Hour)
			} else {
				mc.Get(benchkeys[index])
			}
		}
	})
}

func BenchmarkOtter_SetAndGet(b *testing.B) {
	mc, _ := otter.MustBuilder[string, int](10000 * 128).Build()
	for i := 0; i < benchcount; i++ {
		mc.SetWithTTL(benchkeys[i%benchcount], 1, time.Hour)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			index := i % benchcount
			i++
			if index&7 == 0 {
				mc.SetWithTTL(benchkeys[index], 1, time.Hour)
			} else {
				mc.Get(benchkeys[index])
			}
		}
	})
}
