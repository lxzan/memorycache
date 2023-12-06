package memorycache

import "github.com/lxzan/memorycache/internal/utils"

type queue[K comparable, V any] struct {
	length int
	head   *Element[K, V]
	tail   *Element[K, V]
}

func (c *queue[K, V]) Len() int {
	return c.length
}

func (c *queue[K, V]) Front() *Element[K, V] {
	return c.head
}

func (c *queue[K, V]) PushBack(ele *Element[K, V]) {
	if c.length > 0 {
		c.tail.next = ele
		ele.prev = c.tail
		c.tail = ele
	} else {
		c.head = ele
		c.tail = ele
	}
	c.length++
}

func (c *queue[K, V]) Pop() *Element[K, V] {
	if c.length == 0 {
		return nil
	}
	var v = c.head
	c.head = c.head.next
	v.next = nil
	if c.head != nil {
		c.head.prev = nil
	}
	if c.length == 1 {
		c.tail = nil
	}
	c.length--
	return v
}

// Delete it's safe delete in loop
func (c *queue[K, V]) Delete(iter *Element[K, V]) {
	var prev = iter.prev
	var next = iter.next
	var a = utils.SelectValue(prev == nil, 0, 1)
	var b = utils.SelectValue(next == nil, 0, 2)
	switch a + b {
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
	c.length--
}

func (c *queue[K, V]) MoveToBack(ele *Element[K, V]) {
	c.Delete(ele)
	ele.prev, ele.next = nil, nil

	if c.length > 0 {
		c.tail.next = ele
		ele.prev = c.tail
		c.tail = ele
	} else {
		c.head = ele
		c.tail = ele
	}
	c.length++
}

func (c *queue[K, V]) Keys() []K {
	var keys = make([]K, 0, c.length)
	for i := c.head; i != nil; i = i.next {
		keys = append(keys, i.Key)
	}
	return keys
}
