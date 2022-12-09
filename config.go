package memorycache

import "time"

const (
	DefaultSegment          = 16
	DefaultTTLCheckInterval = 30 * time.Second
)

type (
	Config struct {
		// 过期时间检查周期
		// expiration time inspection cycle
		TTLCheckInterval time.Duration //

		// 分片数, segment=2^n, eg: 4, 8, 16...
		// number of hashmap pieces, eg: 4, 8, 16...
		Segment uint32
	}

	Option func(c *Config)
)

// WithSegment
// Set Segment
func WithSegment(segment uint32) Option {
	return func(c *Config) {
		c.Segment = segment
	}
}

// WithTTLCheckInterval
// Set TTLCheckInterval
func WithTTLCheckInterval(TTLCheckInterval time.Duration) Option {
	return func(c *Config) {
		c.TTLCheckInterval = TTLCheckInterval
	}
}

func withInitialize() Option {
	return func(c *Config) {
		if c.Segment <= 0 {
			c.Segment = DefaultSegment
		} else {
			var segment = uint32(1)
			for segment < c.Segment {
				segment *= 2
			}
			c.Segment = segment
		}

		if c.TTLCheckInterval <= 0 {
			c.TTLCheckInterval = DefaultTTLCheckInterval
		}
	}
}
