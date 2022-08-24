package typer

import (
	"fmt"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/phpdoxer"
	log "github.com/sirupsen/logrus"
)

func (t *typer) Param(
	root *ir.Root,
	funcOrMeth ir.Node,
	param *ir.Parameter,
) phpdoxer.Type {
	kind := ir.GetNodeKind(funcOrMeth)
	if kind != ir.KindClassMethodStmt && kind != ir.KindFunctionStmt {
		panic(fmt.Errorf("Type: %T: %w", funcOrMeth, ErrUnexpectedNodeType))
	}

	if cmntType := findParamComment(funcOrMeth, param.Variable.Name); cmntType != nil {
		resolveFQN(root, cmntType)
		return cmntType
	}

	if hintType := parseTypeHint(param); hintType != nil {
		resolveFQN(root, hintType)
		return hintType
	}

	return nil
}

func findParamComment(node ir.Node, name string) phpdoxer.Type {
	comments := NodeComments(node)
	for _, comment := range comments {
		nodes, err := phpdoxer.ParseDoc(comment)
		if err != nil {
			log.Warn(err)
			continue
		}

		for _, node := range nodes {
			param, ok := node.(*phpdoxer.NodeParam)
			if !ok {
				continue
			}

			if param.Name != name {
				continue
			}

			return param.Type
		}
	}

	return nil
}
