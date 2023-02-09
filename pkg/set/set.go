package set

type Set[V comparable] struct {
	list map[V]struct{} // empty structs occupy 0 memory
}

func (s *Set[V]) Has(v V) bool {
	_, ok := s.list[v]
	return ok
}

func (s *Set[V]) Add(v V) {
	s.list[v] = struct{}{}
}

func (s *Set[V]) Remove(v V) {
	delete(s.list, v)
}

func (s *Set[V]) Clear() {
	s.list = make(map[V]struct{})
}

func (s *Set[V]) Size() int {
	return len(s.list)
}

func New[V comparable]() *Set[V] {
	s := &Set[V]{}
	s.list = make(map[V]struct{})
	return s
}

func NewWithCap[V comparable](capacity int) *Set[V] {
	s := &Set[V]{}
	s.list = make(map[V]struct{}, capacity)
	return s
}

func NewFromSlice[V comparable](slice []V) *Set[V] {
	s := NewWithCap[V](len(slice))
	s.AddMulti(slice...)
	return s
}

func (s *Set[V]) AddMulti(list ...V) {
	for _, v := range list {
		s.Add(v)
	}
}

type FilterFunc[V comparable] func(v V) bool

// Filter returns a subset, that contains only the values that satisfies the given predicate p.
func (s *Set[V]) Filter(p FilterFunc[V]) *Set[V] {
	res := New[V]()
	for v := range s.list {
		if !p(v) {
			continue
		}
		res.Add(v)
	}
	return res
}

func (s *Set[V]) Union(s2 *Set[V]) *Set[V] {
	res := New[V]()
	for v := range s.list {
		res.Add(v)
	}

	for v := range s2.list {
		res.Add(v)
	}
	return res
}

func (s *Set[V]) Intersect(s2 *Set[V]) *Set[V] {
	res := New[V]()
	for v := range s.list {
		if !s2.Has(v) {
			continue
		}
		res.Add(v)
	}
	return res
}

// Difference returns the subset from s, that doesn't exists in s2 (param).
func (s *Set[V]) Difference(s2 *Set[V]) *Set[V] {
	res := New[V]()
	for v := range s.list {
		if s2.Has(v) {
			continue
		}
		res.Add(v)
	}
	return res
}

func (s *Set[V]) Slice() []V {
	slice := make([]V, 0, len(s.list))
	for v := range s.list {
		slice = append(slice, v)
	}

	return slice
}

// / Returns a generator channel for all the values in the set.
// / Useful in range loops.
// /
// / NOTE: You have to exhaust the channel, if you don't it leaves a dangling
// / goroutine.
func (s *Set[V]) Iterator() <-chan V {
	channel := make(chan V)

	go func() {
		for v := range s.list {
			channel <- v
		}

		close(channel)
	}()

	return channel
}
