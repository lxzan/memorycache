package memorycache

// newHeap 新建一个堆
// Create a new heap
func newHeap(cap int) *heap {
	return &heap{Data: make([]*Element, 0, cap)}
}

type heap struct {
	Data []*Element
}

func (c *heap) Less(i, j int) bool { return c.Data[i].ExpireAt < c.Data[j].ExpireAt }

func (c *heap) min(i, j int) int {
	if c.Data[i].ExpireAt < c.Data[j].ExpireAt {
		return i
	}
	return j
}

func (c *heap) Len() int {
	return len(c.Data)
}

func (c *heap) Swap(i, j int) {
	c.Data[i].index, c.Data[j].index = c.Data[j].index, c.Data[i].index
	c.Data[i], c.Data[j] = c.Data[j], c.Data[i]
}

func (c *heap) Push(ele *Element) {
	ele.index = c.Len()
	c.Data = append(c.Data, ele)
	c.Up(c.Len() - 1)
}

func (c *heap) Up(i int) {
	var j = (i - 1) >> 2
	if i >= 1 && c.Less(i, j) {
		c.Swap(i, j)
		c.Up(j)
	}
}

func (c *heap) Pop() (ele *Element) {
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

func (c *heap) Delete(i int) {
	n := c.Len()
	c.Swap(i, n-1)
	c.Data = c.Data[:n-1]
	c.Down(i, n-1)
}

func (c *heap) Down(i, n int) {
	var index1 = i<<2 + 1
	if index1 >= n {
		return
	}

	var index2 = i<<2 + 2
	var index3 = i<<2 + 3
	var index4 = i<<2 + 4
	var j = -1

	if index4 < n {
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
// Accessing the top Element of the heap
func (c *heap) Front() *Element {
	return c.Data[0]
}
