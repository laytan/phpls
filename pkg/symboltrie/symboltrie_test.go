package symboltrie

import (
	"testing"
)

type testNode struct {
	Key    string
	Detail string
}

func TestSymbolTrieSearchPrefix(t *testing.T) {
	trie := New[*testNode]()

	trie.Put("Test", &testNode{"Test", "Test One"})
	trie.Put("Test", &testNode{"Test", "Test Two"})
	trie.Put("Tie", &testNode{"Tie", "Test Three"})

	results := trie.SearchPrefix("", 3)

	var foundOne, foundTwo, foundThree bool
	for _, result := range results {
		if result.Value.Detail == "Test One" {
			foundOne = true
		}

		if result.Value.Detail == "Test Two" {
			foundTwo = true
		}

		if result.Value.Detail == "Test Three" {
			foundThree = true
		}
	}

	if !foundOne || !foundTwo || !foundThree {
		t.Errorf("%v %v %v should all be true", foundOne, foundTwo, foundThree)
	}
}
