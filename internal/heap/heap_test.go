package heap

import (
	"testing"
)

func TestHeap(t *testing.T) {
	var h = Heap(make([]Element, 0))
	h.Push(
		Element{ExpireAt: 1},
		Element{ExpireAt: 3},
		Element{ExpireAt: 5},
		Element{ExpireAt: 7},
		Element{ExpireAt: 9},
		Element{ExpireAt: 2},
		Element{ExpireAt: 4},
		Element{ExpireAt: 6},
		Element{ExpireAt: 8},
		Element{ExpireAt: 10},
	)
	for h.Len() > 0 {
		println(h.Pop().ExpireAt)
	}
}
