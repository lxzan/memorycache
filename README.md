# memorycache

[![Build Status](https://github.com/lxzan/memorycache/workflows/Go%20Test/badge.svg?branch=main)](https://github.com/lxzan/memorycache/actions?query=branch%3Amain)

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
		memorycache.WithSegment(16),
		memorycache.WithTTLCheckInterval(30*time.Second),
	)
	mc.Set("xxx", 1, 500*time.Millisecond)
	time.Sleep(time.Second)
	val, exist := mc.Get("xxx")
	fmt.Printf("val=%v, exist=%v\n", val, exist)
}
```

### Benchmark
- 10,000 elements, 2 threads
```
goos: darwin
goarch: arm64
pkg: github.com/lxzan/memorycache/benchmark
BenchmarkSet
BenchmarkSet-8   	     727	   1589200 ns/op
BenchmarkGet
BenchmarkGet-8   	    2191	    530433 ns/op
PASS
```