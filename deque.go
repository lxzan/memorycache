package memorycache

import (
	"github.com/lxzan/dao/stack"
)

const null = 0

type (
	pointer uint32

	// Deque 可以不使用New函数, 声明为值类型自动初始化
	deque[K comparable, V any] struct {
		head, tail pointer              // 头尾指针
		length     int                  // 长度
		stack      stack.Stack[pointer] // 回收站
		elements   []Element[K, V]      // 元素列表
		template   Element[K, V]        // 空值模板
	}
)

func (c pointer) IsNil() bool {
	return c == null
}

func newDeque[K comparable, V any](capacity int) *deque[K, V] {
	return &deque[K, V]{elements: make([]Element[K, V], 1, 1+capacity)}
}

func (c *deque[K, V]) Get(addr pointer) *Element[K, V] {
	if addr > 0 {
		return &(c.elements[addr])
	}
	return nil
}

// getElement 追加元素一定要先调用此方法, 因为追加可能会造成扩容, 地址发生变化!!!
func (c *deque[K, V]) getElement() *Element[K, V] {
	if c.stack.Len() > 0 {
		addr := c.stack.Pop()
		v := c.Get(addr)
		v.addr = addr
		return v
	}

	addr := pointer(len(c.elements))
	c.elements = append(c.elements, c.template)
	v := c.Get(addr)
	v.addr = addr
	return v
}

func (c *deque[K, V]) putElement(ele *Element[K, V]) {
	c.stack.Push(ele.addr)
	*ele = c.template
}

func (c *deque[K, V]) autoReset() {
	c.head, c.tail, c.length = null, null, 0
	c.stack = c.stack[:0]
	c.elements = c.elements[:1]
}

func (c *deque[K, V]) Len() int {
	return c.length
}

func (c *deque[K, V]) Front() *Element[K, V] {
	return c.Get(c.head)
}

func (c *deque[K, V]) Back() *Element[K, V] {
	return c.Get(c.tail)
}

func (c *deque[K, V]) PushBack() *Element[K, V] {
	ele := c.getElement()
	c.doPushBack(ele)
	return ele
}

func (c *deque[K, V]) doPushBack(ele *Element[K, V]) {
	c.length++

	if c.tail.IsNil() {
		c.head, c.tail = ele.addr, ele.addr
		return
	}

	tail := c.Get(c.tail)
	tail.next = ele.addr
	ele.prev = tail.addr
	c.tail = ele.addr
}

func (c *deque[K, V]) PopFront() (value Element[K, V]) {
	if ele := c.Front(); ele != nil {
		value = *ele
		c.doRemove(ele)
		c.putElement(ele)
		if c.length == 0 {
			c.autoReset()
		}
	}
	return value
}

func (c *deque[K, V]) MoveToBack(addr pointer) {
	if ele := c.Get(addr); ele != nil {
		c.doRemove(ele)
		ele.prev, ele.next = null, null
		c.doPushBack(ele)
	}
}

func (c *deque[K, V]) Remove(addr pointer) {
	if ele := c.Get(addr); ele != nil {
		c.doRemove(ele)
		c.putElement(ele)
		if c.length == 0 {
			c.autoReset()
		}
	}
}

func (c *deque[K, V]) doRemove(ele *Element[K, V]) {
	var prev, next *Element[K, V] = nil, nil
	var state = 0
	if !ele.prev.IsNil() {
		prev = c.Get(ele.prev)
		state += 1
	}
	if !ele.next.IsNil() {
		next = c.Get(ele.next)
		state += 2
	}

	c.length--
	switch state {
	case 3:
		prev.next = next.addr
		next.prev = prev.addr
	case 2:
		next.prev = null
		c.head = next.addr
	case 1:
		prev.next = null
		c.tail = prev.addr
	default:
		c.head = null
		c.tail = null
	}
}

func (c *deque[K, V]) Range(f func(ele *Element[K, V]) bool) {
	for i := c.Get(c.head); i != nil; i = c.Get(i.next) {
		if !f(i) {
			break
		}
	}
}
