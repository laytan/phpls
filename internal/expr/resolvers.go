package expr

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/phprivacy"
	"github.com/laytan/elephp/pkg/symbol"
	"github.com/laytan/elephp/pkg/traversers"
)

var resolvers = map[ExprType]ClassResolver{
	ExprTypeProperty:     &propertyResolver{},
	ExprTypeMethod:       &methodResolver{},
	ExprTypeStaticMethod: &staticMethodResolver{},
}

type propertyResolver struct{}

func (p *propertyResolver) Down(
	node ir.Node,
) (*DownResolvement, ir.Node, bool) {
	propNode, ok := node.(*ir.PropertyFetchExpr)
	if !ok {
		return nil, nil, false
	}

	return &DownResolvement{
		ExprType:   ExprTypeProperty,
		Identifier: symbol.GetIdentifier(propNode.Property),
	}, propNode.Variable, true
}

func (p *propertyResolver) Up(
	ctx *phpdoxer.TypeClassLike,
	toResolve *DownResolvement,
) (*phpdoxer.TypeClassLike, bool) {
	if toResolve.ExprType != ExprTypeProperty {
		return nil, false
	}

	return createAndWalkResolveQueue(ctx, func(wc *walkContext) *phpdoxer.TypeClassLike {
		// TODO: proper privacy handling.
		// TODO: move the property traverser to this package.
		t := traversers.NewProperty(toResolve.Identifier, wc.FQN.Name(), phprivacy.PrivacyPrivate)
		wc.Root.Walk(t)

		if t.Property == nil {
			return nil
		}

		res := Typer().Property(wc.Root, t.Property)
		clsRes, ok := res.(*phpdoxer.TypeClassLike)
		if !ok {
			return nil
		}

		return clsRes
	})
}

type methodResolver struct{}

func (p *methodResolver) Down(
	node ir.Node,
) (*DownResolvement, ir.Node, bool) {
	propNode, ok := node.(*ir.MethodCallExpr)
	if !ok {
		return nil, nil, false
	}

	return &DownResolvement{
		ExprType:   ExprTypeMethod,
		Identifier: symbol.GetIdentifier(propNode),
	}, propNode.Variable, true
}

func (p *methodResolver) Up(
	ctx *phpdoxer.TypeClassLike,
	toResolve *DownResolvement,
) (*phpdoxer.TypeClassLike, bool) {
	if toResolve.ExprType != ExprTypeMethod {
		return nil, false
	}

	return createAndWalkResolveQueue(ctx, func(wc *walkContext) *phpdoxer.TypeClassLike {
		// TODO: proper privacy handling.
		// TODO: move the property method to this package.
		t := traversers.NewMethod(toResolve.Identifier, wc.FQN.Name(), phprivacy.PrivacyPrivate)
		wc.Root.Walk(t)

		if t.Method == nil {
			return nil
		}

		res := Typer().Returns(wc.Root, t.Method)
		clsRes, ok := res.(*phpdoxer.TypeClassLike)
		if !ok {
			return nil
		}

		return clsRes
	})
}

type staticMethodResolver struct{}

func (p *staticMethodResolver) Down(
	node ir.Node,
) (*DownResolvement, ir.Node, bool) {
	propNode, ok := node.(*ir.StaticCallExpr)
	if !ok {
		return nil, nil, false
	}

	return &DownResolvement{
		ExprType:   ExprTypeStaticMethod,
		Identifier: symbol.GetIdentifier(propNode),
	}, propNode.Class, true
}

func (p *staticMethodResolver) Up(
	ctx *phpdoxer.TypeClassLike,
	toResolve *DownResolvement,
) (*phpdoxer.TypeClassLike, bool) {
	if toResolve.ExprType != ExprTypeStaticMethod {
		return nil, false
	}

	return createAndWalkResolveQueue(ctx, func(wc *walkContext) *phpdoxer.TypeClassLike {
		// TODO: proper privacy handling.
		// TODO: move the property method to this package.
		t := traversers.NewMethodStatic(
			toResolve.Identifier,
			wc.FQN.Name(),
			phprivacy.PrivacyPrivate,
		)
		wc.Root.Walk(t)

		if t.Method == nil {
			return nil
		}

		res := Typer().Returns(wc.Root, t.Method)
		clsRes, ok := res.(*phpdoxer.TypeClassLike)
		if !ok {
			return nil
		}

		return clsRes
	})
}
