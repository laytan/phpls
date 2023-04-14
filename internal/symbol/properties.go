package symbol //nolint:dupl // The subtle changes justify the duplication imo.

import (
	"fmt"
	"log"

	"github.com/laytan/elephp/pkg/nodeident"
	"github.com/laytan/elephp/pkg/nodescopes"
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor"
	"github.com/laytan/php-parser/pkg/visitor/traverser"
)

type (
	PropertiesIterFunc func() (property *Property, done bool, genErr error)
)

func (c *ClassLike) PropertiesIter() PropertiesIterFunc {
	pt := &propertiesTraverser{Properties: []ast.Vertex{}}
	tv := traverser.NewTraverser(pt)
	c.node.Accept(tv)
	i := 0

	return func() (p *Property, done bool, genErr error) {
		if len(pt.Properties) <= i {
			return nil, true, nil
		}

		p = NewProperty(c, pt.Properties[i])
		i++
		return p, false, nil
	}
}

func (c *ClassLike) FindProperty(filters ...FilterFunc[*Property]) *Property {
	res := c.FindProperties(true, filters...)
	if len(res) == 0 {
		return nil
	}

	return res[0]
}

func (c *ClassLike) FindProperties(
	shortCircuit bool,
	filters ...FilterFunc[*Property],
) (res []*Property) {
	iter := c.PropertiesIter()
PropertiesIter:
	for p, done, err := iter(); !done; p, done, err = iter() {
		if err != nil {
			log.Println(fmt.Errorf("[symbol.ClassLike.FindProperty]: %w", err))
			continue
		}

		for _, filter := range filters {
			if !filter(p) {
				continue PropertiesIter
			}
		}

		res = append(res, p)
		if shortCircuit {
			return res
		}
	}

	return nil
}

type propertiesTraverser struct {
	visitor.Null
	Properties     []ast.Vertex
	firstTraversed bool
}

func (t *propertiesTraverser) EnterNode(node ast.Vertex) bool {
	if !t.firstTraversed {
		if !nodescopes.IsClassLike(node.GetType()) {
			log.Panicf(
				"[symbol.propertiesTraverser.EnterNode]: propertiesTraverser can only be used on class-like nodes, got %T",
				node,
			)
		}

		t.firstTraversed = true

		return true
	}

	// Check promoted properties in constructor.
	if meth, ok := node.(*ast.StmtClassMethod); ok && nodeident.Get(meth) == "__construct" {
		for _, param := range meth.Params {
			if newModifiedFromNode(param).HasExplicitPrivacy() {
				t.Properties = append(t.Properties, param)
			}
		}
	}

	return !nodescopes.IsScope(node.GetType())
}

func (t *propertiesTraverser) StmtPropertyList(property *ast.StmtPropertyList) {
	t.Properties = append(t.Properties, property)
}
