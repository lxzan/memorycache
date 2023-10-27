package types

import "time"

type Element struct {
	// 索引
	Index int

	// 键
	Key string

	// 值
	Value any

	// 过期时间, 毫秒
	ExpireAt int64
}

func (c *Element) Expired(now int64) bool {
	return now > c.ExpireAt
}

type Config struct {
	Interval       time.Duration // 检查周期
	BucketNum      int           // 存储桶数量
	MaxKeysDeleted int           // 每次检查至多删除key的数量(单个存储桶)
	InitialSize    int           // 初始化大小(单个存储桶)
	MaxCapacity    int           // 最大容量(单个存储桶)
}
