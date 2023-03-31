package queue_test

import (
	"testing"

	"github.com/laytan/elephp/pkg/queue"
	"github.com/stretchr/testify/require"
)

func TestQueue(t *testing.T) {
	t.Parallel()

	q := queue.New[int]()

	v := q.Dequeue()
	require.Equal(t, v, 0)

	q.Enqueue(1)
	v = q.Dequeue()
	require.Equal(t, v, 1)

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
		require.Equal(t, v, item)
	}
}
