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
	trie.Put("Tie", &testNode{"Tie", "Test Four"})
	trie.Put("Tiet", &testNode{"Tiet", "Test Five"})

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

	if trie.Size != 5 {
		t.Errorf("Trie size, expected 5, got %d", trie.Size)
	}

	trie.Delete("Test", func(tn *testNode) bool {
		return tn.Detail == "Test Two"
	})

	if trie.Size != 4 {
		t.Errorf("Trie size, expected 4, got %d", trie.Size)
	}

	trie.Delete("Tiet", func(tn *testNode) bool { return true })

	if trie.Size != 3 {
		t.Errorf("Trie size, expected 3, got %d", trie.Size)
	}
}
