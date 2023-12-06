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

1. Set 方法清理溢出的键值对
2. 周期清理过期的键值对

### 原则：

1. 存储数据限制：受最大容量限制
2. 过期时间：支持
3. 缓存驱逐策略：LRU
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
	mc := memorycache.New[string, any](
		memorycache.WithBucketNum(128),                          // Bucket number, recommended to be a prime number.
		memorycache.WithBucketSize(1000, 10000),                 // Bucket size, initial size and maximum capacity.
		memorycache.WithInterval(5*time.Second, 30*time.Second), // Active cycle cleanup interval and expiration time.
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

### 基准测试

-   1,000,000 元素

```
go test -benchmem -run=^$ -bench . github.com/lxzan/memorycache/benchmark
goos: linux
goarch: amd64
pkg: github.com/lxzan/memorycache/benchmark
cpu: AMD Ryzen 5 PRO 4650G with Radeon Graphics
BenchmarkMemoryCache_Set-12             18891738               109.5 ns/op            11 B/op          0 allocs/op
BenchmarkMemoryCache_Get-12             21813127                48.21 ns/op            0 B/op          0 allocs/op
BenchmarkMemoryCache_SetAndGet-12       22530026                52.14 ns/op            0 B/op          0 allocs/op
BenchmarkRistretto_Set-12               13786928               140.6 ns/op           116 B/op          2 allocs/op
BenchmarkRistretto_Get-12               26299240                45.87 ns/op           16 B/op          1 allocs/op
BenchmarkRistretto_SetAndGet-12         11360748               103.0 ns/op            27 B/op          1 allocs/op
BenchmarkTheine_Set-12                   3527848               358.2 ns/op            19 B/op          0 allocs/op
BenchmarkTheine_Get-12                  23234760                49.37 ns/op            0 B/op          0 allocs/op
BenchmarkTheine_SetAndGet-12             6755134               176.3 ns/op             0 B/op          0 allocs/op
PASS
ok      github.com/lxzan/memorycache/benchmark  65.498s
```
