package symbol

import (
	"fmt"
	"log"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/nodescopes"
)

type (
	MethodsIterFunc func() (method *Method, done bool, genErr error)
)

func (c *ClassLike) MethodsIter() MethodsIterFunc {
	mt := &methodsTraverser{Methods: []*ir.ClassMethodStmt{}}
	c.node.Walk(mt)
	i := 0

	return func() (m *Method, done bool, genErr error) {
		if len(mt.Methods) <= i {
			return nil, true, nil
		}

		m = NewMethod(c.rooter, mt.Methods[i])
		i++
		return m, false, nil
	}
}

func (c *ClassLike) FindMethod(filters ...FilterFunc[*Method]) *Method {
	iter := c.MethodsIter()
MethodsIter:
	for m, done, err := iter(); !done; m, done, err = iter() {
		if err != nil {
			log.Println(fmt.Errorf("[symbol.ClassLike.FindMethod]: %w", err))
			continue
		}

		for _, filter := range filters {
			if !filter(m) {
				continue MethodsIter
			}
		}

		return m
	}

	return nil
}

type methodsTraverser struct {
	Methods        []*ir.ClassMethodStmt
	firstTraversed bool
}

func (m *methodsTraverser) EnterNode(node ir.Node) bool {
	if !m.firstTraversed {
		if !nodescopes.IsClassLike(ir.GetNodeKind(node)) {
			log.Panicf(
				"[symbol.methodsTraverser.EnterNode]: methodsTraverser can only be used on class-like nodes, got %T",
				node,
			)
		}

		m.firstTraversed = true

		return true
	}

	if method, ok := node.(*ir.ClassMethodStmt); ok {
		m.Methods = append(m.Methods, method)
	}

	return !nodescopes.IsScope(ir.GetNodeKind(node))
}

func (m *methodsTraverser) LeaveNode(ir.Node) {}
