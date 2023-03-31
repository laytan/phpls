package expr

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/laytan/elephp/internal/fqner"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/symbol"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/functional"
	"github.com/laytan/elephp/pkg/ie"
	"github.com/laytan/elephp/pkg/nodeident"
	"github.com/laytan/elephp/pkg/nodescopes"
	"github.com/laytan/elephp/pkg/nodevar"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/phprivacy"
	"github.com/laytan/elephp/pkg/traversers"
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor/traverser"
)

var starters = map[Type]StartResolver{
	TypeVariable: &variableResolver{},
	TypeName:     &nameResolver{},
	TypeFunction: &functionResolver{},
	TypeNew:      &newResolver{},
}

type variableResolver struct{}

func (p *variableResolver) Down(
	node ast.Vertex,
) (*DownResolvement, ast.Vertex, bool) {
	propNode, ok := node.(*ast.ExprVariable)
	if !ok {
		return nil, nil, false
	}

	return &DownResolvement{
		ExprType:   TypeVariable,
		Identifier: nodeident.Get(propNode),
		Position:   propNode.Position,
	}, nil, true
}

func (p *variableResolver) Up(
	scopes *Scopes,
	toResolve *DownResolvement,
) (*Resolved, *fqn.FQN, phprivacy.Privacy, bool) {
	wrk := wrkspc.Current
	switch toResolve.Identifier {
	case "$this", "self", "static":
		if node, ok := fqner.FindFullyQualifiedName(scopes.Root, &ast.Name{
			Position: scopes.Block.GetPosition(),
			Parts:    nameParts(nodeident.Get(scopes.Class)),
		}); ok {
			return &Resolved{Path: node.Path, Node: node.ToIRNode(wrk.FIROf(node.Path))},
				node.FQN,
				phprivacy.PrivacyPrivate,
				true
		}

		log.Println("encountered $this, self::, or static:: but could not find the subject class")
		return nil, nil, 0, false

	case "parent":
		node := parentOf(scopes)
		return &Resolved{Path: node.Path, Node: node.ToIRNode(wrk.FIROf(node.Path))},
			node.FQN,
			phprivacy.PrivacyProtected,
			true

	default:
		if toResolve.ExprType != TypeVariable {
			return nil, nil, 0, false
		}

		t := traversers.NewVariable(toResolve.Identifier)
		tv := traverser.NewTraverser(t)
		scopes.Block.Accept(tv)
		if t.Result == nil {
			return nil, nil, 0, false
		}

		ta := traversers.NewAssignment(t.Result)
		tav := traverser.NewTraverser(ta)
		scopes.Block.Accept(tav)
		if ta.Assignment == nil || ta.Scope == nil {
			return nil, nil, 0, false
		}

		var cls *index.INode
		if scopes.Class.GetType() != ast.TypeRoot {
			cls, _ = fqner.FindFullyQualifiedName(scopes.Root, &ast.Name{
				Position: scopes.Class.GetPosition(),
				Parts:    nameParts(nodeident.Get(scopes.Class)),
			})
		}

		varSym := symbol.NewVariable(wrkspc.NewRooter(scopes.Path, scopes.Root), ta.Assignment)
		typ, err := varSym.TypeCls(
			ie.IfElseFunc(cls == nil, nil, func() *fqn.FQN { return cls.FQN }),
		)
		if err != nil && !errors.Is(err, symbol.ErrNoVarType) {
			log.Println(fmt.Errorf("retrieving variable type: %w", err))
			return nil, nil, 0, false
		}

		if len(typ) > 0 {
			return &Resolved{
					Path: scopes.Path,
					Node: ta.Assignment,
				}, fqn.New(
					typ[0].Name,
				), phprivacy.PrivacyPublic, true
		}

		if nodevar.IsAssignment(ta.Scope.GetType()) {
			if res, lastClass, left := Resolve(nodevar.AssignmentExpr(ta.Scope), scopes); left == 0 {
				return res, lastClass, phprivacy.PrivacyPublic, true
			}

			return nil, nil, 0, false
		}

		switch typedScope := ta.Scope.(type) {
		case *ast.Parameter:
			rooter := wrkspc.NewRooter(scopes.Path, scopes.Root)
			param := symbol.NewParameter(rooter, scopes.Block, typedScope)
			typ, err := param.TypeClass()
			if err != nil {
				if errors.Is(err, symbol.ErrNoParam) {
					break
				}

				log.Println(fmt.Errorf("getting param %v typ: %w", typedScope, err))
			}

			if len(typ) == 0 {
				break
			}

			// TODO: only using first typ, needs rewrite.
			return &Resolved{
				Node: ta.Assignment,
				Path: scopes.Path,
			}, fqn.New(typ[0].Name), phprivacy.PrivacyPublic, true

		default:
			log.Printf("TODO: resolve variable out of type %T", ta.Scope)
			return nil, nil, 0, false
		}
	}

	return nil, nil, 0, false
}

