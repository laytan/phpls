package expr

import (
	"fmt"
	"log"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/fqner"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/symbol"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/nodeident"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/phprivacy"
	"github.com/laytan/elephp/pkg/typer"
)

var resolvers = map[Type]ClassResolver{
	TypeProperty:      &propertyResolver{},
	TypeMethod:        &methodResolver{},
	TypeStaticMethod:  &staticMethodResolver{},
	TypeClassConstant: &classConstResolver{},
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
		ExprType:   TypeProperty,
		Identifier: nodeident.Get(propNode.Property),
		Position:   ir.GetPosition(propNode.Property),
	}, propNode.Variable, true
}

// Up finds the non-static property toResolve.Identifier inside the ctx class and
// its inherited classes.
//
// The first arg will contain the property node&path, the 2nd will be the
// type of this property, which is nil if it is not a class.
func (p *propertyResolver) Up(
	ctx *fqn.FQN,
	privacy phprivacy.Privacy,
	toResolve *DownResolvement,
) (*Resolved, *fqn.FQN, bool) {
	if toResolve.ExprType != TypeProperty {
		return nil, nil, false
	}

	cls := expandCtx(ctx)
	if cls == nil {
		return nil, nil, false
	}

	prop := cls.FindProperty(
		symbol.FilterName[*symbol.Property](toResolve.Identifier),
		symbol.FilterNotStatic[*symbol.Property](),
		symbol.FilterCanBeAccessedFrom[*symbol.Property](
			determinePrivacy(privacy, cls.Kind(), &iteration{
				first:      true,
				firstClass: true,
			}),
		),
	)
	if prop != nil {
		resolved, clsType := resolveProp(cls, prop)
		return resolved, clsType, true
	}

	isFirstClass := true
	iter := cls.InheritsIter()
	for inhCls, done, err := iter(); !done; inhCls, done, err = iter() {
		if err != nil {
			log.Println(fmt.Errorf("[expr.propertyResolver.Up]: %w", err))
			continue
		}

		prop := inhCls.FindProperty(
			symbol.FilterName[*symbol.Property](toResolve.Identifier),
			symbol.FilterNotStatic[*symbol.Property](),
			symbol.FilterCanBeAccessedFrom[*symbol.Property](
				determinePrivacy(privacy, inhCls.Kind(), &iteration{
					first:      false,
					firstClass: isFirstClass,
				}),
			),
		)

		if inhCls.Kind() == ir.KindClassStmt {
			isFirstClass = false
		}

		if prop != nil {
			resolved, clsType := resolveProp(inhCls, prop)
			return resolved, clsType, true
		}
	}

	return nil, nil, false
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
		ExprType:   TypeMethod,
		Identifier: nodeident.Get(propNode),
		Position:   propNode.Position,
	}, propNode.Variable, true
}

// Up finds the non-static method toResolve.Identifier inside the ctx class and
// its inherited classes.
//
// The first arg will contain the method node&path, the 2nd will be the return
// type of this method, which is nil if it is not a class.
func (p *methodResolver) Up(
	ctx *fqn.FQN,
	privacy phprivacy.Privacy,
	toResolve *DownResolvement,
) (*Resolved, *fqn.FQN, bool) {
	if toResolve.ExprType != TypeMethod {
		return nil, nil, false
	}

	return methodUp(ctx, privacy, toResolve, symbol.FilterNotStatic[*symbol.Method]())
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
		ExprType:   TypeStaticMethod,
		Identifier: nodeident.Get(propNode),
		Position:   propNode.Position,
	}, propNode.Class, true
}

func (p *staticMethodResolver) Up(
	ctx *fqn.FQN,
	privacy phprivacy.Privacy,
	toResolve *DownResolvement,
) (*Resolved, *fqn.FQN, bool) {
	if toResolve.ExprType != TypeStaticMethod {
		return nil, nil, false
	}

	return methodUp(
		ctx,
		privacy,
		toResolve,
		symbol.FilterStatic[*symbol.Method](),
	)
}

type classConstResolver struct{}

func (c *classConstResolver) Down(
	node ir.Node,
) (resolvement *DownResolvement, next ir.Node, done bool) {
	constFetch, ok := node.(*ir.ClassConstFetchExpr)
	if !ok {
		return nil, nil, false
	}

	next = constFetch.Class

	ident := nodeident.Get(constFetch.Class)
	if ident == "self" || ident == "parent" || ident == "this" || ident == "static" {
		next = &ir.SimpleVar{
			Position: ir.GetPosition(next),
			Name:     ident,
		}
	}

	return &DownResolvement{
		ExprType:   TypeClassConstant,
		Identifier: constFetch.ConstantName.Value,
		Position:   constFetch.ConstantName.Position,
	}, next, true
}

