package memorycache

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestWithBucketNum(t *testing.T) {
	var as = assert.New(t)
	{
		var mc = New(WithBucketNum(3))
		as.Equal(mc.config.BucketNum, uint32(4))
	}
	{
		var mc = New(WithBucketNum(0))
		as.Equal(mc.config.BucketNum, uint32(defaultBucketNum))
	}
}

func TestWithInterval(t *testing.T) {
	var as = assert.New(t)
	{
		var mc = New(WithInterval(0))
		as.Equal(mc.config.Interval, defaultInterval)
	}
	{
		var mc = New(WithInterval(time.Second))
		as.Equal(mc.config.Interval, time.Second)
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
