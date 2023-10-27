package utils

const (
	offset32 = 2166136261
	prime32  = 16777619
)

// Fnv32 returns a new 32-bit FNV-1 hash.Hash.
func Fnv32(s string) uint32 {
	var hash uint32 = offset32
	for _, c := range s {
		hash *= prime32
		hash ^= uint32(c)
	}
	return hash
}
