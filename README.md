<div align="center">
    <h1>MemoryCache</h1>
    <img src="assets/logo.png" alt="logo" width="300px">
</div>


[中文](README_CN.md)

[![Build Status][1]][2] [![codecov][3]][4]

[1]: https://github.com/lxzan/memorycache/workflows/Go%20Test/badge.svg?branch=main
[2]: https://github.com/lxzan/memorycache/actions?query=branch%3Amain
[3]: https://codecov.io/gh/lxzan/memorycache/graph/badge.svg?token=OHD6918OPT
[4]: https://codecov.io/gh/lxzan/memorycache

### Description

Minimalist in-memory KV storage, powered by `HashMap` and `Minimal Quad Heap`, without optimizations for GC.

**Cache Elimination Policy:**

1. Set method cleans up overflowed keys
2. Active cycle cleans up expired keys

### Principle

-   Storage Data Limit: Limited by maximum capacity
-   Expiration Time: Supported
-   Cache Elimination Policy: LRU-Like
-   GC Optimization: None
-   Persistent: None
-   Locking Mechanism: Slicing + Mutual Exclusion Locking

### Advantage

-   Simple and easy to use
-   No third-party dependencies
-   High performance
-   Low memory usage
-   Use Quad Heap to reduce height and increase write speeds

### Methods

-   [x] **Set** : Set key-value pair with expiring time. If the key already exists, the value will be updated. Also the expiration time will be updated.
-   [x] **SetWithCallback** : Set key-value pair with expiring time and callback function. If the key already exists, the value will be updated. Also the expiration time will be updated.
-   [x] **Get** : Get value by key. If the key does not exist, the second return value will be false.
-   [x] **GetWithTTL** : Get value by key. If the key does not exist, the second return value will be false. When return value, method will refresh the expiration time.
-   [x] **Delete** : Delete key-value pair by key.
-   [x] **GetOrCreate** : Get value by key. If the key does not exist, the value will be created.
-   [x] **GetOrCreateWithCallback** : Get value by key. If the key does not exist, the value will be created. Also the callback function will be called.

### Example

```go
package main

import (
	"fmt"
	"github.com/lxzan/memorycache"
	"time"
)

func main() {
	mc := memorycache.New(
		memorycache.WithBucketNum(128),  // Bucket number, recommended to be a prime number.
		memorycache.WithBucketSize(1000, 10000), // Bucket size, initial size and maximum capacity.
		memorycache.WithInterval(5*time.Second, 30*time.Second), // Active cycle cleanup interval and expiration time.
	)
	defer mc.Stop()

	mc.Set("xxx", 1, 10*time.Second)

	val, exist := mc.Get("xxx")
	fmt.Printf("val=%v, exist=%v\n", val, exist)

	time.Sleep(32 * time.Second)

	val, exist = mc.Get("xxx")
	fmt.Printf("val=%v, exist=%v\n", val, exist)
}
```

### Benchmark

-   1,000,000 elements

```
goos: linux
goarch: amd64
pkg: github.com/lxzan/memorycache/benchmark
cpu: AMD EPYC 7763 64-Core Processor
BenchmarkMemoryCache_Set-4      11106261    100.6 ns/op	      18 B/op	       0 allocs/op
BenchmarkMemoryCache_Get-4      635988      77.30 ns/op	       0 B/op	       0 allocs/op
BenchmarkRistretto_Set-4        7933663     491.8 ns/op	     170 B/op	       2 allocs/op
BenchmarkRistretto_Get-4        11085688    98.92 ns/op	      18 B/op	       1 allocs/op
PASS
```
