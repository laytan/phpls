package symbol //nolint:dupl // The subtle changes justify the duplication imo.

import (
	"fmt"
	"log"

	"github.com/laytan/elephp/pkg/nodescopes"
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor"
	"github.com/laytan/php-parser/pkg/visitor/traverser"
)

type (
	ConstantsIterFunc func() (method *ClassConst, done bool, genErr error)
)

func (c *ClassLike) ConstantsIter() ConstantsIterFunc {
	pt := &classConstsTraverser{Constants: []*ast.StmtClassConstList{}}
	tv := traverser.NewTraverser(pt)
	c.node.Accept(tv)
	i := 0

	return func() (p *ClassConst, done bool, genErr error) {
		if len(pt.Constants) <= i {
			return nil, true, nil
		}

		p = NewClassConst(pt.Constants[i])
		i++
		return p, false, nil
	}
}

func (c *ClassLike) FindConstant(filters ...FilterFunc[*ClassConst]) *ClassConst {
	iter := c.ConstantsIter()
ConstantsIter:
	for p, done, err := iter(); !done; p, done, err = iter() {
		if err != nil {
			log.Println(fmt.Errorf("[symbol.ClassLike.FindConstant]: %w", err))
			continue
		}

		for _, filter := range filters {
			if !filter(p) {
				continue ConstantsIter
			}
		}

		return p
	}

	return nil
}

type classConstsTraverser struct {
	visitor.Null
	Constants      []*ast.StmtClassConstList
	firstTraversed bool
}

func (t *classConstsTraverser) EnterNode(node ast.Vertex) bool {
	if !t.firstTraversed {
		if !nodescopes.IsClassLike(node.GetType()) {
			log.Panicf(
				"[symbol.classConstsTraverser.EnterNode]: can only be used on class-like nodes, got %T",
				node,
			)
		}

		t.firstTraversed = true

		return true
	}

	return !nodescopes.IsScope(node.GetType())
}

func (t *classConstsTraverser) StmtClassConstList(node *ast.StmtClassConstList) {
	t.Constants = append(t.Constants, node)
}
