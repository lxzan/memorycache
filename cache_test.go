package memorycache

import (
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/lxzan/memorycache/internal/utils"
	"github.com/stretchr/testify/assert"
)

func getKeys[K comparable, V any](db *MemoryCache[K, V]) []K {
	var keys []K
	db.Range(func(k K, v V) bool {
		keys = append(keys, k)
		return true
	})
	return keys
}

func TestMemoryCache(t *testing.T) {
	var as = assert.New(t)

	t.Run("", func(t *testing.T) {
		var db = New[string, any](
			WithInterval(10*time.Millisecond, 10*time.Millisecond),
			WithBucketNum(1),
			WithCachedTime(false),
		)
		db.Set("a", 1, 100*time.Millisecond)
		db.Set("b", 1, 300*time.Millisecond)
		db.Set("c", 1, 500*time.Millisecond)
		db.Set("d", 1, 700*time.Millisecond)
		db.Set("e", 1, 900*time.Millisecond)
		db.Set("c", 1, time.Millisecond)

		time.Sleep(200 * time.Millisecond)
		var keys = getKeys(db)
		as.ElementsMatch(keys, []string{"b", "d", "e"})
	})

	t.Run("", func(t *testing.T) {
		var db = New[string, any](
			WithInterval(10*time.Millisecond, 10*time.Millisecond),
			WithCachedTime(false),
		)
		db.Set("a", 1, 100*time.Millisecond)
		db.Set("b", 1, 200*time.Millisecond)
		db.Set("c", 1, 500*time.Millisecond)
		db.Set("d", 1, 700*time.Millisecond)
		db.Set("e", 1, 2900*time.Millisecond)
		db.Set("a", 1, 400*time.Millisecond)

		time.Sleep(300 * time.Millisecond)
		var keys = getKeys(db)
		as.ElementsMatch(keys, []string{"a", "c", "d", "e"})
	})

	t.Run("", func(t *testing.T) {
		var db = New[string, any](
			WithInterval(10*time.Millisecond, 10*time.Millisecond),
			WithCachedTime(false),
		)
		db.Set("a", 1, 100*time.Millisecond)
		db.Set("b", 1, 200*time.Millisecond)
		db.Set("c", 1, 400*time.Millisecond)
		db.Set("d", 1, 700*time.Millisecond)
		db.Set("d", 1, 400*time.Millisecond)

		time.Sleep(500 * time.Millisecond)
		var keys = getKeys(db)
		as.Equal(0, len(keys))
	})

	t.Run("batch", func(t *testing.T) {
		var count = 1000
		var mc = New[string, any](
			WithInterval(10*time.Millisecond, 10*time.Millisecond),
			WithBucketNum(1),
			WithCachedTime(false),
		)
		var m1 = make(map[string]int)
		var m2 = make(map[string]int64)
		for i := 0; i < count; i++ {
			key := string(utils.AlphabetNumeric.Generate(16))
			exp := time.Duration(rand.Intn(10)+1) * 100 * time.Millisecond
			mc.Set(key, i, exp)
			m1[key] = i
			m2[key] = mc.getExp(exp)
		}

		time.Sleep(500 * time.Millisecond)
		for k, v := range m1 {
			result, ok := mc.Get(k)
			if ts := time.Now().UnixMilli(); ts > m2[k] {
				if ts-m2[k] >= 10 {
					as.False(ok)
				}
				continue
			}

			as.True(ok)
			as.Equal(result.(int), v)
		}

		var wg = &sync.WaitGroup{}
		wg.Add(1)
		result, exist := mc.GetOrCreateWithCallback(string(utils.AlphabetNumeric.Generate(16)), "x", 500*time.Millisecond, func(ele *Element[string, any], reason Reason) {
			as.Equal(reason, ReasonExpired)
			as.Equal(ele.Value.(string), "x")
			wg.Done()
		})
		as.False(exist)
		as.Equal(result.(string), "x")
		wg.Wait()
	})

	t.Run("expire", func(t *testing.T) {
		var mc = New[string, any](
			WithBucketNum(1),
			WithDeleteLimits(3),
			WithInterval(50*time.Millisecond, 100*time.Millisecond),
			WithCachedTime(false),
		)
		mc.Set("a", 1, 150*time.Millisecond)
		mc.Set("b", 1, 150*time.Millisecond)
		mc.Set("c", 1, 150*time.Millisecond)
		time.Sleep(200 * time.Millisecond)
	})
}

