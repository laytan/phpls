package symbol

import (
	"fmt"
	"log"

	"github.com/laytan/phpls/pkg/nodescopes"
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor"
	"github.com/laytan/php-parser/pkg/visitor/traverser"
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
	res := c.FindMethods(true, filters...)
	if len(res) == 0 {
		return nil
	}

	return res[0]
}

func (c *ClassLike) FindMethods(shortCircuit bool, filters ...FilterFunc[*Method]) (res []*Method) {
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

		res = append(res, m)
		if shortCircuit {
			return res
		}
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

	if mth, ok := node.(*ast.StmtClassMethod); ok {
		m.Methods = append(m.Methods, mth)
	}

	return !nodescopes.IsScope(node.GetType())
}
