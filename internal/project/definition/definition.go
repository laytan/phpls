package definition

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/phprivacy"
	"github.com/laytan/elephp/pkg/symbol"
	"github.com/laytan/elephp/pkg/traversers"
	"github.com/laytan/elephp/pkg/typer"
)

const (
	ErrUnexpectedNodeFmt = "Unexpected node of type %T (expected %s) when retrieving property definition"
)

var (
	ErrUnexpectedNode    = fmt.Errorf(ErrUnexpectedNodeFmt, nil, "")
	ErrNoDefinitionFound = errors.New(
		"No definition found for symbol at given position",
	)
)

type Definition struct {
	Path string
	Node symbol.Symbol
}

func FullyQualify(root *ir.Root, name string) *typer.FQN {
	if strings.HasPrefix(name, `\`) {
		return typer.NewFQN(name)
	}

	t := typer.NewFQNTraverser()
	root.Walk(t)

	return t.ResultFor(&ir.Name{Value: name})
}

func FindFullyQualified(
	root *ir.Root,
	i index.Index,
	name string,
	kinds ...ir.NodeKind,
) (*Definition, bool) {
	FQN := FullyQualify(root, name)
	node, err := i.Find(FQN.String(), kinds...)
	if err != nil {
		if !errors.Is(err, index.ErrNotFound) {
			log.Println(err)
		}

		return nil, false
	}

	return &Definition{Node: node.Symbol, Path: node.Path}, true
}

// TODO: can we merge this into the typer.Variable function?
// TODO: or at least in that package, (or another one, it doesn't fit here)?
func VariableType(
	ctx context.Context,
	variable *ir.SimpleVar,
) (*Definition, phprivacy.Privacy, error) {
	switch variable.Name {
	case "this", "self", "static":
		def, ok := FindFullyQualified(
			ctx.Root(),
			ctx.Index(),
			symbol.GetIdentifier(ctx.ClassScope()),
			symbol.ClassLikeScopes...)
		if !ok {
			return nil, 0, ErrNoDefinitionFound
		}

		return def, phprivacy.PrivacyPrivate, nil

	case "parent":
		parent, err := parentOf(ctx, ctx.ClassScope())
		if err != nil {
			return nil, 0, err
		}

		return parent, phprivacy.PrivacyProtected, nil

	default:
		// TODO: I don't like that we have to create traversers here.
		// TODO: this should ideally call the variableProvider.
		assT := traversers.NewAssignment(variable)
		ctx.Scope().Walk(assT)
		if assT.Assignment == nil || assT.Scope == nil {
			return nil, 0, ErrNoDefinitionFound
		}

		// If scope is a parameter, see if the scope(function/method, passed into this func) has doc with @param,
		// If not, see if the param has type hints,
		// If not, we don't know the type.
		switch varScope := assT.Scope.(type) {
		case *ir.Parameter:
			switch ctx.Scope().(type) {
			case *ir.FunctionStmt, *ir.ClassMethodStmt:
				paramType := ctx.Typer().Param(ctx.Root(), ctx.Scope(), varScope)
				if typedRetType, ok := paramType.(*phpdoxer.TypeClassLike); ok {
					def, ok := FindFullyQualified(ctx.Root(), ctx.Index(), typedRetType.Name, symbol.ClassLikeScopes...)
					if !ok {
						return nil, 0, ErrNoDefinitionFound
					}

					return def, phprivacy.PrivacyPublic, nil
				}

				return nil, 0, ErrNoDefinitionFound

			default:
				return nil, 0, fmt.Errorf(ErrUnexpectedNodeFmt, ctx.Scope(), "*ir.FunctionStmt or *ir.ClassMethodStmt")
			}

		case *ir.Assign:
			// If scope is assign, check traverser.assignment for a phpdoc.
			// If it has a phpdoc, with @var, return the symbol for that type.
			varType := ctx.Typer().Variable(ctx.Root(), assT.Assignment, ctx.Scope())
			if clsLike, ok := varType.(*phpdoxer.TypeClassLike); ok {
				def, ok := FindFullyQualified(ctx.Root(), ctx.Index(), clsLike.Name, symbol.ClassLikeScopes...)
				if !ok {
					return nil, 0, ErrNoDefinitionFound
				}

				// TODO: privacy will be wrong in some cases.
				return def, phprivacy.PrivacyPrivate, nil
			}

			switch exprNode := varScope.Expr.(type) {
			// Do this recursively, with .Expr:
			case *ir.CloneExpr, *ir.ParenExpr, *ir.ErrorSuppressExpr:
				// Same but its a Node:
			case *ir.ReferenceExpr:
				// Functions/methods to get return type from:
			case *ir.ClosureExpr, *ir.MethodCallExpr, *ir.ArrowFunctionExpr, *ir.FunctionCallExpr, *ir.StaticCallExpr, *ir.NullsafeMethodCallExpr:
				// Return left or right type (union):
			case *ir.CoalesceExpr, *ir.TernaryExpr:
				// Return the type being instantiated:
			case *ir.NewExpr:
				if className, ok := exprNode.Class.(*ir.Name); ok {
					def, ok := FindFullyQualified(ctx.Root(), ctx.Index(), className.Value, ir.KindClassStmt)
					if !ok {
						return nil, 0, ErrNoDefinitionFound
					}

					return def, phprivacy.PrivacyPublic, nil
				}

				return nil, 0, fmt.Errorf(ErrUnexpectedNodeFmt, exprNode.Class, "*ir.Name")

				// Fetch properties/variables:
			case *ir.ConstFetchExpr, *ir.PropertyFetchExpr, *ir.ClassConstFetchExpr, *ir.NullsafePropertyFetchExpr, *ir.StaticPropertyFetchExpr:
				// Return the type being casted:
			case *ir.TypeCastExpr:
			// Don't know how:
			case *ir.MatchExpr, *ir.ArrayDimFetchExpr, *ir.AnonClassExpr:
			default:
				return nil, 0, ErrNoDefinitionFound

			}
		}

		return nil, 0, ErrNoDefinitionFound
	}
}

func parentOf(ctx context.Context, classStmt ir.Node) (*Definition, error) {
	switch classStmt.(type) {
	case *ir.ClassStmt:
		break
	case *ir.TraitStmt:
		// Could be called in a trait, but we have no way of knowing what the parent is.
		return nil, ErrNoDefinitionFound

	default:
		return nil, fmt.Errorf(ErrUnexpectedNodeFmt, classStmt, "*ir.ClassStmt")
	}

	class := classStmt.(*ir.ClassStmt)

	if class.Extends == nil {
		return nil, errors.New("Unexpected call of parent:: in a class without a parent")
	}

	fqn := FullyQualify(ctx.Root(), class.Extends.ClassName.Value)
	node, err := ctx.Index().Find(fqn.String(), ir.KindClassStmt)
	if err != nil {
		return nil, fmt.Errorf("Parent class %s is not indexed: %w", fqn.String(), err)
	}

	return &Definition{Node: node.Symbol, Path: node.Path}, nil
}
