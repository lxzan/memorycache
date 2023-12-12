package memorycache

import (
	"testing"
	"time"

	"github.com/dolthub/swiss"
	"github.com/lxzan/memorycache/internal/containers"

	"github.com/stretchr/testify/assert"
)

func TestWithBucketNum(t *testing.T) {
	var as = assert.New(t)
	{
		var mc = New[string, any](WithBucketNum(3))
		as.Equal(mc.conf.BucketNum, 4)
	}
	{
		var mc = New[string, any](WithBucketNum(0))
		as.Equal(mc.conf.BucketNum, defaultBucketNum)
	}
}

func TestWithInterval(t *testing.T) {
	var as = assert.New(t)
	{
		var mc = New[string, any]()
		as.Equal(mc.conf.MinInterval, defaultMinInterval)
		as.Equal(mc.conf.MaxInterval, defaultMaxInterval)
	}
	{
		var mc = New[string, any](WithInterval(time.Second, 2*time.Second))
		as.Equal(mc.conf.MinInterval, time.Second)
		as.Equal(mc.conf.MaxInterval, 2*time.Second)
	}
}

func TestWithCapacity(t *testing.T) {
	var as = assert.New(t)
	{
		var mc = New[string, any](WithBucketSize(0, 0))
		as.Equal(mc.conf.BucketSize, defaultBucketSize)
		as.Equal(mc.conf.BucketCap, defaultBucketCap)
	}
	{
		var mc = New[string, any](WithBucketSize(100, 1000))
		as.Equal(mc.conf.BucketSize, 100)
		as.Equal(mc.conf.BucketCap, 1000)
	}
}

func TestWithDeleteLimits(t *testing.T) {
	var as = assert.New(t)
	{
		var mc = New[string, any](WithDeleteLimits(0))
		as.Equal(mc.conf.DeleteLimits, defaultDeleteLimits)
	}
	{
		var mc = New[string, any](WithDeleteLimits(10))
		as.Equal(mc.conf.DeleteLimits, 10)
	}
}

func TestWithTimeCache(t *testing.T) {
	var as = assert.New(t)
	{
		var mc = New[string, any]()
		as.Equal(mc.conf.CachedTime, true)
	}
	{
		var mc = New[string, any](WithCachedTime(false))
		as.Equal(mc.conf.CachedTime, false)
	}
}

func TestWithSwissTable(t *testing.T) {
	t.Run("", func(t *testing.T) {
		var mc = New[string, int](
			WithSwissTable(true),
		)
		_, ok := mc.storage[0].Map.(*swiss.Map[string, *Element[string, int]])
		assert.True(t, ok)
		assert.True(t, mc.conf.SwissTable)
	})

	t.Run("", func(t *testing.T) {
		var mc = New[string, int]()
		_, ok := mc.storage[0].Map.(containers.Map[string, *Element[string, int]])
		assert.True(t, ok)
		assert.False(t, mc.conf.SwissTable)
	})
}
