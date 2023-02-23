package typer

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/php-parser/pkg/token"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/samber/do"
)

// Typer is responsible of using ir and phpdoxer to retrieve/resolve types
// from phpdoc or type hints of a node.
type Typer interface {
	// Scope should be the method/function the variable is used in, if it is used
	// globally, this can be left nil.
	Variable(root *ir.Root, variable *ir.SimpleVar, scope ir.Node) phpdoxer.Type
}

type typer struct{}

func New() Typer {
	return &typer{}
}

func FromContainer() Typer {
	return do.MustInvoke[Typer](nil)
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
			break
		}

		if typed.Left.Kind() == phpdoxer.KindNull && typed.Right.Kind() == phpdoxer.KindClassLike {
			cl = typed.Right.(*phpdoxer.TypeClassLike)
			break
		}

		return t

	default:
		return t
	}

	if cl.FullyQualified {
		return cl
	}

	tr := fqn.NewTraverserHandlingKeywords(block)
	root.Walk(tr)
	res := tr.ResultFor(&ir.Name{Value: cl.Name, Position: ir.GetPosition(block)})

	cl.FullyQualified = true
	cl.Name = res.String()

	return cl
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
