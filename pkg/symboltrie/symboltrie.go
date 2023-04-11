package symboltrie

import (
	"strings"
	"sync"

	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/functional"
	"github.com/shivamMg/trie"
)

type Trie[T any] struct {
	fqns     *trie.Trie
	fqnsLock sync.RWMutex

	// TODO: having 2 tries here is a bit of a waste of memory,
	// also this shivaMg/trie uses a lot of memory.
	//
	// TODO: Is there a datastructure or modification to a trie,
	// where we can start from the last '/' in the chain?
	// can probably internally keep a list of pointers to all the last '/'.
	// but that requires a custom trie implementation most likely.
	names     *trie.Trie
	namesLock sync.Mutex
}

func New[T any]() *Trie[T] {
	return &Trie[T]{
		fqns:  trie.New(),
		names: trie.New(),
	}
}

func (t *Trie[T]) Put(key *fqn.FQN, value T) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		t.namesLock.Lock()
		defer t.namesLock.Unlock()
		defer wg.Done()

		name := strings.Split(key.Name(), "")

		opts := []func(*trie.SearchOptions){trie.WithExactKey()}
		result := t.names.Search(name, opts...)

		// The leaf trie can have duplicates, so we store slices.
		// If there already is a slice, put it in there, otherwise create a new.
		if len(result.Results) == 0 {
			newNode := make(map[string]T)
			newNode[key.String()] = value
			t.names.Put(name, newNode)
		} else {
			m := result.Results[0].Value.(map[string]T)
			m[key.String()] = value
		}
	}()

	go func() {
		t.fqnsLock.Lock()
		defer t.fqnsLock.Unlock()
		defer wg.Done()

		t.fqns.Put(strings.Split(key.String(), ""), value)
	}()

	wg.Wait()
}

func (t *Trie[T]) SearchExact(key *fqn.FQN) (T, bool) {
	t.fqnsLock.RLock()
	defer t.fqnsLock.RUnlock()

	r := t.fqns.Search(strings.Split(key.String(), ""), trie.WithExactKey())
	if len(r.Results) > 0 {
		res := r.Results[0].Value.(T)
		return res, true
	}

	var defaultT T
	return defaultT, false
}

func (t *Trie[T]) SearchNames(prefix string, maxResults int) []T {
	opts := make([]func(*trie.SearchOptions), 0, 1)
	if maxResults > 0 {
		opts = append(opts, trie.WithMaxResults(int(maxResults)))
	}

	results := t.names.Search(strings.Split(prefix, ""), opts...)
	flatResults := make([]T, 0, len(results.Results))
	for _, result := range results.Results {
		for _, innerResult := range result.Value.(map[string]T) {
			flatResults = append(flatResults, innerResult)
			if maxResults > 0 && len(flatResults) >= maxResults {
				break
			}
		}
	}

	return flatResults
}

func (t *Trie[T]) SearchFqns(prefix string, maxResults int) []T {
	opts := make([]func(*trie.SearchOptions), 0, 1)
	if maxResults > 0 {
		opts = append(opts, trie.WithMaxResults(int(maxResults)))
	}

	r := t.fqns.Search(strings.Split(prefix, ""), opts...)
	return functional.Map(r.Results, func(result *trie.SearchResult) T {
		return result.Value.(T)
	})
}

func (t *Trie[T]) Delete(key *fqn.FQN) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		t.namesLock.Lock()
		defer t.namesLock.Unlock()
		defer wg.Done()

		nameKey := strings.Split(key.Name(), "")
		opts := []func(*trie.SearchOptions){trie.WithExactKey()}
		results := t.names.Search(nameKey, opts...)

		if len(results.Results) > 0 {
			currNames := results.Results[0].Value.(map[string]T)
			delete(currNames, key.String())

			if len(currNames) == 0 {
				t.names.Delete(nameKey)
			}
		}
	}()

	go func() {
		t.fqnsLock.Lock()
		defer t.fqnsLock.Unlock()
		defer wg.Done()

		t.fqns.Delete(strings.Split(key.String(), ""))
	}()

	wg.Wait()
}

type TriePair[T any] struct {
	Key   []byte
	Value T
}
