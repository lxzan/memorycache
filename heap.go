package memorycache

import "github.com/lxzan/dao/algorithm"

// newHeap 新建一个堆
// Create a new heap
func newHeap[K comparable, V any](q *deque[K, V], cap int) *heap[K, V] {
	return &heap[K, V]{List: q, Data: make([]pointer, 0, cap)}
}

type heap[K comparable, V any] struct {
	List *deque[K, V]
	Data []pointer
}

func (c *heap[K, V]) Less(i, j int) bool {
	p0 := c.Data[i]
	p1 := c.Data[j]
	return c.List.Get(p0).ExpireAt < c.List.Get(p1).ExpireAt
}

func (c *heap[K, V]) UpdateTTL(ele *Element[K, V], exp int64) {
	var down = exp > ele.ExpireAt
	ele.ExpireAt = exp
	if down {
		c.Down(ele.index, c.Len())
	} else {
		c.Up(ele.index)
	}
}

func (c *heap[K, V]) Len() int {
	return len(c.Data)
}

func (c *heap[K, V]) Swap(i, j int) {
	x := c.List.Get(c.Data[i])
	y := c.List.Get(c.Data[j])
	x.index, y.index = y.index, x.index
	c.Data[i], c.Data[j] = c.Data[j], c.Data[i]
}

func (c *heap[K, V]) Push(ele *Element[K, V]) {
	ele.index = c.Len()
	c.Data = append(c.Data, ele.addr)
	c.Up(c.Len() - 1)
}

func (c *heap[K, V]) Up(i int) {
	var j = (i - 1) >> 2
	if i >= 1 && c.Less(i, j) {
		c.Swap(i, j)
		c.Up(j)
	}
}

func (c *heap[K, V]) Pop() (ele pointer) {
	var n = c.Len()
	switch n {
	case 0:
	case 1:
		ele = c.Data[0]
		c.Data = c.Data[:0]
	default:
		ele = c.Data[0]
		c.Swap(0, n-1)
		c.Data = c.Data[:n-1]
		c.Down(0, n-1)
	}
	return
}

func (c *heap[K, V]) Delete(i int) {
	if i == 0 {
		c.Pop()
		return
	}

	var n = c.Len()
	var down = c.Less(i, n-1)
	c.Swap(i, n-1)
	c.Data = c.Data[:n-1]
	if i < n-1 {
		if down {
			c.Down(i, n-1)
		} else {
			c.Up(i)
		}
	}
}

func (c *heap[K, V]) Down(i, n int) {
	var base = i << 2
	var index = base + 1
	if index >= n {
		return
	}

	var end = algorithm.Min(base+4, n-1)
	for j := base + 2; j <= end; j++ {
		if c.Less(j, index) {
			index = j
		}
	}

	if c.Less(index, i) {
		c.Swap(i, index)
		c.Down(index, n)
	}
}

// Front 访问堆顶元素
// Accessing the top Element of the heap
func (c *heap[K, V]) Front() *Element[K, V] {
	return c.List.Get(c.Data[0])
}
