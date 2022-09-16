package symboltrie

import (
	"strings"
	"sync"

	"github.com/shivamMg/trie"
)

// A wrapper around a lower level trie, adding generics and allows duplicate keys.
type Trie[T any] struct {
	trie             *trie.Trie
	mu               sync.Mutex
	Size             uint
	MaxDuplicates    uint
	MaxDuplicatesKey string
}

func New[T any]() *Trie[T] {
	return &Trie[T]{
		trie: trie.New(),
	}
}

// Searches the trie for exact matches of the given key.
// This can return multiple results if there are duplicates.
func (s *Trie[T]) SearchExact(key string) []T {
	opts := []func(*trie.SearchOptions){trie.WithExactKey()}
	result := s.trie.Search(strings.Split(key, ""), opts...)

	if len(result.Results) == 0 {
		return nil
	}

	trieNode := result.Results[0].Value.(*trieNode[T])
	return trieNode.Nodes
}

type SearchResult[T any] struct {
	Key   string
	Value T
}

func SearchResultKeys[T any](results []*SearchResult[T]) []string {
	keys := make([]string, 0, len(results))
	for _, result := range results {
		keys = append(keys, result.Key)
	}

	return keys
}

func (s *Trie[T]) SearchPrefix(key string, maxResults uint) []*SearchResult[T] {
	opts := make([]func(*trie.SearchOptions), 0, 1)
	if maxResults > 0 {
		opts = append(opts, trie.WithMaxResults(int(maxResults)))
	}

	result := s.trie.Search(strings.Split(key, ""), opts...)

	flatResults := make([]*SearchResult[T], 0, len(result.Results))
	for _, res := range result.Results {
		trieNode := res.Value.(*trieNode[T])
		key := strings.Join(res.Key, "")

		for _, inRes := range trieNode.Nodes {
			flatResults = append(flatResults, &SearchResult[T]{
				Key:   key,
				Value: inRes,
			})
		}
	}

	return flatResults
}

// Puts the node at key in the trie, if this is a duplicate it stores it alongside it.
func (s *Trie[T]) Put(key string, node T) {
	s.mu.Lock()
	defer s.mu.Unlock()

	splitKey := strings.Split(key, "")
	opts := []func(*trie.SearchOptions){trie.WithExactKey()}
	result := s.trie.Search(splitKey, opts...)

	if len(result.Results) == 1 {
		trieNode := result.Results[0].Value.(*trieNode[T])
		trieNode.Nodes = append(trieNode.Nodes, node)

		if uint(len(trieNode.Nodes)) > s.MaxDuplicates {
			s.MaxDuplicates = uint(len(trieNode.Nodes))
			s.MaxDuplicatesKey = key
		}

		s.Size++
		return
	}

	s.trie.Put(splitKey, newTrieNode(node))
	s.Size++
}

// Calls predicate with all the nodes that exactly match the given key,
// Deleting the first node that returns true.
//
// Returns whether something has been removed.
func (s *Trie[T]) Delete(key string, predicate func(T) bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	splitKey := strings.Split(key, "")
	opts := []func(*trie.SearchOptions){trie.WithExactKey()}
	result := s.trie.Search(splitKey, opts...)

	if len(result.Results) == 0 {
		return false
	}

	trieNode := result.Results[0].Value.(*trieNode[T])
	for i, result := range trieNode.Nodes {
		if !predicate(result) {
			continue
		}

		// Removes the node.
		trieNode.Nodes[i] = trieNode.Nodes[len(trieNode.Nodes)-1]
		trieNode.Nodes = trieNode.Nodes[:len(trieNode.Nodes)-1]

		// If the slice is now empty, remove it completely.
		if len(trieNode.Nodes) == 0 {
			s.trie.Delete(splitKey)
		}

		s.Size--
		return true
	}

	return false
}

func (s *Trie[T]) Print() {
	s.trie.Root().Print()
}

type trieNode[T any] struct {
	Nodes []T
}

func newTrieNode[T any](node T) *trieNode[T] {
	return &trieNode[T]{
		Nodes: []T{node},
	}
}
