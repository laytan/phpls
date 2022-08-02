package project

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/traversers"
)

func (p *Project) FindNodeInTrie(
	fqn *traversers.FQN,
	kind ir.NodeKind,
) *traversers.TrieNode {
	results := p.symbolTrie.SearchExact(fqn.Name())
	for _, res := range results {
		if res.Namespace != fqn.Namespace() {
			continue
		}

		if res.Symbol.NodeKind() != kind {
			continue
		}

		return res
	}

	return nil
}

func (p *Project) FindFileInTrie(fqn *traversers.FQN, kind ir.NodeKind) *File {
	node := p.FindNodeInTrie(fqn, kind)
	if node == nil {
		return nil
	}

	return p.GetFile(node.Path)
}

func (p *Project) FindFileAndSymbolInTrie(
	fqn *traversers.FQN,
	kind ir.NodeKind,
) (*File, *traversers.TrieNode) {
	node := p.FindNodeInTrie(fqn, kind)
	if node == nil {
		return nil, nil
	}

	file := p.GetFile(node.Path)

	return file, node
}
