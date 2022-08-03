package traversers

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/symbol"
	"github.com/laytan/elephp/pkg/symboltrie"
)

type TrieNode struct {
	Path      string
	Namespace string
	Symbol    symbol.Symbol
}

// Symbol implements ir.Visitor.
type Symbol struct {
	trie             *symboltrie.Trie[*TrieNode]
	path             string
	currentNamespace string
}

func NewSymbol(trie *symboltrie.Trie[*TrieNode]) *Symbol {
	return &Symbol{trie: trie}
}

func (s *Symbol) SetPath(path string) {
	s.path = path
	s.currentNamespace = ""
}

func (s *Symbol) EnterNode(node ir.Node) bool {
	switch typedNode := node.(type) {

	case *ir.NamespaceStmt:
		if typedNode.NamespaceName != nil {
			s.currentNamespace = typedNode.NamespaceName.Value
		}

		return true

	case *ir.FunctionStmt:
		node := s.newTrieNode(symbol.NewFunction(typedNode))
		s.trie.Put(node.Symbol.Identifier(), node)
		return false

	case *ir.ClassStmt:
		node := s.newTrieNode(symbol.NewClassLikeClass(typedNode))
		s.trie.Put(node.Symbol.Identifier(), node)

		return false

	case *ir.InterfaceStmt:
		node := s.newTrieNode(symbol.NewClassLikeInterface(typedNode))
		s.trie.Put(node.Symbol.Identifier(), node)

		return false

	case *ir.TraitStmt:
		node := s.newTrieNode(symbol.NewClassLikeTrait(typedNode))
		s.trie.Put(node.Symbol.Identifier(), node)

		return false

	default:
		return true

	}
}

func (s *Symbol) LeaveNode(ir.Node) {}

func (s *Symbol) newTrieNode(node symbol.Symbol) *TrieNode {
	return &TrieNode{
		Path:      s.path,
		Namespace: s.currentNamespace,
		Symbol:    node,
	}
}
