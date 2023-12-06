package memorycache

import (
	"github.com/lxzan/memorycache/internal/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQueue_Delete(t *testing.T) {
	t.Run("", func(t *testing.T) {
		var q = &queue[string, int]{}
		var e0 = &Element[string, int]{Key: "a"}
		var e1 = &Element[string, int]{Key: "b"}
		var e2 = &Element[string, int]{Key: "c"}
		var e3 = &Element[string, int]{Key: "d"}
		q.PushBack(e0)
		q.PushBack(e1)
		q.PushBack(e2)
		q.PushBack(e3)
		q.Delete(e0)
		assert.True(t, utils.IsSameSlice(q.Keys(), []string{"b", "c", "d"}))
		assert.Equal(t, q.head.Key, "b")
		assert.Equal(t, q.tail.Key, "d")
	})

	t.Run("", func(t *testing.T) {
		var q = &queue[string, int]{}
		var e0 = &Element[string, int]{Key: "a"}
		var e1 = &Element[string, int]{Key: "b"}
		var e2 = &Element[string, int]{Key: "c"}
		var e3 = &Element[string, int]{Key: "d"}
		q.PushBack(e0)
		q.PushBack(e1)
		q.PushBack(e2)
		q.PushBack(e3)
		q.Delete(e1)
		assert.True(t, utils.IsSameSlice(q.Keys(), []string{"a", "c", "d"}))
		assert.Equal(t, q.head.Key, "a")
		assert.Equal(t, q.tail.Key, "d")
	})

	t.Run("", func(t *testing.T) {
		var q = &queue[string, int]{}
		var e0 = &Element[string, int]{Key: "a"}
		var e1 = &Element[string, int]{Key: "b"}
		var e2 = &Element[string, int]{Key: "c"}
		var e3 = &Element[string, int]{Key: "d"}
		q.PushBack(e0)
		q.PushBack(e1)
		q.PushBack(e2)
		q.PushBack(e3)
		q.Delete(e3)
		assert.True(t, utils.IsSameSlice(q.Keys(), []string{"a", "b", "c"}))
		assert.Equal(t, q.head.Key, "a")
		assert.Equal(t, q.tail.Key, "c")
	})
}

func TestQueue_Pop(t *testing.T) {
	var q = &queue[string, int]{}
	assert.Nil(t, q.Pop())
	var e0 = &Element[string, int]{Key: "a"}
	var e1 = &Element[string, int]{Key: "b"}
	var e2 = &Element[string, int]{Key: "c"}
	var e3 = &Element[string, int]{Key: "d"}
	q.PushBack(e0)
	q.PushBack(e1)
	q.PushBack(e2)
	q.PushBack(e3)

	var keys []string
	for q.Len() > 0 {
		keys = append(keys, q.Pop().Key)
	}
	assert.True(t, utils.IsSameSlice(keys, []string{"a", "b", "c", "d"}))
}
