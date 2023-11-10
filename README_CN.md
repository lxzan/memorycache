# MemoryCache

[![Build Status][1]][2] [![codecov][3]][4]

[1]: https://github.com/lxzan/memorycache/workflows/Go%20Test/badge.svg?branch=main
[2]: https://github.com/lxzan/memorycache/actions?query=branch%3Amain
[3]: https://codecov.io/gh/lxzan/memorycache/graph/badge.svg?token=OHD6918OPT
[4]: https://codecov.io/gh/lxzan/memorycache

### 简介：

这段文字介绍了一种内存中的极简键值（KV）存储系统，其核心由 HashMap 和最小化的 Quad Heap 构成，没有进行垃圾回收（GC）优化。

**缓存过期策略：**

1. 使用 Set 方法用于清理溢出的键值对。
2. 主动周期清理过期的键值对。

### 原则：

1. 存储数据限制：受最大容量限制。
2. 过期时间：支持为键值对设置过期时间。
3. 缓存淘汰策略：类似于 LRU，使用 Set 方法和周期清理来维护缓存的新鲜度。
4. GC 优化：未进行任何优化。
5. 持久性：不支持数据持久化到磁盘。
6. 锁定机制：采用切片和互斥锁确保线程安全。

### 优势：

1. 简单易用。
2. 无需第三方依赖。
3. 高性能。
4. 内存占用低。
5. 通过 Quad Heap 将 LRU 算法的时间复杂度从 O(n) 降至 O(logn)。

### 方法：

-   [x] **Set** : 设置键值对及其过期时间。如果键已存在，将更新其值和过期时间。
-   [x] **SetWithCallback** : 与 Set 类似，但可指定回调函数。
-   [x] **Get** : 根据键获取值。如果键不存在，第二个返回值为 false。
-   [x] **GetWithTTL** : 根据键获取值，如果键不存在，第二个返回值为 false。在返回值时，该方法将刷新过期时间。
-   [x] **Delete** : 根据键删除键值对。
-   [x] **GetOrCreate** : 根据键获取值。如果键不存在，将创建该值。
-   [x] **GetOrCreateWithCallback** : 根据键获取值。如果键不存在，将创建该值，并可调用回调函数。

### 举例

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
	defer mc.Stop() // Stop memorycache.

	mc.Set("xxx", 1, 10*time.Second)

	val, exist := mc.Get("xxx")
	fmt.Printf("val=%v, exist=%v\n", val, exist)

	time.Sleep(32 * time.Second)

	val, exist = mc.Get("xxx")
	fmt.Printf("val=%v, exist=%v\n", val, exist)
}
```

### 基准测试

-   1,000,000 元素

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
