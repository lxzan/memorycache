package memorycache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithBucketNum(t *testing.T) {
	var as = assert.New(t)
	{
		var mc = New[string, any](WithBucketNum(3))
		as.Equal(mc.config.BucketNum, 4)
	}
	{
		var mc = New[string, any](WithBucketNum(0))
		as.Equal(mc.config.BucketNum, defaultBucketNum)
	}
}

func TestWithInterval(t *testing.T) {
	var as = assert.New(t)
	{
		var mc = New[string, any]()
		as.Equal(mc.config.MinInterval, defaultMinInterval)
		as.Equal(mc.config.MaxInterval, defaultMaxInterval)
	}
	{
		var mc = New[string, any](WithInterval(time.Second, 2*time.Second))
		as.Equal(mc.config.MinInterval, time.Second)
		as.Equal(mc.config.MaxInterval, 2*time.Second)
	}
}

func TestWithCapacity(t *testing.T) {
	var as = assert.New(t)
	{
		var mc = New[string, any](WithBucketSize(0, 0))
		as.Equal(mc.config.InitialSize, defaultInitialSize)
		as.Equal(mc.config.MaxCapacity, defaultMaxCapacity)
	}
	{
		var mc = New[string, any](WithBucketSize(100, 1000))
		as.Equal(mc.config.InitialSize, 100)
		as.Equal(mc.config.MaxCapacity, 1000)
	}
}

func TestWithMaxKeysDeleted(t *testing.T) {
	var as = assert.New(t)
	{
		var mc = New[string, any](WithMaxKeysDeleted(0))
		as.Equal(mc.config.MaxKeysDeleted, defaultMaxKeysDeleted)
	}
	{
		var mc = New[string, any](WithMaxKeysDeleted(10))
		as.Equal(mc.config.MaxKeysDeleted, 10)
	}
}

func TestWithTimeCache(t *testing.T) {
	var as = assert.New(t)
	{
		var mc = New[string, any]()
		as.Equal(mc.config.TimeCacheEnabled, true)
	}
	{
		var mc = New[string, any](WithTimeCache(false))
		as.Equal(mc.config.TimeCacheEnabled, false)
	}
}
