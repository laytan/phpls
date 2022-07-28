package traversers

import (
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/shivamMg/trie"
)

type TrieNode struct {
	Path      string
	Namespace string
	Node      ir.Node
}

func NewSymbol(trie *trie.Trie, path string) *Symbol {
	return &Symbol{trie: trie, path: path}
}

// Symbol implements ir.Visitor.
type Symbol struct {
	trie             *trie.Trie
	path             string
	currentNamespace string
}

func (s *Symbol) EnterNode(node ir.Node) bool {
	switch typedNode := node.(type) {

	case *ir.NamespaceStmt:
		if typedNode.NamespaceName != nil {
			s.currentNamespace = typedNode.NamespaceName.Value
		}

		return true

	case *ir.FunctionStmt:
		if typedNode.FunctionName != nil {
			s.trie.Put(
				strings.Split(typedNode.FunctionName.Value, ""),
				s.newTrieNode(node),
			)
		}

		return false

	default:
		return true

	}
}

func (s *Symbol) LeaveNode(ir.Node) {}

func (s *Symbol) newTrieNode(node ir.Node) *TrieNode {
	return &TrieNode{
		Path:      s.path,
		Namespace: s.currentNamespace,
		Node:      node,
	}
}
