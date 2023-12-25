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

极简的内存键值（KV）存储系统，其核心由哈希表(HashMap) 和最小四叉堆(Minimal Quad Heap) 构成.

**缓存淘汰策略：**

- Set 方法清理溢出的键值对
- 周期清理过期的键值对

### 原则：

- 存储数据限制：受最大容量限制
- 过期时间：支持
- 缓存驱逐策略：LRU
- 持久化：无
- 锁定机制：分片和互斥锁
- GC 优化：无指针技术实现的哈希表, 最小堆和链表(不包括用户KV)

### 优势：

- 简单易用
- 高性能
- 内存占用低
- 使用四叉堆维护过期时间, 有效降低树高度, 提高插入性能

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
	"time"

	"github.com/lxzan/memorycache"
)

func main() {
	mc := memorycache.New[string, any](
		// 设置存储桶数量, y=pow(2,x)
		memorycache.WithBucketNum(128),

		// 设置单个存储桶的初始化容量和最大容量
		memorycache.WithBucketSize(1000, 10000),

		// 设置过期时间检查周期. 如果过期元素较少, 取最大值, 反之取最小值.
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

### 基准测试

- 1,000,000 元素

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
