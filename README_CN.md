<div align="center">
    <h1>MemoryCache</h1>
    <img src="assets/logo.png" alt="logo" width="300px">
    <h5>To the time to life, rather than to life in time.</h5>
</div>

[English](README.md)

[![Build Status][1]][2] [![codecov][3]][4]

[1]: https://github.com/lxzan/memorycache/workflows/Go%20Test/badge.svg?branch=main
[2]: https://github.com/lxzan/memorycache/actions?query=branch%3Amain
[3]: https://codecov.io/gh/lxzan/memorycache/graph/badge.svg?token=OHD6918OPT
[4]: https://codecov.io/gh/lxzan/memorycache

### 简介

极简的内存键值存储，由 `HashMap` 和最小四叉堆（Minimal Quad Heap）驱动。

**缓存驱逐策略：**

- `Set` 方法在容量溢出时驱逐键值对
- 周期性清理过期键值对

### 设计

- 存储数据限制：受最大容量限制
- 过期时间：支持
- 缓存驱逐策略：LRU
- 持久化：无
- 锁定机制：分桶 + 互斥锁
- GC 优化：无指针技术实现的哈希表、堆和链表（不包括用户 KV）

### 优势

- 简单易用
- 高性能
- 内存占用低
- 使用四叉堆维护过期时间。相比二叉堆，可有效降低堆的高度，提高插入性能。

### 方法

- [x] **Set** : 设置键值对及其过期时间。如果键已存在，将更新其值和过期时间。`exp<=0` 表示永不过期。
- [x] **SetWithCallback** : 与 `Set` 类似，但可指定回调函数。容量溢出或过期时触发回调。
- [x] **Get** : 根据键获取值。如果键不存在，第二个返回值为 false。
- [x] **GetTTL** : 获取键的过期时间。
- [x] **GetWithTTL** : 根据键获取值。如果键存在，刷新其过期时间。
- [x] **UpdateTTL** : 更新键的过期时间。`d<=0` 表示永不过期。
- [x] **Delete** : 根据键删除键值对。
- [x] **GetOrCreate** : 根据键获取值。如果键不存在，将创建该值。
- [x] **GetOrCreateWithCallback** : 根据键获取值。如果键不存在，将创建该值，并可指定回调函数。
- [x] **Clear** : 清空所有缓存。对存活非过期元素触发 `ReasonCleared` 回调。
- [x] **Range** : 遍历缓存中的所有键值对。
- [x] **Len** : 获取当前缓存元素数量（可能包含已过期但未清除的元素）。
- [x] **Stop** : 停止后台协程，释放资源。

### 配置项

| 选项                        | 默认值       | 说明                                     |
| --------------------------- | ------------ | ---------------------------------------- |
| `WithBucketNum(num)`        | 16           | 存储桶数量（自动取 2 的幂次）            |
| `WithBucketSize(size, cap)` | 1000, 100000 | 单桶初始大小和最大容量                   |
| `WithInterval(min, max)`    | 5s, 30s      | 自适应 TTL 检查周期                      |
| `WithDeleteLimits(num)`     | 1000         | 每次 TTL 检查单桶最大删除数              |
| `WithCachedTime(enabled)`   | true         | 开启时间缓存，减少 `time.Now()` 调用开销 |
| `WithSwissTable(enabled)`   | false        | 使用 SwissTable 替代 Go runtime map      |

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

		// 设置单个存储桶的初始容量和最大容量
		memorycache.WithBucketSize(1000, 10000),

		// 设置过期时间检查周期。如果过期元素较少, 取最大值, 反之取最小值。
		memorycache.WithInterval(5*time.Second, 30*time.Second),
	)
	defer mc.Stop() // 停止后台协程, 释放资源

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

- 预填充 1,000,000 个元素。benchtime=3s，并发执行，10 CPU 核心。

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
