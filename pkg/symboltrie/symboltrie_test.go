package symboltrie_test

import (
	"fmt"
	"testing"

	"appliedgo.net/what"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/symboltrie"
	"github.com/matryer/is"
)

type testNode struct {
	Key    string
	Detail string
}

func TestSymbolTrieSearchPrefix(t *testing.T) {
	is := is.New(t)
	t.Parallel()
	trie := symboltrie.New[*testNode]()

	trie.Put(fqn.New("\\Test"), &testNode{"Test", "Test One"})
	trie.Put(fqn.New("\\Test\\Test"), &testNode{"Test", "Test Two"})
	trie.Put(fqn.New("\\Test\\Except"), &testNode{"Except", "Test Three"})

	n, ok := trie.SearchExact(fqn.New("\\Test"))
	is.True(ok)
	is.Equal(n.Detail, "Test One")

	n, ok = trie.SearchExact(fqn.New("\\Test\\Test"))
	is.True(ok)
	is.Equal(n.Key, "Test")

	ns := trie.SearchNames("Test", 0)
	is.Equal(len(ns), 2)

	ns = trie.SearchNames("Te", 2)
	is.Equal(len(ns), 2)

	ns = trie.SearchNames("Te", 1)
	is.Equal(len(ns), 1)

	ns = trie.SearchNames("Bla", 0)
	is.Equal(len(ns), 0)

	trie.Delete(fqn.New("\\Test\\Test"))

	_, ok = trie.SearchExact(fqn.New("\\Test\\Test"))
	is.Equal(ok, false)

	ns = trie.SearchNames("Test", 0)
	is.Equal(len(ns), 1)

	itrie := symboltrie.New[int]()
	itrie.Put(fqn.New("\\Test\\Exceptions\\Test"), 3)
	itrie.Put(fqn.New("\\Test"), 5)
	itrie.Put(fqn.New("\\InvalidArgumentException"), 4)
	itrie.Put(fqn.New("\\Drupal"), 2)
	itrie.Put(fqn.New("\\Drupal\\Exceptions\\BaseException"), 6)
	itrie.Put(fqn.New("\\Drupal\\Exceptions\\InvalidException"), 1)

	cancel := make(chan struct{})
	i := 0
	for entry := range itrie.Iterator(cancel) {
		i++
		what.Happens(fmt.Sprint(i))
		is.True(i < 4)
		is.True(i == entry.Value)

		if i == 3 {
			cancel <- struct{}{}
		}
	}

	i = 0
	for range itrie.Iterator(nil) {
		i++
	}

	is.Equal(i, 6)
}
