package symboltrie_test

import (
	"testing"

	"github.com/laytan/phpls/internal/symboltrie"
	"github.com/laytan/phpls/pkg/fqn"
	"github.com/stretchr/testify/require"
)

func TestSymboltrie(t *testing.T) {
	t.Parallel()

	trie := symboltrie.New[string]()

	trie.Put(fqn.New("\\Test"), "Test")
	res := trie.FullSearch(fqn.New("\\Test"))
	require.Len(t, res, 1)
	require.Equal(t, res[0], "Test")

	trie.Put(fqn.New("\\Test"), "Test2")
	res = trie.FullSearch(fqn.New("\\Test"))
	require.Len(t, res, 2)
	require.Equal(t, res, []string{"Test", "Test2"})

	trie.Delete(fqn.New("\\Test"), func(s string) bool { return s == "Test" })
	res = trie.FullSearch(fqn.New("\\Test"))
	require.Len(t, res, 1)
	require.Equal(t, res[0], "Test2")

	trie.Put(fqn.New("\\array_map"), "array_map")
	trie.Put(fqn.New("\\Foo\\array_pop"), "array_pop")
	res = trie.NameSearch("array", -1)
	require.Len(t, res, 2)
	require.Equal(t, res, []string{"array_map", "array_pop"})

	res = trie.NameSearch("T", -1)
	require.Len(t, res, 1)
	require.Equal(t, res[0], "Test2")
}
