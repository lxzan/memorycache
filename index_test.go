package memorycache

import (
	"github.com/lxzan/memorycache/internal/utils"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	var as = assert.New(t)

	t.Run("", func(t *testing.T) {
		var db = New(WithInterval(10 * time.Millisecond))
		db.Set("a", 1, 10*time.Millisecond)
		db.Set("b", 1, 30*time.Millisecond)
		db.Set("c", 1, 50*time.Millisecond)
		db.Set("d", 1, 70*time.Millisecond)
		db.Set("e", 1, 90*time.Millisecond)
		db.Set("c", 1, time.Millisecond)

		time.Sleep(20 * time.Millisecond)
		as.ElementsMatch(db.Keys("*"), []string{"b", "d", "e"})
	})

	t.Run("", func(t *testing.T) {
		var db = New(WithInterval(10 * time.Millisecond))
		db.Set("a", 1, 10*time.Millisecond)
		db.Set("b", 1, 20*time.Millisecond)
		db.Set("c", 1, 50*time.Millisecond)
		db.Set("d", 1, 70*time.Millisecond)
		db.Set("e", 1, 290*time.Millisecond)
		db.Set("a", 1, 40*time.Millisecond)

		time.Sleep(30 * time.Millisecond)
		as.ElementsMatch(db.Keys("*"), []string{"a", "c", "d", "e"})
	})

	t.Run("", func(t *testing.T) {
		var db = New(WithInterval(10 * time.Millisecond))
		db.Set("a", 1, 10*time.Millisecond)
		db.Set("b", 1, 20*time.Millisecond)
		db.Set("c", 1, 40*time.Millisecond)
		db.Set("d", 1, 70*time.Millisecond)
		db.Set("d", 1, 40*time.Millisecond)

		time.Sleep(50 * time.Millisecond)
		as.Equal(0, db.Len())
	})
}

func TestMemoryCache_Set(t *testing.T) {
	var list []string
	var count = 10000
	var mc = New(WithInterval(100 * time.Millisecond))
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
	assert.ElementsMatch(t, utils.Uniq(list), mc.Keys("*"))
}

func TestMemoryCache_Get(t *testing.T) {
	var list0 []string
	var list1 []string
	var count = 10000
	var mc = New(WithInterval(100 * time.Millisecond))
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
	var mc = New(WithInterval(100 * time.Millisecond))
	for i := 0; i < count; i++ {
		key := string(utils.AlphabetNumeric.Generate(8))
		exp := rand.Intn(1000) + 200
		list = append(list, key)
		mc.Set(key, 1, time.Duration(exp)*time.Millisecond)
	}
	var keys = mc.Keys("*")
	for _, key := range keys {
		mc.GetAndRefresh(key, 2*time.Second)
	}

	time.Sleep(1100 * time.Millisecond)

	for _, item := range list {
		_, ok := mc.Get(item)
		assert.True(t, ok)
	}

	mc.Delete(list[0])
	_, ok := mc.GetAndRefresh(list[0], -1)
	assert.False(t, ok)
}

func TestMemoryCache_Delete(t *testing.T) {
	var count = 10000
	var mc = New(WithInterval(100 * time.Millisecond))
	for i := 0; i < count; i++ {
		key := string(utils.AlphabetNumeric.Generate(8))
		exp := rand.Intn(1000) + 200
		mc.Set(key, 1, time.Duration(exp)*time.Millisecond)
	}

	var keys = mc.Keys("*")
	for i := 0; i < 100; i++ {
		deleted := mc.Delete(keys[i])
		assert.True(t, deleted)

		key := string(utils.AlphabetNumeric.Generate(8))
		deleted = mc.Delete(key)
		assert.False(t, deleted)
	}
	assert.Equal(t, mc.Len(), count-100)
}
