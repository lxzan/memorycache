# memorycache

[![Build Status](https://github.com/lxzan/memorycache/workflows/Go%20Test/badge.svg?branch=main)](https://github.com/lxzan/memorycache/actions?query=branch%3Amain)

### Benchmark
- 10,000 elements, 2 threads
```
goos: darwin
goarch: arm64
pkg: github.com/lxzan/memorycache/benchmark
BenchmarkSet
BenchmarkSet-8   	     716	   1709264 ns/op
BenchmarkGet
BenchmarkGet-8   	    2986	    386144 ns/op
PASS
```