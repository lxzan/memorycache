package memorycache

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/lxzan/memorycache/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	var as = assert.New(t)

	t.Run("", func(t *testing.T) {
		var db = New(WithInterval(10*time.Millisecond, 10*time.Millisecond))
		db.Set("a", 1, 10*time.Millisecond)
		db.Set("b", 1, 30*time.Millisecond)
		db.Set("c", 1, 50*time.Millisecond)
		db.Set("d", 1, 70*time.Millisecond)
		db.Set("e", 1, 90*time.Millisecond)
		db.Set("c", 1, time.Millisecond)

		time.Sleep(20 * time.Millisecond)
		as.ElementsMatch(db.Keys(""), []string{"b", "d", "e"})
	})

	t.Run("", func(t *testing.T) {
		var db = New(WithInterval(10*time.Millisecond, 10*time.Millisecond))
		db.Set("a", 1, 10*time.Millisecond)
		db.Set("b", 1, 20*time.Millisecond)
		db.Set("c", 1, 50*time.Millisecond)
		db.Set("d", 1, 70*time.Millisecond)
		db.Set("e", 1, 290*time.Millisecond)
		db.Set("a", 1, 40*time.Millisecond)

		time.Sleep(30 * time.Millisecond)
		as.ElementsMatch(db.Keys(""), []string{"a", "c", "d", "e"})
	})

	t.Run("", func(t *testing.T) {
		var db = New(WithInterval(10*time.Millisecond, 10*time.Millisecond))
		db.Set("a", 1, 10*time.Millisecond)
		db.Set("b", 1, 20*time.Millisecond)
		db.Set("c", 1, 40*time.Millisecond)
		db.Set("d", 1, 70*time.Millisecond)
		db.Set("d", 1, 40*time.Millisecond)

		time.Sleep(50 * time.Millisecond)
		as.Equal(0, len(db.Keys("")))
	})
}

func TestMemoryCache_Set(t *testing.T) {
	var list []string
	var count = 10000
	var mc = New(WithInterval(100*time.Millisecond, 100*time.Millisecond))
	mc.Clear()
	for i := 0; i < count; i++ {
		key := string(utils.AlphabetNumeric.Generate(8))
		exp := rand.Intn(1000)
		if exp == 0 {
			list = append(list, key)
		}
		mc.Set(key, 1, time.Duration(exp)*time.Millisecond)
	}
	for i := 0; i < count; i++ {
		key := string(utils.AlphabetNumeric.Generate(8))
		list = append(list, key)
		exp := rand.Intn(1000) + 3000
		mc.Set(key, 1, time.Duration(exp)*time.Millisecond)
	}
	time.Sleep(1100 * time.Millisecond)
	assert.ElementsMatch(t, utils.Uniq(list), mc.Keys(""))
}

func TestMemoryCache_Get(t *testing.T) {
	var list0 []string
	var list1 []string
	var count = 10000
	var mc = New(WithInterval(100*time.Millisecond, 100*time.Millisecond))
	for i := 0; i < count; i++ {
		key := string(utils.AlphabetNumeric.Generate(8))
		exp := rand.Intn(1000)
		if exp == 0 {
			list1 = append(list1, key)
		} else {
			list0 = append(list0, key)
		}
		mc.Set(key, 1, time.Duration(exp)*time.Millisecond)
	}
	for i := 0; i < count; i++ {
		key := string(utils.AlphabetNumeric.Generate(8))
		list1 = append(list1, key)
		exp := rand.Intn(1000) + 3000
		mc.Set(key, 1, time.Duration(exp)*time.Millisecond)
	}
	time.Sleep(1100 * time.Millisecond)

	for _, item := range list0 {
		_, ok := mc.Get(item)
		assert.False(t, ok)
	}
	for _, item := range list1 {
		_, ok := mc.Get(item)
		assert.True(t, ok)
	}
}

func TestMemoryCache_GetAndRefresh(t *testing.T) {
	var list []string
	var count = 10000
	var mc = New(WithInterval(100*time.Millisecond, 100*time.Millisecond))
	for i := 0; i < count; i++ {
		key := string(utils.AlphabetNumeric.Generate(8))
		exp := rand.Intn(1000) + 200
		list = append(list, key)
		mc.Set(key, 1, time.Duration(exp)*time.Millisecond)
	}
	var keys = mc.Keys("")
	for _, key := range keys {
		mc.GetWithTTL(key, 2*time.Second)
	}

	time.Sleep(1100 * time.Millisecond)

	for _, item := range list {
		_, ok := mc.Get(item)
		assert.True(t, ok)
	}

	mc.Delete(list[0])
	_, ok := mc.GetWithTTL(list[0], -1)
	assert.False(t, ok)
}

