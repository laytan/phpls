package typer

// Typer is responsible of using ir and phpdoxer to retrieve/resolve types
// from phpdoc or type hints of a node.

import (
	"errors"
	"fmt"
	"log"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/php-parser/pkg/token"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/resolvequeue"
	"github.com/samber/do"
)

type Typer interface {
	// Call with either a ir.ClassMethodStmt or ir.FunctionStmt.
	Returns(
		root *ir.Root,
		funcOrMeth ir.Node,
		rootRetriever func(n *resolvequeue.Node) (*ir.Root, error),
	) phpdoxer.Type

	// Call with either a ir.ClassMethodStmt or ir.FunctionStmt.
	Param(root *ir.Root, funcOrMeth ir.Node, param *ir.Parameter) phpdoxer.Type

	// Scope should be the method/function the variable is used in, if it is used
	// globally, this can be left nil.
	Variable(root *ir.Root, variable *ir.SimpleVar, scope ir.Node) phpdoxer.Type

	Property(root *ir.Root, propertyList *ir.PropertyListStmt) phpdoxer.Type
}

var (
	ErrUnsupportedFuncOrMethod = errors.New("Unsupported function or method node")
	ErrUnexpectedNodeType      = errors.New("Node type unexpected")
)

type typer struct{}

func New() Typer {
	return &typer{}
}

func FromContainer() Typer {
	return do.MustInvoke[Typer](nil)
}

func NodeComments(node ir.Node) []string {
	switch typedNode := node.(type) {
	case *ir.FunctionStmt:
		return []string{typedNode.Doc.Raw}

	case *ir.ArrowFunctionExpr:
		return []string{typedNode.Doc.Raw}

	case *ir.ClosureExpr:
		return []string{typedNode.Doc.Raw}

	case *ir.ClassConstListStmt:
		return []string{typedNode.Doc.Raw}

	case *ir.ClassMethodStmt:
		return []string{typedNode.Doc.Raw}

	case *ir.ClassStmt:
		return []string{typedNode.Doc.Raw}

	case *ir.InterfaceStmt:
		return []string{typedNode.Doc.Raw}

	case *ir.PropertyListStmt:
		return []string{typedNode.Doc.Raw}

	case *ir.TraitStmt:
		return []string{typedNode.Doc.Raw}

	case *ir.FunctionCallExpr:
		return NodeComments(typedNode.Function)

	default:
		docs := []string{}
		node.IterateTokens(func(t *token.Token) bool {
			if t.ID != token.T_COMMENT && t.ID != token.T_DOC_COMMENT {
				return true
			}

			docs = append(docs, string(t.Value))
			return true
		})
		return docs
	}
}

func parseTypeHint(node ir.Node) phpdoxer.Type {
	retNode := returnTypeNode(node)
	if retNode == nil {
		return nil
	}

	name, ok := retNode.(*ir.Name)
	if !ok {
		log.Printf("%T is unsupported for a return type hint, expecting *ir.Name\n", retNode)
		return nil
	}

	t, err := phpdoxer.ParseType(name.Value)
	if err != nil {
		log.Println(fmt.Errorf(`Error parsing return type hint "%s": %w`, name.Value, err))
	}

	return t
}

func returnTypeNode(node ir.Node) ir.Node {
	switch typedNode := node.(type) {
	case *ir.FunctionStmt:
		return typedNode.ReturnType
	case *ir.ClassMethodStmt:
		return typedNode.ReturnType
	case *ir.Parameter:
		return typedNode.VariableType
	case *ir.PropertyListStmt:
		return typedNode.Type
	default:
		return nil
	}
}

func resolveFQN(root *ir.Root, block ir.Node, t phpdoxer.Type) phpdoxer.Type {
	var cl *phpdoxer.TypeClassLike
	switch typed := t.(type) {
	case *phpdoxer.TypeClassLike:
		cl = typed

	case *phpdoxer.TypeUnion:
		// Basic check if it is a union between a type and null, ignore null then.
		// NOTE: this is ideal for our use case, but other applications might want to know about the null.
		// TODO: it might be better to move this one package up (thinking about other use cases).

		if typed.Left.Kind() == phpdoxer.KindClassLike && typed.Right.Kind() == phpdoxer.KindNull {
			cl = typed.Left.(*phpdoxer.TypeClassLike)
		} else if typed.Left.Kind() == phpdoxer.KindNull && typed.Right.Kind() == phpdoxer.KindClassLike {
			cl = typed.Right.(*phpdoxer.TypeClassLike)
		}

	default:
		return t
	}

	if cl.FullyQualified {
		return cl
	}

	tr := fqn.NewFQNTraverserHandlingKeywords(block)
	root.Walk(tr)
	res := tr.ResultFor(&ir.Name{Value: cl.Name})

	cl.FullyQualified = true
	cl.Name = res.String()

	return cl
}
