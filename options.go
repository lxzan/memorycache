package memorycache

import (
	"github.com/lxzan/memorycache/internal/types"
	"github.com/lxzan/memorycache/internal/utils"
	"time"
)

const (
	defaultBucketNum      = 16
	defaultInterval       = 30 * time.Second
	defaultMaxKeysDeleted = 1000
	defaultInitialSize    = 1000
	defaultMaxCapacity    = 100000
)

type Option func(c *types.Config)

// WithBucketNum 设置存储桶数量
func WithBucketNum(num uint32) Option {
	return func(c *types.Config) {
		c.BucketNum = num
	}
}

// WithMaxKeysDeleted (单个存储桶)设置每次TTL检查最大删除key数量
func WithMaxKeysDeleted(num int) Option {
	return func(c *types.Config) {
		c.MaxKeysDeleted = num
	}
}

// WithBucketSize (单个存储桶)设置初始化大小, 最大容量. 超过最大容量会被定期清除.
func WithBucketSize(size, cap int) Option {
	return func(c *types.Config) {
		c.InitialSize = size
		c.MaxCapacity = cap
	}
}

// WithInterval 设置TTL检查周期
func WithInterval(d time.Duration) Option {
	return func(c *types.Config) {
		c.Interval = d
	}
}

func withInitialize() Option {
	return func(c *types.Config) {
		if c.BucketNum <= 0 {
			c.BucketNum = defaultBucketNum
		}
		c.BucketNum = utils.ToBinaryNumber(c.BucketNum)

		if c.Interval <= 0 {
			c.Interval = defaultInterval
		}

		if c.MaxKeysDeleted <= 0 {
			c.MaxKeysDeleted = defaultMaxKeysDeleted
		}

		if c.InitialSize <= 0 {
			c.InitialSize = defaultInitialSize
		}

		if c.MaxCapacity <= 0 {
			c.MaxCapacity = defaultMaxCapacity
		}
	}
}
