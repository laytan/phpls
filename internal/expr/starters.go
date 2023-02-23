package expr

import (
	"errors"
	"fmt"
	"log"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/fqner"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/symbol"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/ie"
	"github.com/laytan/elephp/pkg/nodeident"
	"github.com/laytan/elephp/pkg/nodescopes"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/phprivacy"
	"github.com/laytan/elephp/pkg/traversers"
)

var starters = map[Type]StartResolver{
	TypeVariable: &variableResolver{},
	TypeName:     &nameResolver{},
	TypeFunction: &functionResolver{},
	TypeNew:      &newResolver{},
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
		ExprType:   TypeVariable,
		Identifier: nodeident.Get(propNode),
		Position:   propNode.Position,
	}, nil, true
}

func (p *variableResolver) Up(
	scopes *Scopes,
	toResolve *DownResolvement,
) (*Resolved, *fqn.FQN, phprivacy.Privacy, bool) {
	wrk := wrkspc.FromContainer()

	if toResolve.ExprType != TypeVariable {
		return nil, nil, 0, false
	}

	switch toResolve.Identifier {
	case "this", "self", "static":
		if node, ok := fqner.FindFullyQualifiedName(scopes.Root, &ir.Name{
			Position: ir.GetPosition(scopes.Class),
			Value:    nodeident.Get(scopes.Class),
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

		cls, _ := fqner.FindFullyQualifiedName(scopes.Root, &ir.Name{
			Position: ir.GetPosition(scopes.Class),
			Value:    nodeident.Get(scopes.Class),
		})

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

		switch typedScope := ta.Scope.(type) {
		case *ir.Assign:
			if res, lastClass, left := Resolve(typedScope.Expr, scopes); left == 0 {
				return res, lastClass, phprivacy.PrivacyPublic, true
			}

		case *ir.Parameter:
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
	node ir.Node,
) (*DownResolvement, ir.Node, bool) {
	propNode, ok := node.(*ir.Name)
	if !ok {
		return nil, nil, false
	}

	return &DownResolvement{
		ExprType:   TypeName,
		Identifier: nodeident.Get(propNode),
		Position:   propNode.Position,
	}, nil, true
}

func (p *nameResolver) Up(
	scopes *Scopes,
	toResolve *DownResolvement,
) (*Resolved, *fqn.FQN, phprivacy.Privacy, bool) {
	if toResolve.ExprType != TypeName {
		return nil, nil, 0, false
	}

	qualified := fqner.FullyQualifyName(scopes.Root, &ir.Name{
		Position: toResolve.Position,
		Value:    toResolve.Identifier,
	})

	privacy, err := p.DeterminePrivacy(scopes, qualified)
	if err != nil {
		log.Println(fmt.Errorf("[expr.nameResolver.Up]: err determining privacy: %w", err))
		return nil, nil, 0, false
	}

	res, ok := index.FromContainer().Find(qualified)
	if !ok {
		log.Printf("[expr.nameResolver.Up]: unable to find %s in index", qualified)
		return nil, nil, 0, false
	}

	return &Resolved{Path: res.Path, Node: res.ToIRNode(wrkspc.FromContainer().FIROf(res.Path))},
		qualified,
		privacy,
		true
}

func (p *nameResolver) DeterminePrivacy(scopes *Scopes, fqn *fqn.FQN) (phprivacy.Privacy, error) {
	// If we are not in a class, it is automatically public access.
	if !nodescopes.IsClassLike(ir.GetNodeKind(scopes.Class)) {
		return phprivacy.PrivacyPublic, nil
	}

	// If we are in the same class, private access.
	scopeFqn := fqner.FullyQualifyName(scopes.Root, &ir.Name{
		Position: ir.GetPosition(scopes.Class),
		Value:    nodeident.Get(scopes.Class),
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
	node ir.Node,
) (*DownResolvement, ir.Node, bool) {
	propNode, ok := node.(*ir.FunctionCallExpr)
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
	scopes.Block.Walk(t)
	if t.Result == nil {
		log.Printf("could not find function matching function call %s", toResolve.Identifier)
		return nil, nil, 0, false
	}

	typeOfFunc := func(n ir.Node) *fqn.FQN {
		ret, _, err := symbol.NewFunction(
			wrkspc.NewRooter(scopes.Path, scopes.Root),
			n.(*ir.FunctionStmt),
		).Returns()

		if err != nil && !errors.Is(err, symbol.ErrNoReturn) {
			log.Println(fmt.Errorf("getting return type of %v: %w", n, err))
		}

		if retCls, ok := ret.(*phpdoxer.TypeClassLike); ok {
			return fqner.FullyQualifyName(scopes.Root, &ir.Name{
				Position: ir.GetPosition(n),
				Value:    retCls.Name,
			})
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
	if def, ok := fqner.FindFullyQualifiedName(scopes.Root, &ir.Name{
		Position: toResolve.Position,
		Value:    toResolve.Identifier,
	}); ok {
		n := def.ToIRNode(wrkspc.FromContainer().FIROf(def.Path))
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

	n := def.ToIRNode(wrkspc.FromContainer().FIROf(def.Path))
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
			ExprType:   TypeNew,
			Identifier: name.Value,
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
		&ir.Name{Position: toResolve.Position, Value: toResolve.Identifier},
	); qualified != nil {
		def, ok := index.FromContainer().Find(qualified)
		if !ok {
			log.Println(fmt.Errorf("[expr.newResolver.Up]: unable to find %s in index", qualified))
			return nil, nil, 0, false
		}

		return &Resolved{
				Path: def.Path,
				Node: def.ToIRNode(wrkspc.FromContainer().FIROf(def.Path)),
			},
			qualified,
			phprivacy.PrivacyPublic,
			true
	}

	return nil, nil, 0, false
}

func parentOf(scopes *Scopes) *index.INode {
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

	if node, ok := fqner.FindFullyQualifiedName(
		scopes.Root,
		class.Extends.ClassName,
	); ok {
		return node
	}

	log.Printf(
		"could not find fully qualified class for %s in index",
		class.Extends.ClassName.Value,
	)
	return nil
}
