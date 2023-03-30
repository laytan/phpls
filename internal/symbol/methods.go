package symbol

import (
	"fmt"
	"log"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/visitor"
	"github.com/VKCOM/php-parser/pkg/visitor/traverser"
	"github.com/laytan/elephp/pkg/nodescopes"
)

type (
	MethodsIterFunc func() (method *Method, done bool, genErr error)
)

func (c *ClassLike) MethodsIter() MethodsIterFunc {
	mt := &methodsTraverser{Methods: []*ast.StmtClassMethod{}}
	tv := traverser.NewTraverser(mt)
	c.node.Accept(tv)
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
	visitor.Null
	Methods        []*ast.StmtClassMethod
	firstTraversed bool
}

func (m *methodsTraverser) EnterNode(node ast.Vertex) bool {
	if !m.firstTraversed {
		if !nodescopes.IsClassLike(node.GetType()) {
			log.Panicf(
				"[symbol.methodsTraverser.EnterNode]: methodsTraverser can only be used on class-like nodes, got %T",
				node,
			)
		}

		m.firstTraversed = true

		return true
	}

	return !nodescopes.IsScope(node.GetType())
}

func (m *methodsTraverser) StmtClassMethod(method *ast.StmtClassMethod) {
	m.Methods = append(m.Methods, method)
}
