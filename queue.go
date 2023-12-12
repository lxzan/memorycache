package memorycache

func newQueue[K comparable, V any](lru bool) *queue[K, V] { return &queue[K, V]{enabled: lru} }

type queue[K comparable, V any] struct {
	length  int
	enabled bool
	head    *Element[K, V]
	tail    *Element[K, V]
}

func (c *queue[K, V]) Len() int { return c.length }

func (c *queue[K, V]) Front() *Element[K, V] { return c.head }

func (c *queue[K, V]) PushBack(ele *Element[K, V]) {
	if !c.enabled {
		return
	}

	c.length++
	if c.tail != nil {
		c.tail.next = ele
		ele.prev = c.tail
		c.tail = ele
		return
	}

	c.head = ele
	c.tail = ele
}

func (c *queue[K, V]) Pop() *Element[K, V] {
	if c.length == 0 || !c.enabled {
		return nil
	}
	head := c.Front()
	c.Delete(head)
	return head
}

// Delete it's safe delete in loop
func (c *queue[K, V]) Delete(ele *Element[K, V]) {
	if !c.enabled {
		return
	}

	var prev = ele.prev
	var next = ele.next
	var state = 0
	if prev != nil {
		state += 1
	}
	if next != nil {
		state += 2
	}

	switch state {
	case 3:
		prev.next = next
		next.prev = prev
	case 2:
		next.prev = nil
		c.head = next
	case 1:
		prev.next = nil
		c.tail = prev
	default:
		c.head = nil
		c.tail = nil
	}

	ele.prev, ele.next = nil, nil
	c.length--
}

func (c *queue[K, V]) MoveToBack(ele *Element[K, V]) {
	if !c.enabled {
		return
	}

	c.Delete(ele)
	c.PushBack(ele)
}

func (c *queue[K, V]) Keys() []K {
	var keys = make([]K, 0, c.Len())
	for i := c.head; i != nil; i = i.next {
		keys = append(keys, i.Key)
	}
	return keys
}
