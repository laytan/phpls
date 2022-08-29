package typer

// Typer is responsible of using ir and phpdoxer to retrieve/resolve types
// from phpdoc or type hints of a node.

import (
	"errors"
	"fmt"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/php-parser/pkg/token"
	"github.com/laytan/elephp/pkg/phpdoxer"
	log "github.com/sirupsen/logrus"
)

type Typer interface {
	// Call with either a ir.ClassMethodStmt or ir.FunctionStmt.
	Returns(root *ir.Root, funcOrMeth ir.Node) phpdoxer.Type

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

func NodeComments(node ir.Node) []string {
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

func parseTypeHint(node ir.Node) phpdoxer.Type {
	retNode := returnTypeNode(node)
	if retNode == nil {
		return nil
	}

	name, ok := retNode.(*ir.Name)
	if !ok {
		log.Errorf("%T is unsupported for a return type hint, expecting *ir.Name\n", retNode)
		return nil
	}

	t, err := phpdoxer.ParseType(name.Value)
	if err != nil {
		log.Error(fmt.Errorf(`Error parsing return type hint "%s": %w`, name.Value, err))
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

func resolveFQN(root *ir.Root, t phpdoxer.Type) {
	cl, ok := t.(*phpdoxer.TypeClassLike)
	if !ok {
		return
	}

	if cl.FullyQualified {
		return
	}

	tr := NewFQNTraverser()
	root.Walk(tr)
	res := tr.ResultFor(&ir.Name{Value: cl.Name})

	cl.FullyQualified = true
	cl.Name = res.String()
}
