package utils

const (
	prime64  = 1099511628211
	offset64 = 14695981039346656037
)

func Fnv64(s string) uint64 {
	var hash uint64 = offset64
	for _, c := range s {
		hash *= prime64
		hash ^= uint64(c)
	}
	return hash
}
