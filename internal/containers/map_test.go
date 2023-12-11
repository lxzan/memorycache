package containers

import (
	"testing"

	"github.com/dolthub/swiss"
	"github.com/stretchr/testify/assert"
)

func TestHashMap(t *testing.T) {
	var m = HashMap[string, any]{}
	assert.Equal(t, m.Count(), 0)
	m.Put("a", 1)
	assert.Equal(t, m.Count(), 1)

	{
		v, ok := m.Get("a")
		assert.True(t, ok)
		assert.Equal(t, v, 1)
	}

	{
		m.Delete("a")
		_, ok := m.Get("a")
		assert.False(t, ok)
	}
}

func TestHashMap_Iter(t *testing.T) {
	var m = HashMap[string, any]{}
	m.Put("a", 1)
	m.Put("b", 2)
	m.Put("c", 3)

	t.Run("", func(t *testing.T) {
		var keys []string
		m.Iter(func(s string, a any) bool {
			keys = append(keys, s)
			return true
		})
		assert.ElementsMatch(t, keys, []string{"a", "b", "c"})
	})

	t.Run("", func(t *testing.T) {
		var keys []string
		m.Iter(func(s string, a any) bool {
			keys = append(keys, s)
			return len(keys) < 2
		})
		assert.Equal(t, len(keys), 2)
	})
}

func TestNewMap(t *testing.T) {
	t.Run("", func(t *testing.T) {
		m := NewMap[string, any](1, true)
		_, ok := m.(*swiss.Map[string, any])
		assert.True(t, ok)
	})

	t.Run("", func(t *testing.T) {
		m := NewMap[string, any](1, false)
		_, ok := m.(HashMap[string, any])
		assert.True(t, ok)
	})
}
