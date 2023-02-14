package symbol

import (
	"fmt"
	"log"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/nodescopes"
)

type (
	PropertiesIterFunc func() (property *Property, done bool, genErr error)
)

func (c *ClassLike) PropertiesIter() PropertiesIterFunc {
	pt := &propertiesTraverser{Properties: []*ir.PropertyListStmt{}}
	c.node.Walk(pt)
	i := 0

	return func() (p *Property, done bool, genErr error) {
		if len(pt.Properties) <= i {
			return nil, true, nil
		}

		p = NewProperty(pt.Properties[i])
		i++
		return p, false, nil
	}
}

func (c *ClassLike) FindProperty(filters ...FilterFunc[*Property]) *Property {
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

		return p
	}

	return nil
}

type propertiesTraverser struct {
	Properties     []*ir.PropertyListStmt
	firstTraversed bool
}

func (t *propertiesTraverser) EnterNode(node ir.Node) bool {
	if !t.firstTraversed {
		if !nodescopes.IsClassLike(ir.GetNodeKind(node)) {
			log.Panicf(
				"[symbol.propertiesTraverser.EnterNode]: propertiesTraverser can only be used on class-like nodes, got %T",
				node,
			)
		}

		t.firstTraversed = true

		return true
	}

	if property, ok := node.(*ir.PropertyListStmt); ok {
		t.Properties = append(t.Properties, property)
	}

	return !nodescopes.IsScope(ir.GetNodeKind(node))
}

func (t *propertiesTraverser) LeaveNode(ir.Node) {}
