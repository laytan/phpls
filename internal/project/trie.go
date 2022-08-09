package project

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/traversers"
	"github.com/laytan/elephp/pkg/typer"
)

func (p *Project) FindNodeInTrie(
	fqn *typer.FQN,
	kind ir.NodeKind,
) *traversers.TrieNode {
	return p.FindNodeInTrieMultiKinds(fqn, []ir.NodeKind{kind})
}

func (p *Project) FindNodeInTrieMultiKinds(
	fqn *typer.FQN,
	kinds []ir.NodeKind,
) *traversers.TrieNode {
	results := p.symbolTrie.SearchExact(fqn.Name())
	for _, res := range results {
		if res.Namespace != fqn.Namespace() {
			continue
		}

		for _, kind := range kinds {
			if res.Symbol.NodeKind() == kind {
				return res
			}
		}
	}

	return nil
}

func (p *Project) FindFileInTrie(fqn *typer.FQN, kind ir.NodeKind) *File {
	node := p.FindNodeInTrie(fqn, kind)
	if node == nil {
		return nil
	}

	return p.GetFile(node.Path)
}

func (p *Project) FindFileAndSymbolInTrie(
	fqn *typer.FQN,
	kind ir.NodeKind,
) (*File, *traversers.TrieNode) {
	node := p.FindNodeInTrie(fqn, kind)
	if node == nil {
		return nil, nil
	}

	file := p.GetFile(node.Path)

	return file, node
}
