package typer

import (
	"errors"
	"fmt"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/php-parser/pkg/token"
	"github.com/laytan/elephp/pkg/phpdoxer"
)

type Typer interface {
	// Call with either a ir.ClassMethodStmt or ir.FunctionStmt.
	Returns(root *ir.Root, funcOrMeth ir.Node) (Type, error)

	// Call with either a ir.ClassMethodStmt or ir.FunctionStmt.
	Param(root *ir.Root, funcOrMeth ir.Node, param *ir.Parameter) (Type, error)

	// Scope should be the method/function the variable is used in, if it is used
	// globally, this can be left nil.
	Variable(root *ir.Root, variable *ir.SimpleVar, scope ir.Node) (Type, error)
}

var (
	ErrUnsupportedFuncOrMethod = errors.New("Unsupported function or method node")
	ErrUnexpectedNodeType      = errors.New("Node type unexpected")
)

type typer struct {
	doxer phpdoxer.PhpDoxer
}

func (t *typer) Returns(root *ir.Root, funcOrMeth ir.Node) (Type, error) {
	kind := ir.GetNodeKind(funcOrMeth)
	if kind != ir.KindClassMethodStmt && kind != ir.KindFunctionStmt {
		return nil, fmt.Errorf("Type %T: %w", funcOrMeth, ErrUnsupportedFuncOrMethod)
	}

	docReturn := t.docIterReturn(funcOrMeth)
	if len(docReturn) > 0 {
		// TODO: Give class
		parsedT, err := ParseUnion(docReturn)
		if err != nil {
			return nil, err
		}

		return parsedT, nil
	}

	hintReturn := t.returnTypeHint(funcOrMeth)
	if hintReturn == nil {
		return &TypeMixed{}, nil
	}

	// TODO: a scalar (booll, string etc.) is not a name node.
	// TODO: can a typehint be a union?
	name, ok := hintReturn.(*ir.Name)
	if !ok {
		return nil, fmt.Errorf(
			"expected type *ir.Name but got %T: %w",
			hintReturn,
			ErrUnexpectedNodeType,
		)
	}

	trav := NewFQNTraverser()
	root.Walk(trav)
	return &TypeClassLike{Name: trav.ResultFor(name).String()}, nil
}

func (t *typer) Param(Root *ir.Root, funcOrMeth ir.Node, param *ir.Parameter) (Type, error) {
	panic("not implemented") // TODO: Implement
}

func (t *typer) Variable(Root *ir.Root, variable *ir.SimpleVar, scope ir.Node) (Type, error) {
	panic("not implemented") // TODO: Implement
}

func (t *typer) docIterReturn(funcOrMeth ir.Node) []string {
	var docReturn []string
	docIterator := func(tok *token.Token) bool {
		if tok.ID != token.T_COMMENT && tok.ID != token.T_DOC_COMMENT {
			return true
		}

		ret := t.doxer.Return(string(tok.Value))
		if len(ret) > 0 {
			docReturn = ret
		}

		return true
	}

	switch typedNode := funcOrMeth.(type) {
	case *ir.FunctionStmt:
		typedNode.IterateTokens(docIterator)

	case *ir.ClassMethodStmt:
		typedNode.IterateTokens(docIterator)

	default:
		panic("Unsupported node in docIterReturn call")
	}

	return docReturn
}

func (t *typer) returnTypeHint(funcOrMeth ir.Node) ir.Node {
	var returnType ir.Node

	switch typedNode := funcOrMeth.(type) {
	case *ir.FunctionStmt:
		returnType = typedNode.ReturnType

	case *ir.ClassMethodStmt:
		returnType = typedNode.ReturnType
	}

	return returnType
}
