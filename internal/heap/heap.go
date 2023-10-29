package heap

import "github.com/lxzan/memorycache/internal/types"

// New 新建一个堆
// Create a new heap
func New(cap int) *Heap {
	return &Heap{Data: make([]*types.Element, 0, cap)}
}

type Heap struct {
	Data []*types.Element
}

func (c *Heap) Less(i, j int) bool { return c.Data[i].ExpireAt < c.Data[j].ExpireAt }

func (c *Heap) min(i, j int) int {
	if c.Data[i].ExpireAt < c.Data[j].ExpireAt {
		return i
	}
	return j
}

func (c *Heap) Len() int {
	return len(c.Data)
}

func (c *Heap) Swap(i, j int) {
	c.Data[i].Index, c.Data[j].Index = c.Data[j].Index, c.Data[i].Index
	c.Data[i], c.Data[j] = c.Data[j], c.Data[i]
}

func (c *Heap) Push(ele *types.Element) {
	ele.Index = c.Len()
	c.Data = append(c.Data, ele)
	c.Up(c.Len() - 1)
}

func (c *Heap) Up(i int) {
	if i > 0 {
		var j = (i - 1) >> 2
		if c.Less(i, j) {
			c.Swap(i, j)
			c.Up(j)
		}
	}
}

func (c *Heap) Pop() (ele *types.Element) {
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

func (c *Heap) Delete(i int) {
	n := c.Len()
	c.Swap(i, n-1)
	c.Data = c.Data[:n-1]
	c.Down(i, n-1)
}

func (c *Heap) Down(i, n int) {
	var j = -1
	var index1 = i<<2 + 1
	var index2 = i<<2 + 2
	var index3 = i<<2 + 3
	var index4 = i<<2 + 4

	if index1 >= n {
		return
	} else if index4 < n {
		j = c.min(c.min(index1, index2), c.min(index3, index4))
	} else if index3 < n {
		j = c.min(c.min(index1, index2), index3)
	} else if index2 < n {
		j = c.min(index1, index2)
	} else {
		j = index1
	}

	if j >= 0 && c.Less(j, i) {
		c.Swap(i, j)
		c.Down(j, n)
	}
}

// Front 访问堆顶元素
// Accessing the top element of the heap
func (c *Heap) Front() *types.Element {
	return c.Data[0]
}
