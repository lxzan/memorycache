package heap

type (
	Heap []Element

	Element struct {
		Key      string
		ExpireAt int64
	}
)

func (self Heap) Len() int {
	return len(self)
}

func (self *Heap) Swap(i, j int) {
	(*self)[i], (*self)[j] = (*self)[j], (*self)[i]
}

func (self *Heap) Push(item Element) {
	*self = append(*self, item)
	self.Up(self.Len() - 1)
}

func (self *Heap) Up(i int) {
	var j = (i - 1) / 2
	if j >= 0 && (*self)[i].ExpireAt < (*self)[j].ExpireAt {
		self.Swap(i, j)
		self.Up(j)
	}
}

func (self *Heap) Pop() Element {
	var n = self.Len()
	var result = (*self)[0]
	(*self)[0] = (*self)[n-1]
	*self = (*self)[:n-1]
	self.Down(0, n-1)
	return result
}

func (self *Heap) Down(i, n int) {
	var j = 2*i + 1
	if j < n && (*self)[j].ExpireAt < (*self)[i].ExpireAt {
		self.Swap(i, j)
		self.Down(j, n)
	}
	var k = 2*i + 2
	if k < n && (*self)[k].ExpireAt < (*self)[i].ExpireAt {
		self.Swap(i, k)
		self.Down(k, n)
	}
}
