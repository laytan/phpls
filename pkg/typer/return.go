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

	if retHint := parseTypeHint(funcOrMeth); retHint != nil {
		resolveFQN(root, retHint)
		return retHint
	}

	return nil
}

func findReturnComment(node ir.Node) phpdoxer.Type {
	comments := NodeComments(node)
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
