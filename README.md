<div align="center">
    <h1>MemoryCache</h1>
    <img src="assets/logo.png" alt="logo" width="300px">
    <h5>To the time to life, rather than to life in time.</h5>
</div>


[中文](README_CN.md)

[![Build Status][1]][2] [![codecov][3]][4]

[1]: https://github.com/lxzan/memorycache/workflows/Go%20Test/badge.svg?branch=main

[2]: https://github.com/lxzan/memorycache/actions?query=branch%3Amain

[3]: https://codecov.io/gh/lxzan/memorycache/graph/badge.svg?token=OHD6918OPT

[4]: https://codecov.io/gh/lxzan/memorycache

### Description

Minimalist in-memory KV storage, powered by `HashMap` and `Minimal Quad Heap`.

**Cache Elimination Policy:**

- Set method cleans up overflowed keys
- Active cycle cleans up expired keys

### Principle

- Storage Data Limit: Limited by maximum capacity
- Expiration Time: Supported
- Cache Eviction Policy: LRU
- Persistent: None
- Locking Mechanism: Slicing + Mutual Exclusion Locking
- HashMap, Heap and LinkedList (excluding user KVs) implemented in pointerless technology

### Advantage

- Simple and easy to use
- High performance
- Low memory usage
- Use quadruple heap to maintain the expiration time, effectively reduce the height of the tree, and improve the
  insertion performance

### Methods

-   [x] **Set** : Set key-value pair with expiring time. If the key already exists, the value will be updated. Also the
    expiration time will be updated.
-   [x] **SetWithCallback** : Set key-value pair with expiring time and callback function. If the key already exists,
    the value will be updated. Also the expiration time will be updated.
-   [x] **Get** : Get value by key. If the key does not exist, the second return value will be false.
-   [x] **GetWithTTL** : Get value by key. If the key does not exist, the second return value will be false. When return
    value, method will refresh the expiration time.
-   [x] **Delete** : Delete key-value pair by key.
-   [x] **GetOrCreate** : Get value by key. If the key does not exist, the value will be created.
-   [x] **GetOrCreateWithCallback** : Get value by key. If the key does not exist, the value will be created. Also the
    callback function will be called.

### Example

```go
package main

import (
	"fmt"
	"github.com/lxzan/memorycache"
	"time"
)

func main() {
	mc := memorycache.New[string, any](
		// Set the number of storage buckets, y=pow(2,x)
		memorycache.WithBucketNum(128),

		// Set bucket size, initial size and maximum capacity (single bucket)
		memorycache.WithBucketSize(1000, 10000),

		// Set the expiration time check period. 
		// If the number of expired elements is small, take the maximum value, otherwise take the minimum value.
		memorycache.WithInterval(5*time.Second, 30*time.Second),
	)

	mc.SetWithCallback("xxx", 1, time.Second, func(element *memorycache.Element[string, any], reason memorycache.Reason) {
		fmt.Printf("callback: key=%s, reason=%v\n", element.Key, reason)
	})

	val, exist := mc.Get("xxx")
	fmt.Printf("val=%v, exist=%v\n", val, exist)

	time.Sleep(2 * time.Second)

	val, exist = mc.Get("xxx")
	fmt.Printf("val=%v, exist=%v\n", val, exist)
}

```

### Benchmark

- 1,000,000 elements

```
goos: linux
goarch: amd64
pkg: github.com/lxzan/memorycache/benchmark
cpu: AMD Ryzen 5 PRO 4650G with Radeon Graphics
BenchmarkMemoryCache_Set-8              16107153                74.85 ns/op           15 B/op          0 allocs/op
BenchmarkMemoryCache_Get-8              28859542                42.34 ns/op            0 B/op          0 allocs/op
BenchmarkMemoryCache_SetAndGet-8        27317874                63.02 ns/op            0 B/op          0 allocs/op
BenchmarkRistretto_Set-8                13343023               272.6 ns/op           120 B/op          2 allocs/op
BenchmarkRistretto_Get-8                19799044                55.06 ns/op           17 B/op          1 allocs/op
BenchmarkRistretto_SetAndGet-8          11212923               119.6 ns/op            30 B/op          1 allocs/op
BenchmarkTheine_Set-8                    3775975               322.5 ns/op            30 B/op          0 allocs/op
BenchmarkTheine_Get-8                   21579301                54.94 ns/op            0 B/op          0 allocs/op
BenchmarkTheine_SetAndGet-8              6265330               224.6 ns/op             0 B/op          0 allocs/op
PASS
ok      github.com/lxzan/memorycache/benchmark  53.498s
```
