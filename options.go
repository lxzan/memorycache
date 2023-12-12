package memorycache

import (
	"time"

	"github.com/lxzan/memorycache/internal/utils"
)

const (
	defaultBucketNum    = 16
	defaultMinInterval  = 5 * time.Second
	defaultMaxInterval  = 30 * time.Second
	defaultDeleteLimits = 1000
	defaultBucketSize   = 1000
	defaultBucketCap    = 100000
)

type Option func(c *config)

// WithBucketNum 设置存储桶数量
// Setting the number of storage buckets
func WithBucketNum(num int) Option {
	return func(c *config) {
		c.BucketNum = num
	}
}

// WithDeleteLimits 设置每次TTL检查最大删除key数量. (单个存储桶)
// Set the maximum number of keys to be deleted per TTL check (single bucket)
func WithDeleteLimits(num int) Option {
	return func(c *config) {
		c.DeleteLimits = num
	}
}

// WithBucketSize 设置初始化大小和最大容量. 超过最大容量会被清除. (单个存储桶)
// Set the initial size and maximum capacity. Exceeding the maximum capacity will be erased. (Single bucket)
func WithBucketSize(size, cap int) Option {
	return func(c *config) {
		c.BucketSize = size
		c.BucketCap = cap
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

// WithCachedTime 是否开启时间缓存
// Whether to turn on time caching
func WithCachedTime(enabled bool) Option {
	return func(c *config) {
		c.CachedTime = enabled
	}
}

// WithSwissTable 使用swiss table替代runtime map
// Using swiss table instead of runtime map
func WithSwissTable(enabled bool) Option {
	return func(c *config) {
		c.SwissTable = enabled
	}
}

// WithLRU 是否开启LRU缓存驱逐算法. 默认为true
// Whether to enable LRU cache eviction. Default is true
func WithLRU(enabled bool) Option {
	return func(c *config) {
		c.LRU = enabled
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

		if c.DeleteLimits <= 0 {
			c.DeleteLimits = defaultDeleteLimits
		}

		if c.BucketSize <= 0 {
			c.BucketSize = defaultBucketSize
		}

		if c.BucketCap <= 0 {
			c.BucketCap = defaultBucketCap
		}
	}
}

type config struct {
	// 检查周期, 默认30s, 最小检查周期为5s
	// Check period, default 30s, minimum 5s.
	MinInterval, MaxInterval time.Duration

	// 存储桶的初始化大小和最大容量, 默认为1000和100000
	// Initialized bucket size and maximum capacity, defaults to 1000 and 100000.
	BucketSize, BucketCap int

	// 存储桶数量, 默认为16
	// Number of buckets, default is 16
	BucketNum int

	// 每次检查至多删除key的数量(单个存储桶)
	// The number of keys to be deleted per check (for a single bucket).
	DeleteLimits int

	// 是否开启时间缓存, 默认为true
	// Whether to enable time caching, true by default.
	CachedTime bool

	// 是否使用swiss table, 默认为false
	// Whether to use swiss table, false by default.
	SwissTable bool

	// 是否开启LRU缓存驱逐算法. 默认为true
	// Whether to enable LRU cache eviction. Default is true
	LRU bool
}
