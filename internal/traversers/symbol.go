package traversers

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/symboltrie"
)

type TrieNode struct {
	Path      string
	Namespace string
	Node      ir.Node
}

func NewSymbol(trie *symboltrie.Trie[*TrieNode]) *Symbol {
	return &Symbol{trie: trie}
}

// Symbol implements ir.Visitor.
type Symbol struct {
	trie             *symboltrie.Trie[*TrieNode]
	path             string
	currentNamespace string
}

func (s *Symbol) SetPath(path string) {
	s.path = path
	s.currentNamespace = ""
}

func (s *Symbol) EnterNode(node ir.Node) bool {
	switch typedNode := node.(type) {

	// TODO: abstract away getting a node's name, like with ir.GetPosition.

	case *ir.NamespaceStmt:
		if typedNode.NamespaceName != nil {
			s.currentNamespace = typedNode.NamespaceName.Value
		}

		return true

	case *ir.FunctionStmt:
		if typedNode.FunctionName != nil {
			s.trie.Put(typedNode.FunctionName.Value, s.newTrieNode(node))
		}

		return false

	case *ir.ClassStmt:
		s.trie.Put(typedNode.ClassName.Value, s.newTrieNode(node))

		return false

	case *ir.InterfaceStmt:
		s.trie.Put(typedNode.InterfaceName.Value, s.newTrieNode(node))

		return false

	case *ir.TraitStmt:
		s.trie.Put(typedNode.TraitName.Value, s.newTrieNode(node))

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
