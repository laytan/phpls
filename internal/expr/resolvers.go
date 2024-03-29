package expr

import (
	"errors"
	"fmt"
	"log"

	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/phpls/internal/index"
	"github.com/laytan/phpls/internal/symbol"
	"github.com/laytan/phpls/internal/wrkspc"
	"github.com/laytan/phpls/pkg/fqn"
	"github.com/laytan/phpls/pkg/nodeident"
	"github.com/laytan/phpls/pkg/phprivacy"
)

var resolvers = map[Type]ClassResolver{
	TypeProperty:      &propertyResolver{},
	TypeMethod:        &methodResolver{},
	TypeStaticMethod:  &staticMethodResolver{},
	TypeClassConstant: &classConstResolver{},
}

type propertyResolver struct{}

func (p *propertyResolver) Down(
	node ast.Vertex,
) (*DownResolvement, ast.Vertex, bool) {
	propNode, ok := node.(*ast.ExprPropertyFetch)
	if !ok {
		return nil, nil, false
	}

	return &DownResolvement{
		ExprType:   TypeProperty,
		Identifier: nodeident.Get(propNode.Prop),
		Position:   propNode.Prop.GetPosition(),
	}, propNode.Var, true
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
		symbol.FilterName[*symbol.Property]("$"+toResolve.Identifier),
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
			symbol.FilterName[*symbol.Property]("$"+toResolve.Identifier),
			symbol.FilterNotStatic[*symbol.Property](),
			symbol.FilterCanBeAccessedFrom[*symbol.Property](
				determinePrivacy(privacy, inhCls.Kind(), &iteration{
					first:      false,
					firstClass: isFirstClass,
				}),
			),
		)

		if inhCls.Kind() == ast.TypeStmtClass {
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
	node ast.Vertex,
) (*DownResolvement, ast.Vertex, bool) {
	propNode, ok := node.(*ast.ExprMethodCall)
	if !ok {
		return nil, nil, false
	}

	return &DownResolvement{
		ExprType:   TypeMethod,
		Identifier: nodeident.Get(propNode),
		Position:   propNode.Position,
	}, propNode.Var, true
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
	node ast.Vertex,
) (*DownResolvement, ast.Vertex, bool) {
	propNode, ok := node.(*ast.ExprStaticCall)
	if !ok {
		return nil, nil, false
	}

	className := nodeident.Get(propNode.Class)

	// Special case, parent::foo() calls are not static calls
	// but they do look like them so they get here.
	if className == "parent" {
		return &DownResolvement{
			ExprType:   TypeMethod,
			Identifier: nodeident.Get(propNode),
			Position:   propNode.Position,
		}, propNode.Class, true
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
	node ast.Vertex,
) (resolvement *DownResolvement, next ast.Vertex, done bool) {
	constFetch, ok := node.(*ast.ExprClassConstFetch)
	if !ok {
		return nil, nil, false
	}

	next = constFetch.Class

	ident := nodeident.Get(constFetch.Class)
	if ident == "self" || ident == "parent" || ident == "$this" || ident == "static" {
		next = &ast.ExprVariable{
			Position: next.GetPosition(),
			Name:     &ast.Identifier{Value: []byte(ident)},
		}
	}

	return &DownResolvement{
		ExprType:   TypeClassConstant,
		Identifier: nodeident.Get(constFetch.Const),
		Position:   constFetch.Const.GetPosition(),
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

		if inhCls.Kind() == ast.TypeStmtClass {
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

		if inhCls.Kind() == ast.TypeStmtClass {
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

	typ, err := prop.ClsType()
	if err != nil {
		if !errors.Is(err, symbol.ErrNoPropertyType) {
			log.Println(fmt.Errorf("resolving prop type: %w", err))
		}

		return resolvement, nil
	}

	if len(typ) == 0 {
		return resolvement, nil
	}

	qualified := fqn.New(typ[0].Name)
	return resolvement, qualified
}

func resolveMethod(
	cls *symbol.ClassLike,
	m *symbol.Method,
) (*Resolved, *fqn.FQN) {
	resolvement := &Resolved{
		Node: m.Node(),
		Path: cls.Path(),
	}

	res, err := m.ReturnsClass()
	if err != nil && !errors.Is(err, symbol.ErrNoReturn) {
		log.Println(fmt.Errorf("resolving method %s return: %w", m.Name(), err))
	}
	if len(res) > 0 {
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
	iNode, ok := index.Current.Find(ctx)
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
