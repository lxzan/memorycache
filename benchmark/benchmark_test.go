package benchmark

import (
	"github.com/lxzan/memorycache"
	"github.com/lxzan/memorycache/internal/utils"
	"sync"
	"testing"
	"time"
)

const benchcount = 10000

var benchkeys = make([]string, 0, benchcount)

func init() {
	for i := 0; i < benchcount; i++ {
		benchkeys = append(benchkeys, utils.Alphabet.Generate(16))
	}
}

func BenchmarkSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var mc = memorycache.New(memorycache.WithSegment(16))
		var wg = sync.WaitGroup{}
		wg.Add(2)
		go func() {
			var d = utils.Rand.Intn(5)
			for j := 0; j < benchcount/2; j++ {
				mc.Set(benchkeys[j], 1, time.Duration(d)*time.Second)
			}
			wg.Done()
		}()
		go func() {
			var d = utils.Rand.Intn(5)
			for j := benchcount / 2; j < benchcount; j++ {
				mc.Set(benchkeys[j], 1, time.Duration(d)*time.Second)
			}
			wg.Done()
		}()
		wg.Wait()
	}
}

func BenchmarkGet(b *testing.B) {
	var mc = memorycache.New(memorycache.WithSegment(16))
	for j := 0; j < benchcount; j++ {
		mc.Set(benchkeys[j], 1, -1)
	}

	for i := 0; i < b.N; i++ {
		var wg = sync.WaitGroup{}
		wg.Add(2)
		go func() {
			for j := 0; j < benchcount/2; j++ {
				mc.Get(benchkeys[j])
			}
			wg.Done()
		}()
		go func() {
			for j := benchcount / 2; j < benchcount; j++ {
				mc.Get(benchkeys[j])
			}
			wg.Done()
		}()
		wg.Wait()
	}
}
