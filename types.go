package memorycache

import "github.com/lxzan/dao/deque"

// Reason 回调函数触发原因
type Reason uint8

const (
	ReasonExpired = Reason(0)
	ReasonEvicted = Reason(1)
	ReasonDeleted = Reason(2)
)

type CallbackFunc[T any] func(element T, reason Reason)

type Element[K comparable, V any] struct {
	// 地址
	addr deque.Pointer

	// 索引
	index int

	// 回调函数
	cb CallbackFunc[*Element[K, V]]

	// 键
	Key K

	// 值
	Value V

	// 过期时间, 毫秒
	ExpireAt int64
}

func (c *Element[K, V]) expired(now int64) bool {
	return now > c.ExpireAt
}
