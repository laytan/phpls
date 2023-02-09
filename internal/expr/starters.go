package expr

import (
	"fmt"
	"log"

	"appliedgo.net/what"
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/common"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/phprivacy"
	"github.com/laytan/elephp/pkg/symbol"
	"github.com/laytan/elephp/pkg/traversers"
	"github.com/laytan/elephp/pkg/typer"
)

var starters = map[ExprType]StartResolver{
	ExprTypeVariable: &variableResolver{},
	ExprTypeName:     &nameResolver{},
	ExprTypeFunction: &functionResolver{},
	ExprTypeNew:      &newResolver{},
}

type variableResolver struct{}

func (p *variableResolver) Down(
	node ir.Node,
) (*DownResolvement, ir.Node, bool) {
	propNode, ok := node.(*ir.SimpleVar)
	if !ok {
		return nil, nil, false
	}

	return &DownResolvement{
		ExprType:   ExprTypeVariable,
		Identifier: symbol.GetIdentifier(propNode),
	}, nil, true
}

func (p *variableResolver) Up(
	scopes Scopes,
	toResolve *DownResolvement,
) (*Resolved, *phpdoxer.TypeClassLike, phprivacy.Privacy, bool) {
	typ := typer.FromContainer()

	if toResolve.ExprType != ExprTypeVariable {
		return nil, nil, 0, false
	}

	switch toResolve.Identifier {
	case "this", "self", "static":
		if node, ok := common.FindFullyQualified(
			scopes.Root,
			symbol.GetIdentifier(scopes.Class),
			symbol.ClassLikeScopes...); ok {
			_, n, err := common.SymbolToNode(node.Path, node.Symbol)
			if err != nil {
				log.Println(fmt.Errorf("[expr.variableResolver.Up]: %w", err))
				return nil, nil, 0, false
			}

			return &Resolved{Path: node.Path, Node: n},
				&phpdoxer.TypeClassLike{Name: node.FQN.String(), FullyQualified: true},
				phprivacy.PrivacyPrivate,
				true
		}

		log.Println("encountered $this, self::, or static:: but could not find the subject class")
		return nil, nil, 0, false

	case "parent":
		node := parentOf(scopes)
		_, n, err := common.SymbolToNode(node.Path, node.Symbol)
		if err != nil {
			log.Println(fmt.Errorf("[expr.variableResolver.Up]: %w", err))
			return nil, nil, 0, false
		}

		return &Resolved{Path: node.Path, Node: n},
			&phpdoxer.TypeClassLike{Name: node.FQN.String(), FullyQualified: true},
			phprivacy.PrivacyProtected,
			true

	default:
		t := traversers.NewVariable(toResolve.Identifier)
		scopes.Block.Walk(t)
		if t.Result == nil {
			return nil, nil, 0, false
		}

		ta := traversers.NewAssignment(t.Result)
		scopes.Block.Walk(ta)
		if ta.Assignment == nil || ta.Scope == nil {
			return nil, nil, 0, false
		}

		if docType := typ.Variable(scopes.Root, ta.Assignment, scopes.Block); docType != nil {
			if clsDocType, ok := docType.(*phpdoxer.TypeClassLike); ok {
				return &Resolved{Path: scopes.Path, Node: ta.Assignment},
					clsDocType,
					phprivacy.PrivacyPublic,
					true
			}

			what.Happens("@var doc is not a class-like type")
			return nil, nil, 0, false
		}

		switch typedScope := ta.Scope.(type) {
		case *ir.Assign:
			if res, lastClass, left := Resolve(typedScope.Expr, scopes); left == 0 {
				return res, lastClass, phprivacy.PrivacyPublic, true
			}

		case *ir.Parameter:
			if t := typ.Param(scopes.Root, scopes.Block, typedScope); t != nil {
				if tc, ok := t.(*phpdoxer.TypeClassLike); ok {
					return &Resolved{Node: ta.Assignment, Path: scopes.Path},
						tc,
						phprivacy.PrivacyPublic,
						true
				}
			}

		default:
			log.Printf("TODO: resolve variable out of type %T", ta.Scope)
			return nil, nil, 0, false
		}
	}

	return nil, nil, 0, false
}