func TestMemoryCache_Delete(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		var count = 10000
		var mc = New(WithInterval(100*time.Millisecond, 100*time.Millisecond))
		for i := 0; i < count; i++ {
			key := string(utils.AlphabetNumeric.Generate(8))
			exp := rand.Intn(1000) + 200
			mc.Set(key, 1, time.Duration(exp)*time.Millisecond)
		}

		var keys = mc.Keys("")
		for i := 0; i < 100; i++ {
			deleted := mc.Delete(keys[i])
			assert.True(t, deleted)

			key := string(utils.AlphabetNumeric.Generate(8))
			deleted = mc.Delete(key)
			assert.False(t, deleted)
		}
		assert.Equal(t, mc.Len(), count-100)
	})

	t.Run("2", func(t *testing.T) {
		var mc = New()
		var wg = &sync.WaitGroup{}
		wg.Add(1)
		mc.SetWithCallback("ming", 1, -1, func(ele *Element, reason Reason) {
			assert.Equal(t, reason, ReasonDeleted)
			wg.Done()
		})
		mc.SetWithCallback("ting", 2, -1, func(ele *Element, reason Reason) {
			wg.Done()
		})
		go mc.Delete("ming")
		wg.Wait()
	})

	t.Run("3", func(t *testing.T) {
		var mc = New()
		var wg = &sync.WaitGroup{}
		wg.Add(1)
		mc.GetOrCreateWithCallback("ming", 1, -1, func(ele *Element, reason Reason) {
			assert.Equal(t, reason, ReasonDeleted)
			wg.Done()
		})
		mc.GetOrCreateWithCallback("ting", 2, -1, func(ele *Element, reason Reason) {
			wg.Done()
		})
		go mc.Delete("ting")
		wg.Wait()
	})
}

func TestMaxCap(t *testing.T) {
	var mc = New(
		WithBucketNum(1),
		WithBucketSize(10, 100),
		WithInterval(100*time.Millisecond, 100*time.Millisecond),
	)

	var wg = &sync.WaitGroup{}
	wg.Add(900)
	for i := 0; i < 1000; i++ {
		key := string(utils.AlphabetNumeric.Generate(16))
		mc.SetWithCallback(key, 1, -1, func(ele *Element, reason Reason) {
			assert.Equal(t, reason, ReasonOverflow)
			wg.Done()
		})
	}
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, mc.Len(), 100)
	wg.Wait()
}

func TestMemoryCache_SetWithCallback(t *testing.T) {
	var as = assert.New(t)
	var count = 1000
	var mc = New(
		WithBucketNum(16),
		WithInterval(10*time.Millisecond, 100*time.Millisecond),
	)
	defer mc.Clear()

	var wg = &sync.WaitGroup{}
	wg.Add(count)
	for i := 0; i < count; i++ {
		key := string(utils.AlphabetNumeric.Generate(16))
		exp := time.Duration(rand.Intn(1000)+10) * time.Millisecond
		mc.SetWithCallback(key, i, exp, func(ele *Element, reason Reason) {
			as.True(time.Now().UnixMilli() > ele.ExpireAt)
			as.Equal(reason, ReasonExpired)
			wg.Done()
		})
	}
	wg.Wait()
}

func TestMemoryCache_GetOrCreate(t *testing.T) {

	var count = 1000
	var mc = New(
		WithBucketNum(16),
		WithInterval(10*time.Millisecond, 100*time.Millisecond),
	)
	defer mc.Clear()

	for i := 0; i < count; i++ {
		key := string(utils.AlphabetNumeric.Generate(16))
		exp := time.Duration(rand.Intn(1000)+10) * time.Millisecond
		mc.GetOrCreate(key, i, exp)
	}
}

func TestMemoryCache_GetOrCreateWithCallback(t *testing.T) {
	var as = assert.New(t)
	var count = 1000
	var mc = New(
		WithBucketNum(16),
		WithInterval(10*time.Millisecond, 100*time.Millisecond),
	)
	defer mc.Clear()

	var wg = &sync.WaitGroup{}
	wg.Add(count)
	for i := 0; i < count; i++ {
		key := string(utils.AlphabetNumeric.Generate(16))
		exp := time.Duration(rand.Intn(1000)+10) * time.Millisecond
		mc.GetOrCreateWithCallback(key, i, exp, func(ele *Element, reason Reason) {
			as.True(time.Now().UnixMilli() > ele.ExpireAt)
			as.Equal(reason, ReasonExpired)
			wg.Done()
		})
	}
	wg.Wait()
}
