package queue

import (
	"testing"

	"github.com/matryer/is"
)

func TestQueue(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	q := New[int]()

	v := q.Dequeue()
	is.Equal(v, 0)

	q.Enqueue(1)
	v = q.Dequeue()
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
		q.Enqueue(item)
	}

	for _, item := range items {
		v := q.Dequeue()
		is.Equal(v, item)
	}
}
