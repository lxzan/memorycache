package memorycache

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/lxzan/memorycache/internal/utils"
	"github.com/stretchr/testify/assert"
)

func isSorted[K comparable, V any](h *heap[K, V]) bool {
	var list0 []int
	var list1 []int
	for h.Len() > 0 {
		addr := h.Pop()
		v := h.List.Get(addr)
		list0 = append(list0, int(v.ExpireAt))
		list1 = append(list1, int(v.ExpireAt))
	}
	sort.Ints(list1)
	return utils.IsSameSlice(list0, list1)
}

func TestHeap_Sort(t *testing.T) {
	var as = assert.New(t)
	var q = newDeque[string, int](0)
	var h = newHeap[string, int](q, 0)
	for i := 0; i < 1000; i++ {
		num := rand.Int63n(1000)
		ele := q.PushBack()
		ele.ExpireAt = num
		h.Push(ele)
	}

	as.LessOrEqual(h.Front().ExpireAt, h.List.Get(h.Data[1]).ExpireAt)
	as.LessOrEqual(h.Front().ExpireAt, h.List.Get(h.Data[2]).ExpireAt)
	as.LessOrEqual(h.Front().ExpireAt, h.List.Get(h.Data[3]).ExpireAt)
	as.LessOrEqual(h.Front().ExpireAt, h.List.Get(h.Data[4]).ExpireAt)
	as.True(isSorted(h))
	as.Zero(h.Pop())
}

func TestHeap_Delete(t *testing.T) {
	var as = assert.New(t)
	var q = newDeque[string, int](0)
	var h = newHeap[string, int](q, 0)
	var push = func(exp int64) {
		ele := q.PushBack()
		ele.ExpireAt = exp
		h.Push(ele)
	}
	push(1)
	push(2)
	push(3)
	push(4)
	push(5)
	push(6)
	push(7)
	push(8)
	push(9)
	push(10)
	h.Delete(3)
	h.Delete(5)

	var list []int64
	for _, item := range h.Data {
		ele := h.List.Get(item)
		list = append(list, ele.ExpireAt)
	}
	as.ElementsMatch(list, []int64{1, 2, 3, 8, 5, 9, 7, 10})
}
