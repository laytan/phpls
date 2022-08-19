package typer

// Typer is responsible of using ir and phpdoxer to retrieve/resolve types
// from phpdoc or type hints of a node.

import (
	"errors"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/php-parser/pkg/token"
	"github.com/laytan/elephp/pkg/phpdoxer"
)

type Typer interface {
	// Call with either a ir.ClassMethodStmt or ir.FunctionStmt.
	Returns(root *ir.Root, funcOrMeth ir.Node) (phpdoxer.Type, error)

	// Call with either a ir.ClassMethodStmt or ir.FunctionStmt.
	Param(root *ir.Root, funcOrMeth ir.Node, param *ir.Parameter) (phpdoxer.Type, error)

	// Scope should be the method/function the variable is used in, if it is used
	// globally, this can be left nil.
	Variable(root *ir.Root, variable *ir.SimpleVar, scope ir.Node) (phpdoxer.Type, error)
}

var (
	ErrUnsupportedFuncOrMethod = errors.New("Unsupported function or method node")
	ErrUnexpectedNodeType      = errors.New("Node type unexpected")
)

type typer struct{}

func (t *typer) Param(
	Root *ir.Root,
	funcOrMeth ir.Node,
	param *ir.Parameter,
) (phpdoxer.Type, error) {
	panic("not implemented") // TODO: Implement
}

func (t *typer) Variable(
	Root *ir.Root,
	variable *ir.SimpleVar,
	scope ir.Node,
) (phpdoxer.Type, error) {
	panic("not implemented") // TODO: Implement
}

func nodeComments(node ir.Node) []string {
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
