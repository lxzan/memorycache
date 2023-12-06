package memorycache

import (
	"time"

	"github.com/lxzan/memorycache/internal/utils"
)

const (
	defaultBucketNum      = 16
	defaultMinInterval    = 5 * time.Second
	defaultMaxInterval    = 30 * time.Second
	defaultMaxKeysDeleted = 1000
	defaultInitialSize    = 1000
	defaultMaxCapacity    = 100000
)

type Option func(c *config)

// WithBucketNum 设置存储桶数量
// Setting the number of storage buckets
func WithBucketNum(num int) Option {
	return func(c *config) {
		c.BucketNum = num
	}
}

// WithMaxKeysDeleted 设置每次TTL检查最大删除key数量. (单个存储桶)
// Set the maximum number of keys to be deleted per TTL check (single bucket)
func WithMaxKeysDeleted(num int) Option {
	return func(c *config) {
		c.MaxKeysDeleted = num
	}
}

// WithBucketSize 设置初始化大小和最大容量. 超过最大容量会被清除. (单个存储桶)
// Set the initial size and maximum capacity. Exceeding the maximum capacity will be erased. (Single bucket)
func WithBucketSize(size, cap int) Option {
	return func(c *config) {
		c.InitialSize = size
		c.MaxCapacity = cap
	}
}

// WithInterval 设置TTL检查周期
// Setting the TTL check period
func WithInterval(min, max time.Duration) Option {
	return func(c *config) {
		c.MinInterval = min
		c.MaxInterval = max
	}
}

// WithTimeCache 是否开启时间缓存
// Whether to turn on time caching
func WithTimeCache(enabled bool) Option {
	return func(c *config) {
		c.TimeCacheEnabled = enabled
	}
}

func withInitialize() Option {
	return func(c *config) {
		if c.BucketNum <= 0 {
			c.BucketNum = defaultBucketNum
		}
		c.BucketNum = utils.ToBinaryNumber(c.BucketNum)

		if c.MinInterval <= 0 {
			c.MinInterval = defaultMinInterval
		}

		if c.MaxInterval <= 0 {
			c.MaxInterval = defaultMaxInterval
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
