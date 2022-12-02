package heap

type (
	Heap []Element

	Element struct {
		Key      string
		ExpireAt int64
	}
)

func (c *Heap) Len() int {
	return len(*c)
}

func (c *Heap) Swap(i, j int) {
	(*c)[i], (*c)[j] = (*c)[j], (*c)[i]
}

func (c *Heap) Push(item Element) {
	*c = append(*c, item)
	c.Up(c.Len() - 1)
}

func (c *Heap) Up(i int) {
	var j = (i - 1) / 2
	if j >= 0 && (*c)[i].ExpireAt < (*c)[j].ExpireAt {
		c.Swap(i, j)
		c.Up(j)
	}
}

func (c *Heap) Pop() Element {
	var n = c.Len()
	var result = (*c)[0]
	(*c)[0] = (*c)[n-1]
	*c = (*c)[:n-1]
	c.Down(0, n-1)
	return result
}

func (c *Heap) Down(i, n int) {
	var j = 2*i + 1
	if j < n && (*c)[j].ExpireAt < (*c)[i].ExpireAt {
		c.Swap(i, j)
		c.Down(j, n)
	}
	var k = 2*i + 2
	if k < n && (*c)[k].ExpireAt < (*c)[i].ExpireAt {
		c.Swap(i, k)
		c.Down(k, n)
	}
}
