package heap

import (
	"github.com/lxzan/memorycache/internal/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHeap_Sort(t *testing.T) {
	var as = assert.New(t)
	var h = New(0)
	h.Push(&types.Element{ExpireAt: 1})
	h.Push(&types.Element{ExpireAt: 3})
	h.Push(&types.Element{ExpireAt: 5})
	h.Push(&types.Element{ExpireAt: 7})
	h.Push(&types.Element{ExpireAt: 9})
	h.Push(&types.Element{ExpireAt: 2})
	h.Push(&types.Element{ExpireAt: 4})
	h.Push(&types.Element{ExpireAt: 6})
	h.Push(&types.Element{ExpireAt: 8})
	h.Push(&types.Element{ExpireAt: 10})

	as.Equal(h.Front().ExpireAt, int64(1))
	var listA = make([]int64, 0)
	for h.Len() > 0 {
		listA = append(listA, h.Pop().ExpireAt)
	}
	as.ElementsMatch(listA, []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	as.Nil(h.Pop())
}

func TestHeap_Delete(t *testing.T) {
	var as = assert.New(t)
	var h = New(0)
	h.Push(&types.Element{ExpireAt: 1})
	h.Push(&types.Element{ExpireAt: 2})
	h.Push(&types.Element{ExpireAt: 3})
	h.Push(&types.Element{ExpireAt: 4})
	h.Push(&types.Element{ExpireAt: 5})
	h.Push(&types.Element{ExpireAt: 6})
	h.Push(&types.Element{ExpireAt: 7})
	h.Push(&types.Element{ExpireAt: 8})
	h.Push(&types.Element{ExpireAt: 9})
	h.Push(&types.Element{ExpireAt: 10})
	h.Delete(3)
	h.Delete(5)

	var list []int64
	for _, item := range h.Data {
		list = append(list, item.ExpireAt)
	}
	as.ElementsMatch(list, []int64{1, 2, 3, 8, 5, 9, 7, 10})
}
