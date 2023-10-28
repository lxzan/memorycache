package benchmark

import (
	"github.com/dgraph-io/ristretto"
	"github.com/lxzan/memorycache"
	"github.com/lxzan/memorycache/internal/utils"
	"sync/atomic"
	"testing"
	"time"
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
	var i = atomic.Int64{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			index := i.Add(1) % benchcount
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

	var i = atomic.Int64{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			index := i.Add(1) % benchcount
			mc.Get(benchkeys[index])
		}
	})
}

func BenchmarkRistretto_Set(b *testing.B) {
	var mc, _ = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	var i = atomic.Int64{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			index := i.Add(1) % benchcount
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

	var i = atomic.Int64{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			index := i.Add(1) % benchcount
			mc.Get(benchkeys[index])
		}
	})
}
