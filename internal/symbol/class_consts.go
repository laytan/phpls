package symbol //nolint:dupl // The subtle changes justify the duplication imo.

import (
	"fmt"
	"log"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/nodescopes"
)

type (
	ConstantsIterFunc func() (method *ClassConst, done bool, genErr error)
)

func (c *ClassLike) ConstantsIter() ConstantsIterFunc {
	pt := &classConstsTraverser{Constants: []*ir.ClassConstListStmt{}}
	c.node.Walk(pt)
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
	Constants      []*ir.ClassConstListStmt
	firstTraversed bool
}

func (t *classConstsTraverser) EnterNode(node ir.Node) bool {
	if !t.firstTraversed {
		if !nodescopes.IsClassLike(ir.GetNodeKind(node)) {
			log.Panicf(
				"[symbol.classConstsTraverser.EnterNode]: can only be used on class-like nodes, got %T",
				node,
			)
		}

		t.firstTraversed = true

		return true
	}

	if constant, ok := node.(*ir.ClassConstListStmt); ok {
		t.Constants = append(t.Constants, constant)
	}

	return !nodescopes.IsScope(ir.GetNodeKind(node))
}

func (t *classConstsTraverser) LeaveNode(ir.Node) {}
