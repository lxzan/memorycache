package memorycache

import "time"

const (
	DefaultSegment          = 16
	DefaultTTLCheckInterval = 30 * time.Second
)

type (
	Config struct {
		TTLCheckInterval time.Duration // 过期检查周期
		Segment          uint32        // 分片数, segment=2^n, eg: 4, 8, 16...
	}

	Option func(c *Config)
)

func WithSegment(segment uint32) Option {
	return func(c *Config) {
		c.Segment = segment
	}
}

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
