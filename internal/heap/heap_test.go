package heap

import (
	"github.com/lxzan/memorycache/internal/types"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"sort"
	"testing"
)

func TestHeap_Sort(t *testing.T) {
	var as = assert.New(t)
	var h = New(0)
	for i := 0; i < 1000; i++ {
		num := rand.Int63n(1000)
		h.Push(&types.Element{ExpireAt: num})
	}

	as.LessOrEqual(h.Front().ExpireAt, h.Data[1].ExpireAt)
	as.LessOrEqual(h.Front().ExpireAt, h.Data[2].ExpireAt)
	as.LessOrEqual(h.Front().ExpireAt, h.Data[3].ExpireAt)
	as.LessOrEqual(h.Front().ExpireAt, h.Data[4].ExpireAt)

	var list = make([]int64, 0)
	for h.Len() > 0 {
		list = append(list, h.Pop().ExpireAt)
	}
	ok := sort.SliceIsSorted(list, func(i, j int) bool {
		return list[i] < list[j]
	})
	as.True(ok)
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