func TestMemoryCache_Set(t *testing.T) {
	t.Run("", func(t *testing.T) {
		var list []string
		var count = 10000
		var mc = New[string, any](WithInterval(100*time.Millisecond, 100*time.Millisecond))
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
		var keys = getKeys(mc)
		assert.ElementsMatch(t, utils.Uniq(list), keys)
	})

	t.Run("evict", func(t *testing.T) {
		var mc = New[string, any](
			WithBucketNum(1),
			WithBucketSize(0, 2),
		)
		mc.Set("ming", 1, 3*time.Hour)
		mc.Set("hong", 1, 1*time.Hour)
		mc.Set("feng", 1, 2*time.Hour)
		var keys = getKeys(mc)
		assert.ElementsMatch(t, keys, []string{"hong", "feng"})
	})

	t.Run("update ttl", func(t *testing.T) {
		var mc = New[string, any](WithBucketNum(1))
		var count = 1000
		for i := 0; i < 10*count; i++ {
			key := strconv.Itoa(utils.Numeric.Intn(count))
			exp := time.Duration(utils.Numeric.Intn(count)+10) * time.Second
			mc.Set(key, 1, exp)
		}

		var list1 []int
		var list2 []int
		for _, b := range mc.storage {
			b.Lock()
			for _, item := range b.Heap.Data {
				ele := b.List.Get(item)
				list1 = append(list1, int(ele.ExpireAt))
			}
			b.Unlock()
		}
		sort.Ints(list1)

		for _, b := range mc.storage {
			b.Lock()
			for b.Heap.Len() > 0 {
				ele := b.List.Get(b.Heap.Pop())
				list2 = append(list2, int(ele.ExpireAt))
			}
			b.Unlock()
		}

		assert.Equal(t, len(list1), len(list2))
		for i, v := range list2 {
			assert.Equal(t, list1[i], v)
		}
	})
}

func TestMemoryCache_Get(t *testing.T) {
	t.Run("", func(t *testing.T) {
		var list0 []string
		var list1 []string
		var count = 10000
		var mc = New[string, any](WithInterval(100*time.Millisecond, 100*time.Millisecond))
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
	})

	t.Run("expire", func(t *testing.T) {
		var mc = New[string, any](
			WithInterval(10*time.Second, 10*time.Second),
		)

		var wg = &sync.WaitGroup{}
		wg.Add(1)

		mc.SetWithCallback("ming", 128, 10*time.Millisecond, func(ele *Element[string, any], reason Reason) {
			assert.Equal(t, reason, ReasonExpired)
			assert.Equal(t, ele.Value.(int), 128)
			wg.Done()
		})

		time.Sleep(2 * time.Second)
		v, ok := mc.Get("ming")
		assert.False(t, ok)
		assert.Nil(t, v)
		wg.Wait()
	})
}

func TestMemoryCache_GetWithTTL(t *testing.T) {
	t.Run("", func(t *testing.T) {
		var list []string
		var count = 10000
		var mc = New[string, any](WithInterval(100*time.Millisecond, 100*time.Millisecond))
		for i := 0; i < count; i++ {
			key := string(utils.AlphabetNumeric.Generate(8))
			exp := rand.Intn(1000) + 200
			list = append(list, key)
			mc.Set(key, 1, time.Duration(exp)*time.Millisecond)
		}
		var keys = getKeys(mc)
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
	})

	t.Run("update ttl", func(t *testing.T) {
		var mc = New[string, any](WithBucketNum(1))
		var count = 1000
		for i := 0; i < count; i++ {
			key := strconv.Itoa(utils.Numeric.Intn(count))
			exp := time.Duration(utils.Numeric.Intn(count)+10) * time.Second
			mc.Set(key, 1, exp)
		}

		for i := 0; i < count; i++ {
			key := strconv.Itoa(utils.Numeric.Intn(count))
			exp := time.Duration(utils.Numeric.Intn(count)+10) * time.Second
			mc.GetWithTTL(key, exp)
		}

		var list1 []int
		var list2 []int
		for _, b := range mc.storage {
			b.Lock()
			for _, item := range b.Heap.Data {
				ele := b.List.Get(item)
				list1 = append(list1, int(ele.ExpireAt))
			}
			b.Unlock()
		}
		sort.Ints(list1)

		for _, b := range mc.storage {
			b.Lock()
			for b.Heap.Len() > 0 {
				ele := b.List.Get(b.Heap.Pop())
				list2 = append(list2, int(ele.ExpireAt))
			}
			b.Unlock()
		}

		assert.Equal(t, len(list1), len(list2))
		for i, v := range list2 {
			assert.Equal(t, list1[i], v)
		}
	})
}

