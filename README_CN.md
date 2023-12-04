<div align="center">
    <h1>MemoryCache</h1>
    <img src="assets/logo.png" alt="logo" width="300px">
    <h5>To the time to life, rather than to life in time.</h5>
</div>

[![Build Status][1]][2] [![codecov][3]][4]

[1]: https://github.com/lxzan/memorycache/workflows/Go%20Test/badge.svg?branch=main
[2]: https://github.com/lxzan/memorycache/actions?query=branch%3Amain
[3]: https://codecov.io/gh/lxzan/memorycache/graph/badge.svg?token=OHD6918OPT
[4]: https://codecov.io/gh/lxzan/memorycache

### 简介：

极简的内存键值（KV）存储系统，其核心由哈希表(HashMap) 和最小四叉堆(Minimal Quad Heap) 构成，没有进行垃圾回收（GC）优化。

**缓存淘汰策略：**

1. Set 方法用于清理溢出的键值对
2. 周期清理过期的键值对

### 原则：

1. 存储数据限制：受最大容量限制
2. 过期时间：支持
3. 缓存淘汰策略：类似 LRU
4. GC 优化：无
5. 持久化：无
6. 锁定机制：分片和互斥锁

### 优势：

1. 简单易用
2. 无需第三方依赖
3. 高性能
4. 内存占用低
5. 使用四叉堆维护过期时间, 有效降低树高度, 提高插入性能

### 方法：

-   [x] **Set** : 设置键值对及其过期时间。如果键已存在，将更新其值和过期时间。
-   [x] **SetWithCallback** : 与 Set 类似，但可指定回调函数。
-   [x] **Get** : 根据键获取值。如果键不存在，第二个返回值为 false。
-   [x] **GetWithTTL** : 根据键获取值，如果键不存在，第二个返回值为 false。在返回值时，该方法将刷新过期时间。
-   [x] **Delete** : 根据键删除键值对。
-   [x] **GetOrCreate** : 根据键获取值。如果键不存在，将创建该值。
-   [x] **GetOrCreateWithCallback** : 根据键获取值。如果键不存在，将创建该值，并可调用回调函数。

### 使用

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

### 基准测试

-   1,000,000 元素

```
go test -benchmem -run=^$ -bench . github.com/lxzan/memorycache/benchmark
goos: linux
goarch: amd64
pkg: github.com/lxzan/memorycache/benchmark
cpu: AMD Ryzen 5 PRO 4650G with Radeon Graphics
BenchmarkMemoryCache_Set-12             22848898                62.83 ns/op            8 B/op          0 allocs/op
BenchmarkMemoryCache_Get-12             47904933                30.94 ns/op            0 B/op          0 allocs/op
BenchmarkMemoryCache_SetAndGet-12       48951848                34.41 ns/op            0 B/op          0 allocs/op
BenchmarkRistretto_Set-12               12992732               139.3 ns/op           118 B/op          2 allocs/op
BenchmarkRistretto_Get-12               27832851                45.11 ns/op           16 B/op          1 allocs/op
BenchmarkRistretto_SetAndGet-12         12232522               102.9 ns/op            32 B/op          1 allocs/op
PASS
ok      github.com/lxzan/memorycache/benchmark  31.772s
```