func (c *classConstResolver) Up(
	ctx *fqn.FQN,
	privacy phprivacy.Privacy,
	toResolve *DownResolvement,
) (result *Resolved, nextCtx *fqn.FQN, done bool) {
	if toResolve.ExprType != TypeClassConstant {
		return nil, nil, false
	}

	cls := expandCtx(ctx)
	if cls == nil {
		return nil, nil, false
	}

	con := cls.FindConstant(
		symbol.FilterName[*symbol.ClassConst](toResolve.Identifier),
		symbol.FilterCanBeAccessedFrom[*symbol.ClassConst](
			determinePrivacy(privacy, cls.Kind(), &iteration{
				first:      true,
				firstClass: true,
			}),
		),
	)
	if con != nil {
		resolved, clsType := resolveConst(cls, con)
		return resolved, clsType, true
	}

	isFirstClass := true
	iter := cls.InheritsIter()
	for inhCls, done, err := iter(); !done; inhCls, done, err = iter() {
		if err != nil {
			log.Println(fmt.Errorf("[expr.ClassConstResolver.Up]: %w", err))
			continue
		}

		con := inhCls.FindConstant(
			symbol.FilterName[*symbol.ClassConst](toResolve.Identifier),
			symbol.FilterCanBeAccessedFrom[*symbol.ClassConst](
				determinePrivacy(privacy, inhCls.Kind(), &iteration{
					first:      false,
					firstClass: isFirstClass,
				}),
			),
		)

		if inhCls.Kind() == ir.KindClassStmt {
			isFirstClass = false
		}

		if con != nil {
			resolved, clsType := resolveConst(inhCls, con)
			return resolved, clsType, true
		}
	}

	return nil, nil, false
}

func methodUp(
	ctx *fqn.FQN,
	privacy phprivacy.Privacy,
	toResolve *DownResolvement,
	extraFilter symbol.FilterFunc[*symbol.Method],
) (*Resolved, *fqn.FQN, bool) {
	cls := expandCtx(ctx)
	if cls == nil {
		return nil, nil, false
	}

	m := cls.FindMethod(
		symbol.FilterName[*symbol.Method](toResolve.Identifier),
		symbol.FilterCanBeAccessedFrom[*symbol.Method](
			determinePrivacy(privacy, cls.Kind(), &iteration{
				first:      true,
				firstClass: true,
			}),
		),
		extraFilter,
	)

	if m != nil {
		resolved, clsType := resolveMethod(cls, m)
		return resolved, clsType, true
	}

	isFirstClass := true
	iter := cls.InheritsIter()
	for inhCls, done, err := iter(); !done; inhCls, done, err = iter() {
		if err != nil {
			log.Println(fmt.Errorf("[expr.methodResolver.Up]: %w", err))
			continue
		}

		if inhCls.Kind() == ir.KindClassStmt {
			isFirstClass = false
		}

		m := inhCls.FindMethod(
			symbol.FilterName[*symbol.Method](toResolve.Identifier),
			symbol.FilterCanBeAccessedFrom[*symbol.Method](
				determinePrivacy(privacy, inhCls.Kind(), &iteration{
					first:      false,
					firstClass: isFirstClass,
				}),
			),
			extraFilter,
		)

		if m != nil {
			resolved, clsType := resolveMethod(inhCls, m)
			return resolved, clsType, true
		}
	}

	return nil, nil, false
}

func resolveProp(
	cls *symbol.ClassLike,
	prop *symbol.Property,
) (*Resolved, *fqn.FQN) {
	resolvement := &Resolved{
		Node: prop.Node(),
		Path: cls.Path(),
	}

	res := typer.FromContainer().Property(cls.Root(), prop.Node())
	clsRes, ok := res.(*phpdoxer.TypeClassLike)
	if !ok {
		return resolvement, nil
	}

	return resolvement, fqner.FullyQualifyName(cls.Root(), &ir.Name{
		Position: prop.Node().Position,
		Value:    clsRes.Name,
	})
}

func resolveMethod(
	cls *symbol.ClassLike,
	m *symbol.Method,
) (*Resolved, *fqn.FQN) {
	resolvement := &Resolved{
		Node: m.Node(),
		Path: cls.Path(),
	}

	res := m.ReturnsClass(cls.GetFQN())
	if len(res) > 0 {
		// TODO: return all returned classes, and check them all accordingly.
		// Will require some rewrite of this pkg.
		return resolvement, fqn.New(res[0].Name)
	}

	return resolvement, nil
}

func resolveConst(
	cls *symbol.ClassLike,
	cnst *symbol.ClassConst,
) (*Resolved, *fqn.FQN) {
	resolvement := &Resolved{
		Node: cnst.Node(),
		Path: cls.Path(),
	}

	return resolvement, nil
}

func expandCtx(ctx *fqn.FQN) *symbol.ClassLike {
	iNode, ok := index.FromContainer().Find(ctx)
	if !ok {
		log.Println(fmt.Errorf("[expr.expandCtx(%v)]: can't find in index", ctx))
		return nil
	}

	rooter := wrkspc.NewRooter(iNode.Path)
	cls, err := symbol.NewClassLikeFromFQN(rooter, ctx)
	if err != nil {
		log.Println(fmt.Errorf("[expr.expandCtx(%v)]: %w", ctx, err))
		return nil
	}

	return cls
}
