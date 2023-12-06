package memorycache

import "time"

// 回调函数触发原因
type Reason uint8

const (
	ReasonExpired = Reason(0)
	ReasonEvicted = Reason(1)
	ReasonDeleted = Reason(2)
)

type CallbackFunc[T any] func(element T, reason Reason)

type Element[K comparable, V any] struct {
	// 前后指针
	prev, next *Element[K, V]

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

type config struct {
	MinInterval, MaxInterval time.Duration // 检查周期
	BucketNum                int           // 存储桶数量
	MaxKeysDeleted           int           // 每次检查至多删除key的数量(单个存储桶)
	InitialSize              int           // 初始化大小(单个存储桶)
	MaxCapacity              int           // 最大容量(单个存储桶)
	TimeCacheEnabled         bool          // 是否开启时间缓存
}
