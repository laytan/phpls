package stack

import (
	"testing"

	"github.com/matryer/is"
)

func TestQueue(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	q := New[int]()

	v := q.Pop()
	is.Equal(v, 0)

	q.Push(1)
	v = q.Pop()
	is.Equal(v, 1)

	items := []int{
		1,
		2,
		3,
		4,
		5,
		6,
		10,
		83,
	}
	for _, item := range items {
		q.Push(item)
	}

	for i := len(items) - 1; i >= 0; i-- {
		v := q.Pop()
		is.Equal(v, items[i])
	}
}
