package memorycache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithBucketNum(t *testing.T) {
	var as = assert.New(t)
	{
		var mc = New(WithBucketNum(3))
		as.Equal(mc.config.BucketNum, 4)
	}
	{
		var mc = New(WithBucketNum(0))
		as.Equal(mc.config.BucketNum, defaultBucketNum)
	}
}

func TestWithInterval(t *testing.T) {
	var as = assert.New(t)
	{
		var mc = New()
		as.Equal(mc.config.MinInterval, defaultMinInterval)
		as.Equal(mc.config.MaxInterval, defaultMaxInterval)
	}
	{
		var mc = New(WithInterval(time.Second, 2*time.Second))
		as.Equal(mc.config.MinInterval, time.Second)
		as.Equal(mc.config.MaxInterval, 2*time.Second)
	}
}

func TestWithCapacity(t *testing.T) {
	var as = assert.New(t)
	{
		var mc = New(WithBucketSize(0, 0))
		as.Equal(mc.config.InitialSize, defaultInitialSize)
		as.Equal(mc.config.MaxCapacity, defaultMaxCapacity)
	}
	{
		var mc = New(WithBucketSize(100, 1000))
		as.Equal(mc.config.InitialSize, 100)
		as.Equal(mc.config.MaxCapacity, 1000)
	}
}

func TestWithMaxKeysDeleted(t *testing.T) {
	var as = assert.New(t)
	{
		var mc = New(WithMaxKeysDeleted(0))
		as.Equal(mc.config.MaxKeysDeleted, defaultMaxKeysDeleted)
	}
	{
		var mc = New(WithMaxKeysDeleted(10))
		as.Equal(mc.config.MaxKeysDeleted, 10)
	}
}
