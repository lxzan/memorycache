package heap

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHeap(t *testing.T) {
	var as = assert.New(t)
	var h = Heap(make([]Element, 0))
	h.Push(Element{ExpireAt: 1})
	h.Push(Element{ExpireAt: 3})
	h.Push(Element{ExpireAt: 5})
	h.Push(Element{ExpireAt: 7})
	h.Push(Element{ExpireAt: 9})
	h.Push(Element{ExpireAt: 2})
	h.Push(Element{ExpireAt: 4})
	h.Push(Element{ExpireAt: 6})
	h.Push(Element{ExpireAt: 8})
	h.Push(Element{ExpireAt: 10})
	var listA = make([]int64, 0)
	for h.Len() > 0 {
		listA = append(listA, h.Pop().ExpireAt)
	}
	as.ElementsMatch(listA, []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
}
