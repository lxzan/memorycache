package memorycache

import (
	"container/list"
	"math/rand"
	"testing"

	"github.com/lxzan/memorycache/internal/utils"

	"github.com/stretchr/testify/assert"
)

func validate(q *deque[int, int]) bool {
	var sum = 0
	for i := q.Get(q.head); i != nil; i = q.Get(i.next) {
		sum++
		next := q.Get(i.next)
		if next == nil {
			continue
		}
		if i.next != next.addr {
			return false
		}
		if next.prev != i.addr {
			return false
		}
	}

	if q.Len() != sum {
		return false
	}

	if head := q.Front(); head != nil {
		if head.prev != 0 {
			return false
		}
	}

	if tail := q.Back(); tail != nil {
		if tail.next != 0 {
			return false
		}
	}

	if q.Len() == 1 && q.Front().Value != q.Back().Value {
		return false
	}

	return true
}

func TestQueue_Random(t *testing.T) {
	var count = 10000
	var q = newDeque[int, int](0)
	var linkedlist = list.New()
	for i := 0; i < count; i++ {
		var flag = rand.Intn(7)
		var val = rand.Int()
		switch flag {
		case 0, 1, 2, 3:
			ele := q.PushBack()
			ele.Value = val
			linkedlist.PushBack(val)
		case 4:
			if q.Len() > 0 {
				q.PopFront()
				linkedlist.Remove(linkedlist.Front())
			}
		case 5:
			var n = rand.Intn(10)
			var index = 0
			for iter := q.Front(); iter != nil; iter = q.Get(iter.next) {
				index++
				if index >= n {
					q.MoveToBack(iter.addr)
					break
				}
			}

			index = 0
			for iter := linkedlist.Front(); iter != nil; iter = iter.Next() {
				index++
				if index >= n {
					linkedlist.MoveToBack(iter)
					break
				}
			}

		case 6:
			var n = rand.Intn(10)
			var index = 0
			for iter := q.Front(); iter != nil; iter = q.Get(iter.next) {
				index++
				if index >= n {
					q.Remove(iter.addr)
					break
				}
			}

			index = 0
			for iter := linkedlist.Front(); iter != nil; iter = iter.Next() {
				index++
				if index >= n {
					linkedlist.Remove(iter)
					break
				}
			}
		default:

		}
	}

	assert.True(t, validate(q))
	for i := linkedlist.Front(); i != nil; i = i.Next() {
		var ele = q.PopFront()
		assert.Equal(t, i.Value, ele.Value)
	}
}

func TestDeque_Range(t *testing.T) {
	var q = newDeque[string, int](8)
	var push = func(values ...int) {
		for _, v := range values {
			q.PushBack().Value = v
		}
	}
	push(1, 3, 5, 7, 9)

	{
		var arr []int
		q.Range(func(ele *Element[string, int]) bool {
			arr = append(arr, ele.Value)
			return true
		})
		assert.True(t, utils.IsSameSlice(arr, []int{1, 3, 5, 7, 9}))
	}

	{
		var arr []int
		q.Range(func(ele *Element[string, int]) bool {
			arr = append(arr, ele.Value)
			return len(arr) < 3
		})
		assert.True(t, utils.IsSameSlice(arr, []int{1, 3, 5}))
	}
}
