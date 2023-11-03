package memorycache

import "time"

// 回调函数触发原因
type Reason uint8

const (
	ReasonExpired  = Reason(0)
	ReasonOverflow = Reason(1)
	ReasonDeleted  = Reason(2)
)

type CallbackFunc func(ele *Element, reason Reason)

var emptyCallback CallbackFunc = func(ele *Element, reason Reason) {}

type Element struct {
	// 索引
	index int

	// 回调函数
	cb CallbackFunc

	// 键
	Key string

	// 值
	Value any

	// 过期时间, 毫秒
	ExpireAt int64
}

func (c *Element) expired(now int64) bool {
	return now > c.ExpireAt
}

type config struct {
	MinInterval, MaxInterval time.Duration // 检查周期
	BucketNum                int           // 存储桶数量
	MaxKeysDeleted           int           // 每次检查至多删除key的数量(单个存储桶)
	InitialSize              int           // 初始化大小(单个存储桶)
	MaxCapacity              int           // 最大容量(单个存储桶)
}
