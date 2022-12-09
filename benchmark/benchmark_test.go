package benchmark

import (
	"github.com/lxzan/memorycache"
	"github.com/lxzan/memorycache/internal/utils"
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

func BenchmarkSet(b *testing.B) {
	var f = func(n, count int) {
		var mc = memorycache.New(memorycache.WithBucketNum(16))
		for i := 0; i < n; i++ {
			var key = benchkeys[i%count]
			mc.Set(key, 1, time.Hour)
		}
	}

	b.Run("10000", func(b *testing.B) { f(b.N, 10000) })
	b.Run("1000000", func(b *testing.B) { f(b.N, 1000000) })
}

func BenchmarkGet(b *testing.B) {
	var f = func(n, count int) {
		var mc = memorycache.New(memorycache.WithBucketNum(16))
		for i := 0; i < count; i++ {
			var key = benchkeys[i]
			mc.Set(key, 1, time.Hour)
		}

		for i := 0; i < n; i++ {
			var key = benchkeys[i%count]
			mc.Get(key)
		}
	}

	b.Run("10000", func(b *testing.B) { f(b.N, 10000) })
	b.Run("1000000", func(b *testing.B) { f(b.N, 1000000) })
}
