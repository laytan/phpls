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

func (t *TrieNode) Fqn() string {
	return `\` + t.Namespace + `\` + t.Symbol.Identifier()
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

	case *ir.FunctionStmt, *ir.ClassStmt, *ir.InterfaceStmt, *ir.TraitStmt:
		node := s.newTrieNode(symbol.New(typedNode))
		s.trie.Put(node.Symbol.Identifier(), node)
		return true

	case *ir.FunctionCallExpr:
		if fn, ok := typedNode.Function.(*ir.Name); ok && fn.Value == "define" {
			node := s.newGlobalTrieNode(symbol.New(typedNode))
			s.trie.Put(node.Symbol.Identifier(), node)
		}

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

func (s *Symbol) newGlobalTrieNode(node symbol.Symbol) *TrieNode {
	return &TrieNode{
		Path:      s.path,
		Namespace: "",
		Symbol:    node,
	}
}