func TestMemoryCache_Delete(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		var count = 10000
		var mc = New[string, any](WithInterval(100*time.Millisecond, 100*time.Millisecond))
		for i := 0; i < count; i++ {
			key := string(utils.AlphabetNumeric.Generate(8))
			exp := rand.Intn(1000) + 200
			mc.Set(key, 1, time.Duration(exp)*time.Millisecond)
		}

		var keys = getKeys(mc)
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
		var mc = New[string, any]()
		var wg = &sync.WaitGroup{}
		wg.Add(1)
		mc.SetWithCallback("ming", 1, -1, func(ele *Element[string, any], reason Reason) {
			assert.Equal(t, reason, ReasonDeleted)
			wg.Done()
		})
		mc.SetWithCallback("ting", 2, -1, func(ele *Element[string, any], reason Reason) {
			wg.Done()
		})
		go mc.Delete("ming")
		wg.Wait()
	})

	t.Run("3", func(t *testing.T) {
		var mc = New[string, any]()
		var wg = &sync.WaitGroup{}
		wg.Add(1)
		mc.GetOrCreateWithCallback("ming", 1, -1, func(ele *Element[string, any], reason Reason) {
			assert.Equal(t, reason, ReasonDeleted)
			wg.Done()
		})
		mc.GetOrCreateWithCallback("ting", 2, -1, func(ele *Element[string, any], reason Reason) {
			wg.Done()
		})
		go mc.Delete("ting")
		wg.Wait()
	})

	t.Run("batch delete", func(t *testing.T) {
		var mc = New[string, any](WithBucketNum(1))
		var count = 1000
		for i := 0; i < count; i++ {
			key := strconv.Itoa(utils.Numeric.Intn(count))
			exp := time.Duration(utils.Numeric.Intn(count)+10) * time.Second
			mc.Set(key, 1, exp)
		}
		for i := 0; i < count/2; i++ {
			key := strconv.Itoa(utils.Numeric.Intn(count))
			mc.Delete(key)
		}

		var list1 []int
		var list2 []int
		for _, b := range mc.storage {
			b.Lock()
			for _, item := range b.Heap.Data {
				ele := b.List.Get(item)
				list1 = append(list1, int(ele.ExpireAt))
			}
			b.Unlock()
		}
		sort.Ints(list1)

		for _, b := range mc.storage {
			b.Lock()
			for b.Heap.Len() > 0 {
				ele := b.List.Get(b.Heap.Pop())
				list2 = append(list2, int(ele.ExpireAt))
			}
			b.Unlock()
		}

		assert.Equal(t, len(list1), len(list2))
		for i, v := range list2 {
			assert.Equal(t, list1[i], v)
		}
	})
}