type nameResolver struct{}

func (p *nameResolver) Down(
	node ast.Vertex,
) (*DownResolvement, ast.Vertex, bool) {
	if !nodescopes.IsName(node.GetType()) {
		return nil, nil, false
	}

	return &DownResolvement{
		ExprType:   TypeName,
		Identifier: nodeident.Get(node),
		Position:   node.GetPosition(),
	}, nil, true
}

func (p *nameResolver) Up(
	scopes *Scopes,
	toResolve *DownResolvement,
) (*Resolved, *fqn.FQN, phprivacy.Privacy, bool) {
	if toResolve.ExprType != TypeName {
		return nil, nil, 0, false
	}

	qualified := fqner.FullyQualifyName(scopes.Root, &ast.Name{
		Position: toResolve.Position,
		Parts:    nameParts(toResolve.Identifier),
	})

	privacy, err := p.DeterminePrivacy(scopes, qualified)
	if err != nil {
		log.Println(fmt.Errorf("[expr.nameResolver.Up]: err determining privacy: %w", err))
		return nil, nil, 0, false
	}

	res, ok := index.Current.Find(qualified)
	if !ok {
		log.Printf("[expr.nameResolver.Up]: unable to find %s in index", qualified)
		return nil, nil, 0, false
	}

	return &Resolved{Path: res.Path, Node: res.ToIRNode(wrkspc.Current.FIROf(res.Path))},
		qualified,
		privacy,
		true
}

func (p *nameResolver) DeterminePrivacy(scopes *Scopes, fqn *fqn.FQN) (phprivacy.Privacy, error) {
	// If we are not in a class, it is automatically public access.
	if !nodescopes.IsClassLike(scopes.Class.GetType()) {
		return phprivacy.PrivacyPublic, nil
	}

	// If we are in the same class, private access.
	scopeFqn := fqner.FullyQualifyName(scopes.Root, &ast.Name{
		Position: scopes.Class.GetPosition(),
		Parts:    nameParts(nodeident.Get(scopes.Class)),
	})
	if fqn.String() == scopeFqn.String() {
		return phprivacy.PrivacyPrivate, nil
	}

	cls, err := symbol.NewClassLikeFromFQN(wrkspc.NewRooter(scopes.Path, scopes.Root), scopeFqn)
	if err != nil {
		return 0, fmt.Errorf("[nameResolver.DeterminePrivacy]: %w", err)
	}

	iter := cls.InheritsIter()
	for inhCls, done, err := iter(); !done; inhCls, done, err = iter() {
		if err != nil {
			log.Println(fmt.Errorf("[nameResolver.DeterminePrivacy]: %w", err))
			continue
		}

		if inhCls.GetFQN().String() == fqn.String() {
			return phprivacy.PrivacyProtected, nil
		}
	}

	return phprivacy.PrivacyPublic, nil
}

type functionResolver struct{}

func (p *functionResolver) Down(
	node ast.Vertex,
) (*DownResolvement, ast.Vertex, bool) {
	propNode, ok := node.(*ast.ExprFunctionCall)
	if !ok {
		return nil, nil, false
	}

	return &DownResolvement{
		ExprType:   TypeFunction,
		Identifier: nodeident.Get(propNode),
		Position:   propNode.Position,
	}, nil, true
}

