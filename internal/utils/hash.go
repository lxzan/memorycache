package utils

const (
	prime64  = 1099511628211
	offset64 = 14695981039346656037

	prime32  = 16777619
	offset32 = 2166136261
)

func Fnv64(s string) uint64 {
	var hash uint64 = offset64
	for _, c := range s {
		hash *= prime64
		hash ^= uint64(c)
	}
	return hash
}

func Fnv32(s string) uint32 {
	var hash uint32 = offset32
	for _, c := range s {
		hash *= prime32
		hash ^= uint32(c)
	}
	return hash
}

type Hasher[K comparable] interface {
	Hash(K) uint64
}

// Fnv32Hasher 用于测试哈希冲突的数据集
// [O4XOUsgCQqkVCvLQ wYLAGPVADrDTi7VT]
// [e7p5kjn8U6SDvI5B wbMm2kjYjwkBeqzc]
// [SfZaE3dDLWcYxT6G x12qmBRf3TVb0oZA]
// [d3n5BOTvkYif9o5T x4vw8ToKcrwQ8aYc]
// [eR8LklziA5C9XsSl xXcUr3WtBNJQomaK]
// [E9rj2ySsqr7DZUuU xkWJQCMpAWrIczTY]
// [kAsRa2rcRAXvEgiB xtLwyg9fYRclMpsW]
// [xOWb2UMFqAEML9d5 xxVWZzckpn6LMhW7]
type Fnv32Hasher struct{}

func (c *Fnv32Hasher) Hash(key string) uint64 {
	return uint64(Fnv32(key))
}
