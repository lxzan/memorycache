# memorycache

[![Build Status][1]][2] [![codecov][3]][4]

[1]: https://github.com/lxzan/memorycache/workflows/Go%20Test/badge.svg?branch=main

[2]: https://github.com/lxzan/memorycache/actions?query=branch%3Amain

[3]: https://codecov.io/gh/lxzan/memorycache/graph/badge.svg?token=OHD6918OPT

[4]: https://codecov.io/gh/lxzan/memorycache

### Description

Minimalist in-memory KV storage, powered by hashmap and minimal quad heap, without optimizations for GC.
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
		memorycache.WithBucketNum(128),
		memorycache.WithBucketSize(1000, 10000),
		memorycache.WithInterval(5*time.Second, 30*time.Second),
	)

	mc.Set("xxx", 1, 10*time.Second)

	val, exist := mc.Get("xxx")
	fmt.Printf("val=%v, exist=%v\n", val, exist)

	time.Sleep(32 * time.Second)

	val, exist = mc.Get("xxx")
	fmt.Printf("val=%v, exist=%v\n", val, exist)
}
```

### Benchmark

- 1,000,000 elements

```
goos: windows
goarch: amd64
pkg: github.com/lxzan/memorycache/benchmark
cpu: AMD Ryzen 5 PRO 4650G with Radeon Graphics
BenchmarkMemoryCache_Set-12     14058852                73.00 ns/op           14 B/op          0 allocs/op
BenchmarkMemoryCache_Get-12     30767100                34.70 ns/op            0 B/op          0 allocs/op
BenchmarkRistretto_Set-12       15583969               218.4 ns/op           114 B/op          2 allocs/op
BenchmarkRistretto_Get-12       27272788                42.05 ns/op           16 B/op          1 allocs/op
PASS
ok      github.com/lxzan/memorycache/benchmark  17.232s
```
