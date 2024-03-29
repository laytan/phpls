package symbol

import (
	"errors"
	"fmt"
	"log"

	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor/traverser"
	"github.com/laytan/phpls/internal/doxcontext"
	"github.com/laytan/phpls/pkg/fqn"
	"github.com/laytan/phpls/pkg/nodeident"
	"github.com/laytan/phpls/pkg/phpdoxer"
)

var ErrNoPropertyType = errors.New("property has no type declaration")

type Property struct {
	*modified
	*doxed

	cls  *ClassLike
	node ast.Vertex // A *ast.Parameter or *ast.StmtPropertyList.
}

// node should either be a *ast.Parameter or *ast.StmtPropertyList.
// *ast.Parameter is a constructor promoted property.
func NewProperty(cls *ClassLike, node ast.Vertex) *Property {
	t := node.GetType()
	if t != ast.TypeStmtPropertyList && t != ast.TypeParameter {
		log.Panicf(
			"[FATAL]: %T is not a property, *ast.Parameter or *ast.StmtPropertyList should be given",
			node,
		)
	}

	return &Property{
		cls:      cls,
		node:     node,
		modified: newModifiedFromNode(node),
		doxed:    NewDoxed(node),
	}
}

func (p *Property) Name() string {
	return nodeident.Get(p.node)
}

func (p *Property) Node() ast.Vertex {
	return p.node
}

// Implements the Member interface.
func (p *Property) Vertex() ast.Vertex {
	return p.node
}

// Type resolves the property's type. The 2nd return value is the enclosing class
// of the property node that had the type definition.
func (p *Property) Type() (phpdoxer.Type, *ClassLike, error) {
	typ, err := p.ownType()
	if err != nil && !errors.Is(err, ErrNoPropertyType) {
		return nil, nil, fmt.Errorf("getting own type of prop %s: %w", p.Name(), err)
	}

	if typ != nil {
		return typ, p.cls, nil
	}

	iter := p.cls.InheritsIter()
	for inhCls, done, err := iter(); !done; inhCls, done, err = iter() {
		if err != nil {
			return nil, nil, fmt.Errorf("iterating inherited classes of %s: %w", p.cls.Name(), err)
		}

		prop := inhCls.FindProperty(FilterOverwrittenBy(p, true))
		if prop == nil {
			continue
		}

		propTyp, err := prop.ownType()
		if err != nil && !errors.Is(err, ErrNoPropertyType) {
			return nil, nil, fmt.Errorf("getting inherited property's own type: %w", err)
		}

		if propTyp != nil {
			return propTyp, inhCls, nil
		}
	}

	return nil, nil, fmt.Errorf(
		"iterated all inherited classes for prop %s type: %w",
		p.Name(),
		ErrNoPropertyType,
	)
}

func (p *Property) ClsType() ([]*phpdoxer.TypeClassLike, error) {
	doc, cls, err := p.Type()
	if err != nil {
		return nil, fmt.Errorf("getting type to apply context to: %w", err)
	}

	fqnt := fqn.NewTraverser()
	tv := traverser.NewTraverser(fqnt)
	cls.Root().Accept(tv)

	return doxcontext.ApplyContext(fqnt, cls.GetFQN(), p.node.GetPosition(), doc), nil
}

func (p *Property) ownType() (phpdoxer.Type, error) {
	varDoc := p.FindDoc(FilterDocKind(phpdoxer.KindVar))
	if varDoc != nil {
		return varDoc.(*phpdoxer.NodeVar).Type, nil
	}

	var typ ast.Vertex
	switch tn := p.node.(type) {
	case *ast.StmtPropertyList:
		typ = tn.Type
	case *ast.Parameter:
		typ = tn.Type
	}

	if typ == nil {
		return nil, fmt.Errorf("no @var or type hint for prop %s: %w", p.Name(), ErrNoPropertyType)
	}

	doctyp, err := TypeHintToDocType(typ)
	if err != nil {
		return nil, fmt.Errorf("parsing property %s type hint: %w", p.Name(), err)
	}

	return doctyp, nil
}