type nameResolver struct{}

func (p *nameResolver) Down(
	node ir.Node,
) (*DownResolvement, ir.Node, bool) {
	propNode, ok := node.(*ir.Name)
	if !ok {
		return nil, nil, false
	}

	return &DownResolvement{
		ExprType:   ExprTypeName,
		Identifier: symbol.GetIdentifier(propNode),
	}, nil, true
}

func (p *nameResolver) Up(
	scopes Scopes,
	toResolve *DownResolvement,
) (*Resolved, *phpdoxer.TypeClassLike, phprivacy.Privacy, bool) {
	if toResolve.ExprType != ExprTypeName {
		return nil, nil, 0, false
	}

	fqn := common.FullyQualify(scopes.Root, toResolve.Identifier)

	privacy, err := p.DeterminePrivacy(scopes, fqn)
	if err != nil {
		log.Println(fmt.Errorf("[expr.nameResolver.Up]: err determining privacy: %w", err))
		return nil, nil, 0, false
	}

	res, ok := index.FromContainer().Find(fqn)
	if !ok {
		log.Printf("[expr.nameResolver.Up]: unable to find %s in index", fqn)
		return nil, nil, 0, false
	}

	_, n, err := common.SymbolToNode(res.Path, res.Symbol)
	if err != nil {
		log.Println(fmt.Errorf("[expr.nameResolver.Up]: %w", err))
		return nil, nil, 0, false
	}

	return &Resolved{Path: res.Path, Node: n},
		&phpdoxer.TypeClassLike{Name: fqn.String(), FullyQualified: true},
		privacy,
		true
}

