package containers

import "github.com/dolthub/swiss"

type Map[K comparable, V any] interface {
	Count() int
	Get(K) (V, bool)
	Put(k K, v V)
	Delete(key K) bool
	Iter(f func(K, V) bool)
}

type HashMap[K comparable, V any] map[K]V

func (c HashMap[K, V]) Count() int {
	return len(c)
}

func (c HashMap[K, V]) Put(k K, v V) {
	c[k] = v
}

func (c HashMap[K, V]) Get(k K) (V, bool) {
	v, ok := c[k]
	return v, ok
}

func (c HashMap[K, V]) Delete(k K) bool {
	delete(c, k)
	return true
}

func (c HashMap[K, V]) Iter(f func(K, V) bool) {
	for k, v := range c {
		if !f(k, v) {
			return
		}
	}
}

func NewMap[K comparable, V any](capacity int, swissTable bool) Map[K, V] {
	if swissTable {
		return swiss.NewMap[K, V](uint32(capacity))
	}
	return make(HashMap[K, V], capacity)
}
