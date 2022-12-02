package heap

import (
	"testing"
)

func TestHeap(t *testing.T) {
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
	for h.Len() > 0 {
		println(h.Pop().ExpireAt)
	}
}