func (p *nameResolver) DeterminePrivacy(scopes Scopes, fqn *fqn.FQN) (phprivacy.Privacy, error) {
	// If we are not in a class, it is automatically public access.
	if !symbol.IsClassLike(ir.GetNodeKind(scopes.Class)) {
		return phprivacy.PrivacyPublic, nil
	}

	// If we are in the same class, private access.
	scopeFqn := common.FullyQualify(scopes.Root, symbol.GetIdentifier(scopes.Class))
	if fqn.String() == scopeFqn.String() {
		return phprivacy.PrivacyPrivate, nil
	}

	queue, err := newResolveQueue(
		&phpdoxer.TypeClassLike{Name: scopeFqn.String(), FullyQualified: true},
	)
	if err != nil {
		return 0, err
	}

	// Check if the class is in the scope's resolve queue, that means protected access.
	result := phprivacy.PrivacyPublic
	err = walkResolveQueue(queue, func(wc *walkContext) (done bool, err error) {
		if wc.FQN.String() == fqn.String() {
			result = phprivacy.PrivacyProtected
			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return 0, err
	}

	return result, nil
}

type functionResolver struct{}

func (p *functionResolver) Down(
	node ir.Node,
) (*DownResolvement, ir.Node, bool) {
	propNode, ok := node.(*ir.FunctionCallExpr)
	if !ok {
		return nil, nil, false
	}

	return &DownResolvement{
		ExprType:   ExprTypeFunction,
		Identifier: symbol.GetIdentifier(propNode),
	}, nil, true
}

func (p *functionResolver) Up(
	scopes Scopes,
	toResolve *DownResolvement,
) (*Resolved, *phpdoxer.TypeClassLike, phprivacy.Privacy, bool) {
	if toResolve.ExprType != ExprTypeFunction {
		return nil, nil, 0, false
	}

	t := traversers.NewFunctionCall(toResolve.Identifier)
	scopes.Block.Walk(t)
	if t.Result == nil {
		log.Printf("could not find function matching function call %s", toResolve.Identifier)
		return nil, nil, 0, false
	}

	typeOfFunc := func(n ir.Node) *phpdoxer.TypeClassLike {
		ret := typer.FromContainer().Returns(scopes.Root, n, rootRetriever)
		if clsRet, ok := ret.(*phpdoxer.TypeClassLike); ok {
			return clsRet
		}

		return nil
	}

	// If have local scope, check it for the function.
	if ir.GetNodeKind(scopes.Block) != ir.KindRoot {
		ft := traversers.NewFunction(toResolve.Identifier)
		scopes.Block.Walk(ft)

		if ft.Function != nil {
			return &Resolved{
				Node: ft.Function,
				Path: scopes.Path,
			}, typeOfFunc(ft.Function), phprivacy.PrivacyPublic, true
		}
	}

	// Check for functions defined in the used namespaces.
	if def, ok := common.FindFullyQualified(scopes.Root, toResolve.Identifier, ir.KindFunctionStmt); ok {
		_, n, err := common.SymbolToNode(def.Path, def.Symbol)
		if err != nil {
			log.Println(fmt.Errorf("[expr.functionResolver.Up]: namespace check: %w", err))
		}

		return &Resolved{
			Node: n,
			Path: def.Path,
		}, typeOfFunc(n), phprivacy.PrivacyPublic, true
	}

	// Check for global functions.
	key := fqn.New(fqn.PartSeperator + toResolve.Identifier)
	def, ok := index.FromContainer().Find(key)
	if !ok {
		log.Println(fmt.Errorf("[expr.functionResolver.Up]: unable to find %s in index", key))
		return nil, nil, 0, false
	}

	_, n, err := common.SymbolToNode(def.Path, def.Symbol)
	if err != nil {
		log.Println(fmt.Errorf("[expr.functionResolver.Up]: global check: %w", err))
		return nil, nil, 0, false
	}

	return &Resolved{
		Node: n,
		Path: def.Path,
	}, typeOfFunc(n), phprivacy.PrivacyPublic, true
}

type newResolver struct{}

func (newresolver *newResolver) Down(
	node ir.Node,
) (*DownResolvement, ir.Node, bool) {
	newNode, ok := node.(*ir.NewExpr)
	if !ok {
		return nil, nil, false
	}

	// TODO: new expression using a non-name node
	if name, ok := newNode.Class.(*ir.Name); ok {
		return &DownResolvement{
			ExprType:   ExprTypeNew,
			Identifier: name.Value,
		}, nil, true
	}

	log.Println("TODO: new expression using a non-name node")
	return nil, nil, false
}

func (newresolver *newResolver) Up(
	scopes Scopes,
	toResolve *DownResolvement,
) (resolved *Resolved, nextCtx *phpdoxer.TypeClassLike, privacy phprivacy.Privacy, done bool) {
	if toResolve.ExprType != ExprTypeNew {
		return nil, nil, 0, false
	}

	if fqn := common.FullyQualify(scopes.Root, toResolve.Identifier); fqn != nil {
		def, ok := index.FromContainer().Find(fqn)
		if !ok {
			log.Println(fmt.Errorf("[expr.newResolver.Up]: unable to find %s in index", fqn))
			return nil, nil, 0, false
		}

		_, n, err := common.SymbolToNode(def.Path, def.Symbol)
		if err != nil {
			log.Println(fmt.Errorf("[expr.newResolver.Up]: %w", err))
			return nil, nil, 0, false
		}

		return &Resolved{Path: def.Path, Node: n},
			&phpdoxer.TypeClassLike{Name: fqn.String(), FullyQualified: true},
			phprivacy.PrivacyPublic,
			true
	}

	return nil, nil, 0, false
}

func parentOf(scopes Scopes) *index.INode {
	switch scopes.Class.(type) {
	case *ir.ClassStmt:
		break
	case *ir.TraitStmt:
		// TODO: get parent of trait
		log.Println("TODO: get parent of trait")
		return nil

	default:
		log.Println("encountered parent:: but could not find a class node")
		return nil
	}

	class := scopes.Class.(*ir.ClassStmt)

	if class.Extends == nil {
		log.Println("encountered parent:: but current class does not extend anything")
		return nil
	}

	if node, ok := common.FindFullyQualified(
		scopes.Root,
		class.Extends.ClassName.Value,
		ir.KindClassStmt,
	); ok {
		return node
	}

	log.Printf(
		"could not find fully qualified class for %s in index",
		class.Extends.ClassName.Value,
	)
	return nil
}
