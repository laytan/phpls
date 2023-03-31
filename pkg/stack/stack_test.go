package stack_test

import (
	"testing"

	"github.com/laytan/elephp/pkg/stack"
	"github.com/stretchr/testify/require"
)

func TestQueue(t *testing.T) {
	t.Parallel()

	q := stack.New[int]()

	v := q.Pop()
	require.Equal(t, v, 0)

	q.Push(1)
	require.Equal(t, q.Length(), 1, "equal length")

	v = q.Pop()
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
	for i, item := range items {
		q.Push(item)
		require.Equal(t, q.Length(), i+1)
	}

	for i := len(items) - 1; i >= 0; i-- {
		v := q.Pop()
		require.Equal(t, v, items[i])
	}
}
