package queue

import (
	"sync"
)

type node[V any] struct {
	next  *node[V]
	value V
}

// A simple FIFO Queue.
type Queue[V any] struct {
	mu sync.Mutex

	head *node[V]
	tail *node[V]
}

func New[V any]() *Queue[V] {
	return &Queue[V]{}
}

func (q *Queue[V]) Enqueue(value V) {
	q.mu.Lock()
	defer q.mu.Unlock()

	node := &node[V]{value: value}

	if q.head == nil {
		q.head = node
		return
	}

	if q.tail == nil {
		q.tail = node
		q.head.next = q.tail
		return
	}

	q.tail.next = node
	q.tail = node
}

func (q *Queue[V]) Dequeue() V {
	q.mu.Lock()
	defer q.mu.Unlock()

	node := q.head
	if node == nil {
		var v V
		return v
	}

	q.head = node.next
	return node.value
}
