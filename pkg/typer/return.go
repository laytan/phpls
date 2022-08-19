package typer

import (
	"fmt"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/phpdoxer"
	log "github.com/sirupsen/logrus"
)

// Returns the return type of the method or function, prioritizing phpdoc
// @return over the return type hint.
func (t *typer) Returns(root *ir.Root, funcOrMeth ir.Node) phpdoxer.Type {
	kind := ir.GetNodeKind(funcOrMeth)
	if kind != ir.KindClassMethodStmt && kind != ir.KindFunctionStmt {
		panic(fmt.Errorf("Type: %T: %w", funcOrMeth, ErrUnexpectedNodeType))
	}

	if retDoc := findReturnComment(funcOrMeth); retDoc != nil {
		resolveFQN(root, retDoc)
		return retDoc
	}

	if retHint := findReturnHint(funcOrMeth); retHint != nil {
		resolveFQN(root, retHint)
		return retHint
	}

	return nil
}

func findReturnComment(node ir.Node) phpdoxer.Type {
	comments := nodeComments(node)
	for _, comment := range comments {
		nodes, err := phpdoxer.ParseDoc(comment)
		if err != nil {
			log.Warn(err)
			continue
		}

		for _, node := range nodes {
			if node.Kind() != phpdoxer.KindReturn {
				continue
			}

			return node.(*phpdoxer.NodeReturn).Type
		}
	}

	return nil
}

func findReturnHint(node ir.Node) phpdoxer.Type {
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
