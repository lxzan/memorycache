package memorycache

import (
	"github.com/lxzan/memorycache/internal/utils"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"sort"
	"testing"
)

func isSorted[K comparable, V any](h *heap[K, V]) bool {
	var list0 []int
	var list1 []int
	for h.Len() > 0 {
		v := h.Pop()
		list0 = append(list0, int(v.ExpireAt))
		list1 = append(list1, int(v.ExpireAt))
	}
	sort.Ints(list1)
	return utils.IsSameSlice(list0, list1)
}

func TestHeap_Sort(t *testing.T) {
	var as = assert.New(t)
	var h = newHeap[string, int](0)
	for i := 0; i < 1000; i++ {
		num := rand.Int63n(1000)
		h.Push(&Element[string, int]{ExpireAt: num})
	}

	as.LessOrEqual(h.Front().ExpireAt, h.Data[1].ExpireAt)
	as.LessOrEqual(h.Front().ExpireAt, h.Data[2].ExpireAt)
	as.LessOrEqual(h.Front().ExpireAt, h.Data[3].ExpireAt)
	as.LessOrEqual(h.Front().ExpireAt, h.Data[4].ExpireAt)
	as.True(isSorted(h))
	as.Nil(h.Pop())
}

func TestHeap_Delete(t *testing.T) {
	var as = assert.New(t)
	var h = newHeap[string, int](0)
	h.Push(&Element[string, int]{ExpireAt: 1})
	h.Push(&Element[string, int]{ExpireAt: 2})
	h.Push(&Element[string, int]{ExpireAt: 3})
	h.Push(&Element[string, int]{ExpireAt: 4})
	h.Push(&Element[string, int]{ExpireAt: 5})
	h.Push(&Element[string, int]{ExpireAt: 6})
	h.Push(&Element[string, int]{ExpireAt: 7})
	h.Push(&Element[string, int]{ExpireAt: 8})
	h.Push(&Element[string, int]{ExpireAt: 9})
	h.Push(&Element[string, int]{ExpireAt: 10})
	h.Delete(3)
	h.Delete(5)

	var list []int64
	for _, item := range h.Data {
		list = append(list, item.ExpireAt)
	}
	as.ElementsMatch(list, []int64{1, 2, 3, 8, 5, 9, 7, 10})
}
