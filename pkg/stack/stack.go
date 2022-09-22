package stack

import (
	"fmt"
	"strings"
	"sync"

	"appliedgo.net/what"
)

type node[V any] struct {
	value V
	next  *node[V]
	prev  *node[V]
}

type Stack[V any] struct {
	mu sync.Mutex

	tail *node[V]
}

func New[V any]() *Stack[V] {
	return &Stack[V]{}
}

func (s *Stack[V]) Push(value V) {
	s.mu.Lock()
	defer s.mu.Unlock()

	node := &node[V]{value: value, prev: s.tail}
	s.tail = node
}

func (s *Stack[V]) Pop() V {
	s.mu.Lock()
	defer s.mu.Unlock()

	node := s.tail
	if node == nil {
		var v V
		return v
	}

	if s.tail.prev == nil {
		s.tail = nil
		return node.value
	}

	s.tail.prev.next = nil
	s.tail = s.tail.prev
	return node.value
}

func (s *Stack[V]) Peek() V {
	if s.tail == nil {
		var v V
		return v
	}

	return s.tail.value
}

func (s *Stack[V]) String() string {
	out := []string{}

	t := s.tail
	for curr := t; curr != nil; curr = curr.next {
		what.Happens("%T", curr.value)
		out = append(out, fmt.Sprintf("%v", curr.value))
	}

	return strings.Join(out, "-")
}
