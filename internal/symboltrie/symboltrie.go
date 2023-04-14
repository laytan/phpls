package symboltrie

import (
	"sync"

	"github.com/laytan/elephp/pkg/fqn"
)

const (
	separator = '\\'
)

type node[T comparable] struct {
	value  []T
	childs map[rune]*node[T]
}

func newNode[T comparable]() *node[T] {
	return &node[T]{childs: make(map[rune]*node[T])}
}

type Trie[T comparable] struct {
	root       node[T]
	namespaces []*node[T]
	mu         sync.Mutex
}

func New[T comparable]() *Trie[T] {
	root := newNode[T]()
	return &Trie[T]{
		root:       *root,
		namespaces: []*node[T]{root},
	}
}

func (t *Trie[T]) Put(fqn *fqn.FQN, value T) {
	t.mu.Lock()
	defer t.mu.Unlock()

	target := &t.root
	uncharted := true
	fqnS := fqn.String()
	for _, ch := range fqnS[1:] {
		if uncharted {
			if tn, ok := target.childs[ch]; ok {
				target = tn
				continue
			}

			uncharted = false
		}

		nt := newNode[T]()
		target.childs[ch] = nt
		target = nt
		if ch == separator {
			t.namespaces = append(t.namespaces, target)
		}
	}

	target.value = append(target.value, value)
}

func DelAll[T comparable](T) bool { return true }

// This only deletes the leaf node
// and does not make an effort to delete nodes that branch backwards which after deletion would have no use.
// Because that would not happen much in practice, so ok to leave out.
func (t *Trie[T]) Delete(fqn *fqn.FQN, confirm func(T) bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	target := &t.root
	fqnS := fqn.String()
	for _, ch := range fqnS[1:] {
		if tn, ok := target.childs[ch]; ok {
			target = tn
			continue
		}

		return
	}

	temp := target.value[:0]
	for _, x := range target.value {
		if !confirm(x) {
			temp = append(temp, x)
		}
	}
	target.value = temp
}

func (t *Trie[T]) FullSearch(fqn *fqn.FQN) []T {
	t.mu.Lock()
	defer t.mu.Unlock()

	target := &t.root
	fqnS := fqn.String()
	for _, ch := range fqnS[1:] {
		if tn, ok := target.childs[ch]; ok {
			target = tn
			continue
		}

		return nil
	}

	return target.value
}

func (t *Trie[T]) FqnSearch(prefix string, max int) []T {
	t.mu.Lock()
	defer t.mu.Unlock()

	target := &t.root
	for _, ch := range prefix[1:] {
		if tn, ok := target.childs[ch]; ok {
			target = tn
			continue
		}

		return nil
	}

	return unpack(target, 0, max)
}

func (t *Trie[T]) NameSearch(prefix string, max int) (res []T) {
	t.mu.Lock()
	defer t.mu.Unlock()

Namespaces:
	for _, ns := range t.namespaces {
		target := ns
		for _, ch := range prefix {
			if tn, ok := target.childs[ch]; ok {
				target = tn
				continue
			}

			continue Namespaces
		}

		res = append(res, unpack(target, separator, max)...)
		if max > 0 && len(res) >= max {
			return res
		}
	}

	return res
}

func unpack[T comparable](n *node[T], abortOn rune, max int) (res []T) {
	res = append(res, n.value...)
	for k, v := range n.childs {
		if k == abortOn {
			continue
		}

		res = append(res, unpack(v, abortOn, max)...)
		if max > 0 && len(res) >= max {
			return res
		}
	}

	return res
}
