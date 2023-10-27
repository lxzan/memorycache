# memorycache

[![Build Status][1]][2] [![codecov][3]][4]

[1]: https://github.com/lxzan/memorycache/workflows/Go%20Test/badge.svg?branch=main

[2]: https://github.com/lxzan/memorycache/actions?query=branch%3Amain

[3]: https://codecov.io/gh/lxzan/memorycache/graph/badge.svg?token=OHD6918OPT

[4]: https://codecov.io/gh/lxzan/memorycache

### Description
Minimalist in-memory KV storage, powered by hashmap and minimal heap, with no special optimizations for GC.
It has O(1) read efficiency, O(logN) write efficiency.
Cache deprecation policy: obsolete or overflowed keys are flushed, with a 30s (default) check.

### Usage
```go
package main

import (
	"fmt"
	"github.com/lxzan/memorycache"
	"time"
)

func main() {
	mc := memorycache.New(
		memorycache.WithBucketNum(16),
		memorycache.WithBucketSize(1000, 100000),
		memorycache.WithInterval(100*time.Millisecond),
	)

	mc.Set("xxx", 1, 500*time.Millisecond)

	val, exist := mc.Get("xxx")
	fmt.Printf("val=%v, exist=%v\n", val, exist)

	time.Sleep(time.Second)

	val, exist = mc.Get("xxx")
	fmt.Printf("val=%v, exist=%v\n", val, exist)
}
```

### Benchmark
- 10,000 elements
- 1,000,000 elements
```
go test -benchmem -run=^$ -bench . github.com/lxzan/memorycache/benchmark
goos: darwin
goarch: arm64
pkg: github.com/lxzan/memorycache/benchmark
BenchmarkSet/10000-8            13830640                87.25 ns/op            0 B/op          0 allocs/op
BenchmarkSet/1000000-8           3615801               326.6 ns/op            58 B/op          0 allocs/op
BenchmarkGet/10000-8            14347058                82.28 ns/op            0 B/op          0 allocs/op
BenchmarkGet/1000000-8           3899768               262.6 ns/op            54 B/op          0 allocs/op
PASS
ok      github.com/lxzan/memorycache/benchmark  13.037s
```