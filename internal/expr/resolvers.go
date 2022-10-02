package expr

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/phprivacy"
	"github.com/laytan/elephp/pkg/symbol"
	"github.com/laytan/elephp/pkg/traversers"
	"github.com/laytan/elephp/pkg/typer"
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
	privacy phprivacy.Privacy,
	toResolve *DownResolvement,
) (*Resolved, *phpdoxer.TypeClassLike, bool) {
	if toResolve.ExprType != ExprTypeProperty {
		return nil, nil, false
	}

	// Is this the first iteration/classlike we are checking?
	isFirst := true

	// Is this the first class or traits used by the first class?
	isFirstClass := true

	return createAndWalkResolveQueue(
		ctx,
		func(wc *walkContext) (*Resolved, *phpdoxer.TypeClassLike) {
			defer func() { isFirst = false }()

			currKind := wc.Curr.Symbol.NodeKind()

			// if this is a class, but not the first one.
			if !isFirst && currKind == ir.KindClassStmt {
				isFirstClass = false
			}

			actPrivacy := determinePrivacy(privacy, currKind, isFirst, isFirstClass)

			// TODO: move the property traverser to this package.
			t := traversers.NewProperty(toResolve.Identifier, wc.FQN.Name(), actPrivacy)
			wc.Root.Walk(t)

			if t.Property == nil {
				return nil, nil
			}

			resolved := &Resolved{
				Node: t.Property,
				Path: wc.Curr.Path,
			}

			res := typer.FromContainer().Property(wc.Root, t.Property)
			clsRes, ok := res.(*phpdoxer.TypeClassLike)
			if !ok {
				return resolved, nil
			}

			return resolved, clsRes
		},
	)
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
	privacy phprivacy.Privacy,
	toResolve *DownResolvement,
) (*Resolved, *phpdoxer.TypeClassLike, bool) {
	if toResolve.ExprType != ExprTypeMethod {
		return nil, nil, false
	}

	return methodUp(ctx, privacy, toResolve, traversers.NewMethod)
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
	privacy phprivacy.Privacy,
	toResolve *DownResolvement,
) (*Resolved, *phpdoxer.TypeClassLike, bool) {
	if toResolve.ExprType != ExprTypeStaticMethod {
		return nil, nil, false
	}

	return methodUp(
		ctx,
		privacy,
		toResolve,
		traversers.NewMethodStatic,
	)
}

// Determines the privacy to search for based on all the conditions determined
// by PHP.
func determinePrivacy(
	startPrivacy phprivacy.Privacy,
	currKind ir.NodeKind,
	isFirst, isFirstClass bool,
) phprivacy.Privacy {
	actPrivacy := startPrivacy

	// If we are in the class, the first run can check >= private members,
	// the rest only >= protected members.
	if !isFirst && actPrivacy == phprivacy.PrivacyPrivate {
		actPrivacy = phprivacy.PrivacyProtected
	}

	// If this is a trait, and it is used from the first class,
	// private methods are also accessible.
	if isFirstClass && actPrivacy == phprivacy.PrivacyProtected &&
		currKind == ir.KindTraitStmt {
		actPrivacy = phprivacy.PrivacyPrivate
	}

	return actPrivacy
}

func methodUp(
	ctx *phpdoxer.TypeClassLike,
	privacy phprivacy.Privacy,
	toResolve *DownResolvement,
	newTraverser func(name, classLikeName string, privacy phprivacy.Privacy) *traversers.Method,
) (*Resolved, *phpdoxer.TypeClassLike, bool) {
	// Is this the first iteration/classlike we are checking?
	isFirst := true

	// Is this the first class or traits used by the first class?
	isFirstClass := true

	return createAndWalkResolveQueue(
		ctx,
		func(wc *walkContext) (*Resolved, *phpdoxer.TypeClassLike) {
			defer func() { isFirst = false }()

			currKind := wc.Curr.Symbol.NodeKind()

			// if this is a class, but not the first one.
			if !isFirst && currKind == ir.KindClassStmt {
				isFirstClass = false
			}

			actPrivacy := determinePrivacy(privacy, currKind, isFirst, isFirstClass)

			t := newTraverser(toResolve.Identifier, wc.FQN.Name(), actPrivacy)
			wc.Root.Walk(t)

			if t.Method == nil {
				return nil, nil
			}

			resolved := &Resolved{
				Node: t.Method,
				Path: wc.Curr.Path,
			}

			if res := typer.FromContainer().Returns(wc.Root, t.Method, rootRetriever); res != nil {
				if clsRes, ok := res.(*phpdoxer.TypeClassLike); ok {
					return resolved, clsRes
				}
			}

			return resolved, nil
		},
	)
}
