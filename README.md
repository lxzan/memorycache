# memorycache

[![Build Status][1]][2] [![codecov][3]][4]

[1]: https://github.com/lxzan/memorycache/workflows/Go%20Test/badge.svg?branch=main

[2]: https://github.com/lxzan/memorycache/actions?query=branch%3Amain

[3]: https://codecov.io/gh/lxzan/memorycache/graph/badge.svg?token=OHD6918OPT

[4]: https://codecov.io/gh/lxzan/memorycache

### Description
Minimalist in-memory KV storage, powered by hashmap and minimal heap, without optimizations for GC.
It has O(1) read efficiency, O(logN) write efficiency.
Cache deprecation policy: the set method cleans up overflowed keys; the cycle cleans up expired keys.

### Principle
- Storage Data Limit: Limited by maximum capacity
- Expiration Time: Supported
- Cache Elimination Policy: LRU-Like, Set method and Cycle Cleanup
- GC Optimization: None
- Persistent: None
- Locking Mechanism: Slicing + Mutual Exclusion Locking

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
- 1,000,000 elements
```
go test -benchmem -run=^$ -bench . github.com/lxzan/memorycache/benchmark
goos: darwin
goarch: arm64
pkg: github.com/lxzan/memorycache/benchmark
BenchmarkMemoryCache_Set-8   	 7038808	       153.9 ns/op	      26 B/op	       0 allocs/op
BenchmarkMemoryCache_Get-8   	22969712	        50.92 ns/op	       0 B/op	       0 allocs/op
BenchmarkRistretto_Set-8     	13417420	       242.9 ns/op	     138 B/op	       2 allocs/op
BenchmarkRistretto_Get-8     	15895714	        75.81 ns/op	      18 B/op	       1 allocs/op
PASS
ok  	github.com/lxzan/memorycache/benchmark	10.849s
```