func (p *functionResolver) Up(
	scopes *Scopes,
	toResolve *DownResolvement,
) (*Resolved, *fqn.FQN, phprivacy.Privacy, bool) {
	if toResolve.ExprType != TypeFunction {
		return nil, nil, 0, false
	}

	t := traversers.NewFunctionCall(toResolve.Identifier)
	tv := traverser.NewTraverser(t)
	scopes.Block.Accept(tv)
	if t.Result == nil {
		log.Printf("could not find function matching function call %s", toResolve.Identifier)
		return nil, nil, 0, false
	}

	typeOfFunc := func(n ast.Vertex) *fqn.FQN {
		ret, _, err := symbol.NewFunction(
			wrkspc.NewRooter(scopes.Path, scopes.Root),
			n.(*ast.StmtFunction),
		).Returns()

		if err != nil && !errors.Is(err, symbol.ErrNoReturn) {
			log.Println(fmt.Errorf("getting return type of %v: %w", n, err))
		}

		if retCls, ok := ret.(*phpdoxer.TypeClassLike); ok {
			return fqner.FullyQualifyName(scopes.Root, &ast.Name{
				Position: n.GetPosition(),
				Parts:    nameParts(retCls.Name),
			})
		}

		return nil
	}

	// If have local scope, check it for the function.
	if scopes.Block.GetType() != ast.TypeRoot {
		ft := traversers.NewFunction(toResolve.Identifier)
		tv := traverser.NewTraverser(ft)
		scopes.Block.Accept(tv)

		if ft.Function != nil {
			return &Resolved{
				Node: ft.Function,
				Path: scopes.Path,
			}, typeOfFunc(ft.Function), phprivacy.PrivacyPublic, true
		}
	}

	// Check for functions defined in the used namespaces.
	if def, ok := fqner.FindFullyQualifiedName(scopes.Root, &ast.Name{
		Position: toResolve.Position,
		Parts:    nameParts(toResolve.Identifier),
	}); ok {
		n := def.ToIRNode(wrkspc.Current.FIROf(def.Path))
		return &Resolved{
			Node: n,
			Path: def.Path,
		}, typeOfFunc(n), phprivacy.PrivacyPublic, true
	}

	// Check for global functions.
	key := fqn.New(fqn.PartSeperator + toResolve.Identifier)
	def, ok := index.Current.Find(key)
	if !ok {
		log.Println(fmt.Errorf("[expr.functionResolver.Up]: unable to find %s in index", key))
		return nil, nil, 0, false
	}

	n := def.ToIRNode(wrkspc.Current.FIROf(def.Path))
	return &Resolved{
		Node: n,
		Path: def.Path,
	}, typeOfFunc(n), phprivacy.PrivacyPublic, true
}

type newResolver struct{}

func (newresolver *newResolver) Down(
	node ast.Vertex,
) (*DownResolvement, ast.Vertex, bool) {
	newNode, ok := node.(*ast.ExprNew)
	if !ok {
		return nil, nil, false
	}

	// TODO: new expression using a non-name node
	if name, ok := newNode.Class.(*ast.Name); ok {
		return &DownResolvement{
			ExprType:   TypeNew,
			Identifier: nodeident.Get(name),
			Position:   name.Position,
		}, nil, true
	}

	log.Println("TODO: new expression using a non-name node")
	return nil, nil, false
}

func (newresolver *newResolver) Up(
	scopes *Scopes,
	toResolve *DownResolvement,
) (resolved *Resolved, nextCtx *fqn.FQN, privacy phprivacy.Privacy, done bool) {
	if toResolve.ExprType != TypeNew {
		return nil, nil, 0, false
	}

	if qualified := fqner.FullyQualifyName(
		scopes.Root,
		&ast.Name{Position: toResolve.Position, Parts: nameParts(toResolve.Identifier)},
	); qualified != nil {
		def, ok := index.Current.Find(qualified)
		if !ok {
			log.Println(fmt.Errorf("[expr.newResolver.Up]: unable to find %s in index", qualified))
			return nil, nil, 0, false
		}

		return &Resolved{
				Path: def.Path,
				Node: def.ToIRNode(wrkspc.Current.FIROf(def.Path)),
			},
			qualified,
			phprivacy.PrivacyPublic,
			true
	}

	return nil, nil, 0, false
}

func parentOf(scopes *Scopes) *index.INode {
	switch scopes.Class.(type) {
	case *ast.StmtClass:
		break
	case *ast.StmtTrait:
		// TODO: get parent of trait
		log.Println("TODO: get parent of trait")
		return nil

	default:
		log.Println("encountered parent:: but could not find a class node")
		return nil
	}

	class := scopes.Class.(*ast.StmtClass)
	if class.Extends == nil {
		log.Println("encountered parent:: but current class does not extend anything")
		return nil
	}

	if node, ok := fqner.FindFullyQualifiedName(
		scopes.Root,
		class.Extends.(*ast.Name),
	); ok {
		return node
	}

	log.Printf(
		"could not find fully qualified class for %s in index",
		nodeident.Get(class.Extends),
	)
	return nil
}

func nameParts(name string) []ast.Vertex {
	return functional.Map(
		strings.Split(name, "\\"),
		func(s string) ast.Vertex { return &ast.NamePart{Value: []byte(s)} },
	)
}
