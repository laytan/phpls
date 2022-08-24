package typer

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/phpdoxer"
	log "github.com/sirupsen/logrus"
)

func (t *typer) Variable(
	root *ir.Root,
	variable *ir.SimpleVar,
	scope ir.Node,
) phpdoxer.Type {
	if cmntType := findVarComment(variable); cmntType != nil {
		resolveFQN(root, cmntType)
		return cmntType
	}

	return nil
}

func findVarComment(node ir.Node) phpdoxer.Type {
	comments := NodeComments(node)
	for _, comment := range comments {
		nodes, err := phpdoxer.ParseDoc(comment)
		if err != nil {
			log.Warn(err)
			continue
		}

		for _, node := range nodes {
			param, ok := node.(*phpdoxer.NodeVar)
			if !ok {
				continue
			}

			return param.Type
		}
	}

	return nil
}
