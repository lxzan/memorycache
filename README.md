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

**Cache Eviction Policy:**

- `Set` evicts LRU keys on capacity overflow
- Periodic cycle evicts expired keys

### Design

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
- Uses a 4-ary min-heap to maintain expiration times. Compared to binary heaps, this reduces tree height and improves insertion performance.

### Methods

- [x] **Set** : Set key-value pair with expiring time. If the key already exists, the value will be updated. Also the
      expiration time will be updated. `exp<=0` means never expire.
- [x] **SetWithCallback** : Similar to `Set`, but with a callback function. The callback is triggered on capacity overflow or expiration.
- [x] **Get** : Get value by key. If the key does not exist, the second return value will be false.
- [x] **GetTTL** : Get the expiration time of the key.
- [x] **GetWithTTL** : Get value by key. If the key exists, refreshes the expiration time.
- [x] **UpdateTTL** : Update the expiration time of the key. `d<=0` means never expire.
- [x] **Delete** : Delete key-value pair by key.
- [x] **GetOrCreate** : Get value by key. If the key does not exist, the value will be created.
- [x] **GetOrCreateWithCallback** : Get value by key. If the key does not exist, the value will be created. Also the
      callback function will be called.
- [x] **Clear** : Clear all caches. Triggers `ReasonCleared` callback for alive non-expired elements.
- [x] **Range** : Iterate all key-value pairs in the cache.
- [x] **Len** : Get the current number of cached elements (may include expired but un-cleaned elements).
- [x] **Stop** : Stop the background goroutines and release resources.

### Options

| Option                      | Default      | Description                                            |
| --------------------------- | ------------ | ------------------------------------------------------ |
| `WithBucketNum(num)`        | 16           | Number of storage buckets (auto-rounded to power of 2) |
| `WithBucketSize(size, cap)` | 1000, 100000 | Initial size and max capacity per bucket               |
| `WithInterval(min, max)`    | 5s, 30s      | Adaptive TTL check period                              |
| `WithDeleteLimits(num)`     | 1000         | Max keys deleted per TTL check per bucket              |
| `WithCachedTime(enabled)`   | true         | Enable time caching for lower `time.Now()` overhead    |
| `WithSwissTable(enabled)`   | false        | Use SwissTable instead of Go runtime map               |

### Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/lxzan/memorycache"
)

func main() {
	mc := memorycache.New[string, any](
		// Set the number of storage buckets, y=pow(2,x)
		memorycache.WithBucketNum(128),

		// Set bucket size, initial size and maximum capacity (per bucket)
		memorycache.WithBucketSize(1000, 10000),

		// Set the expiration time check period.
		// If the number of expired elements is small, take the maximum value, otherwise take the minimum value.
		memorycache.WithInterval(5*time.Second, 30*time.Second),
	)
	defer mc.Stop() // Stop background goroutines and release resources

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

- Pre-populated with 1,000,000 elements. benchtime=3s, parallel, 10 CPU cores.

```
goos: darwin
goarch: arm64
pkg: github.com/lxzan/memorycache/benchmark
cpu: Apple M1 Max
BenchmarkMemoryCache_Set-10           90382902        41.77 ns/op        4 B/op        0 allocs/op
BenchmarkMemoryCache_Get-10           79059558        50.51 ns/op        0 B/op        0 allocs/op
BenchmarkMemoryCache_SetAndGet-10     78810615        48.48 ns/op        0 B/op        0 allocs/op
BenchmarkMemoryCache_Delete-10       153390056        23.76 ns/op        0 B/op        0 allocs/op
BenchmarkMemoryCache_UpdateTTL-10    100000000        32.73 ns/op        0 B/op        0 allocs/op
BenchmarkRistretto_Set-10             68245581       673.1 ns/op        112 B/op        2 allocs/op
BenchmarkRistretto_Get-10             86053975        45.96 ns/op         17 B/op        1 allocs/op
BenchmarkRistretto_SetAndGet-10       27956532       110.6 ns/op         31 B/op        1 allocs/op
BenchmarkTheine_Set-10                 9754111       381.7 ns/op         22 B/op        0 allocs/op
BenchmarkTheine_Get-10                32876373       109.1 ns/op          0 B/op        0 allocs/op
BenchmarkTheine_SetAndGet-10          23014824       172.4 ns/op          0 B/op        0 allocs/op
PASS
ok      github.com/lxzan/memorycache/benchmark  103.750s
```