func TestMaxCap(t *testing.T) {
	var mc = New[string, any](
		WithBucketNum(1),
		WithBucketSize(10, 100),
		WithInterval(100*time.Millisecond, 100*time.Millisecond),
	)

	var wg = &sync.WaitGroup{}
	wg.Add(900)
	for i := 0; i < 1000; i++ {
		key := string(utils.AlphabetNumeric.Generate(16))
		mc.SetWithCallback(key, 1, -1, func(ele *Element[string, any], reason Reason) {
			assert.Equal(t, reason, ReasonEvicted)
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
	var mc = New[string, any](
		WithBucketNum(16),
		WithInterval(10*time.Millisecond, 100*time.Millisecond),
	)
	defer mc.Clear()

	var wg = &sync.WaitGroup{}
	wg.Add(count)
	for i := 0; i < count; i++ {
		key := string(utils.AlphabetNumeric.Generate(16))
		exp := time.Duration(rand.Intn(1000)+10) * time.Millisecond
		mc.SetWithCallback(key, i, exp, func(ele *Element[string, any], reason Reason) {
			as.True(time.Now().UnixMilli() > ele.ExpireAt)
			as.Equal(reason, ReasonExpired)
			wg.Done()
		})
	}
	wg.Wait()
}

func TestMemoryCache_GetOrCreate(t *testing.T) {

	var count = 1000
	var mc = New[string, any](
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

	t.Run("", func(t *testing.T) {
		var count = 1000
		var mc = New[string, any](
			WithBucketNum(16),
			WithInterval(10*time.Millisecond, 100*time.Millisecond),
		)
		defer mc.Clear()

		var wg = &sync.WaitGroup{}
		wg.Add(count)
		for i := 0; i < count; i++ {
			key := string(utils.AlphabetNumeric.Generate(16))
			exp := time.Duration(rand.Intn(1000)+10) * time.Millisecond
			mc.GetOrCreateWithCallback(key, i, exp, func(ele *Element[string, any], reason Reason) {
				as.True(time.Now().UnixMilli() > ele.ExpireAt)
				as.Equal(reason, ReasonExpired)
				wg.Done()
			})
		}
		wg.Wait()
	})

	t.Run("exists", func(t *testing.T) {
		var mc = New[string, any]()
		mc.Set("ming", 1, -1)
		v, exist := mc.GetOrCreateWithCallback("ming", 2, time.Second, func(ele *Element[string, any], reason Reason) {})
		as.True(exist)
		as.Equal(v.(int), 1)
	})

	t.Run("create", func(t *testing.T) {
		var mc = New[string, any](
			WithBucketNum(1),
			WithBucketSize(0, 1),
		)
		mc.Set("ming", 1, -1)
		v, exist := mc.GetOrCreateWithCallback("wang", 2, time.Second, func(ele *Element[string, any], reason Reason) {})
		as.False(exist)
		as.Equal(v.(int), 2)
		as.Equal(mc.Len(), 1)
	})
}

func TestMemoryCache_Stop(t *testing.T) {
	var mc = New[string, any]()
	mc.Stop()
	mc.Stop()

	select {
	case <-mc.ctx.Done():
	default:
		t.Fail()
	}
}

func TestMemoryCache_Range(t *testing.T) {
	const count = 1000
	var mc = New[string, int]()
	for i := 0; i < count; i++ {
		var key = string(utils.AlphabetNumeric.Generate(16))
		mc.Set(key, 1, time.Hour)
	}

	t.Run("", func(t *testing.T) {
		var keys []string
		mc.Range(func(s string, i int) bool {
			keys = append(keys, s)
			return true
		})
		assert.Equal(t, len(keys), count)
	})

	t.Run("", func(t *testing.T) {
		var keys []string
		mc.Range(func(s string, i int) bool {
			keys = append(keys, s)
			return len(keys) < 500
		})
		assert.Equal(t, len(keys), 500)
	})

	t.Run("", func(t *testing.T) {
		mc.Set("exp", 1, time.Millisecond)
		time.Sleep(100 * time.Millisecond)
		var keys []string
		mc.Range(func(s string, i int) bool {
			keys = append(keys, s)
			return true
		})
		assert.Equal(t, len(keys), count)
	})
}

func TestMemoryCache_LRU(t *testing.T) {
	const count = 10000
	var mc = New[string, int](
		WithBucketNum(1),
	)
	var indexes []int
	for i := 0; i < count; i++ {
		indexes = append(indexes, i)
		mc.Set(strconv.Itoa(i), 1, time.Hour)
	}
	for i := 0; i < count; i++ {
		a, b := utils.AlphabetNumeric.Intn(count), utils.AlphabetNumeric.Intn(count)
		indexes[a], indexes[b] = indexes[b], indexes[a]
	}
	for _, item := range indexes {
		key := strconv.Itoa(item)
		mc.Get(key)
	}

	var keys []string
	var q = mc.storage[0].List
	for q.Len() > 0 {
		keys = append(keys, q.PopFront().Key)
	}
	for i, item := range indexes {
		key := strconv.Itoa(item)
		assert.Equal(t, key, keys[i])
	}
}

func TestMemoryCache_Conflict(t *testing.T) {
	t.Run("", func(t *testing.T) {
		var mc = New[string, any]()
		var wg = &sync.WaitGroup{}
		wg.Add(1)
		mc.hasher = new(utils.Fnv32Hasher)
		mc.SetWithCallback("O4XOUsgCQqkVCvLQ", 1, time.Hour, func(element *Element[string, any], reason Reason) {
			assert.Equal(t, element.Value, 1)
			wg.Done()
		})
		assert.False(t, mc.Set("wYLAGPVADrDTi7VT", 2, time.Hour))
		assert.True(t, mc.Set("wYLAGPVADrDTi7VT", 2, time.Hour))

		v1, ok1 := mc.Get("O4XOUsgCQqkVCvLQ")
		assert.False(t, ok1)
		assert.Nil(t, v1)

		v2, ok2 := mc.Get("wYLAGPVADrDTi7VT")
		assert.True(t, ok2)
		assert.Equal(t, v2, 2)
		assert.Equal(t, mc.Len(), 1)
		wg.Wait()
	})

	t.Run("", func(t *testing.T) {
		var mc = New[string, any]()
		var wg = &sync.WaitGroup{}
		wg.Add(1)
		mc.hasher = new(utils.Fnv32Hasher)
		mc.SetWithCallback("O4XOUsgCQqkVCvLQ", 1, time.Hour, func(element *Element[string, any], reason Reason) {
			assert.Equal(t, element.Value, 1)
			wg.Done()
		})

		v1, ok1 := mc.Get("O4XOUsgCQqkVCvLQ")
		assert.True(t, ok1)
		assert.Equal(t, v1, 1)

		v2, ok2 := mc.GetOrCreate("wYLAGPVADrDTi7VT", 2, time.Hour)
		assert.False(t, ok2)
		assert.Equal(t, v2, 2)

		v3, ok3 := mc.GetOrCreate("wYLAGPVADrDTi7VT", 3, time.Hour)
		assert.True(t, ok3)
		assert.Equal(t, v3, 2)
		assert.Equal(t, mc.Len(), 1)

		wg.Wait()
	})
}

func TestMemoryCache_Random(t *testing.T) {
	t.Run("with lru", func(t *testing.T) {
		const count = 10000
		var mc = New[string, int](
			WithBucketNum(16),
			WithBucketSize(100, 625),
		)
		for i := 0; i < count; i++ {
			var key = string(utils.AlphabetNumeric.Generate(3))
			var val = utils.AlphabetNumeric.Intn(count)
			mc.Set(key, val, time.Hour)
		}

		for i := 0; i < count; i++ {
			var key = string(utils.AlphabetNumeric.Generate(3))
			var val = utils.AlphabetNumeric.Intn(count)
			switch utils.AlphabetNumeric.Intn(8) {
			case 0, 1:
				mc.Set(key, val, time.Hour)
			case 2:
				mc.SetWithCallback(key, val, time.Hour, func(entry *Element[string, int], reason Reason) {})
			case 3:
				mc.Get(key)
			case 4:
				mc.GetWithTTL(key, time.Hour)
			case 5:
				mc.GetOrCreate(key, val, time.Hour)
			case 6:
				mc.GetOrCreateWithCallback(key, val, time.Hour, func(entry *Element[string, int], reason Reason) {})
			case 7:
				mc.Delete(key)
			}
		}

		for _, b := range mc.storage {
			assert.Equal(t, b.Map.Count(), b.Heap.Len())
			assert.Equal(t, b.Heap.Len(), b.List.Len())
			b.List.Range(func(ele *Element[string, int]) bool {
				var v = ele
				var v1 = b.Heap.Data[v.index]
				assert.Equal(t, v.addr, v1)

				var v2, _ = b.Map.Get(v.hashcode)
				assert.Equal(t, v.addr, v2)
				return true
			})
			assert.True(t, isSorted(b.Heap))
		}
	})
}
