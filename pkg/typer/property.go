package typer

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/phpdoxer"
	log "github.com/sirupsen/logrus"
)

func (typer *typer) Property(root *ir.Root, propertyList *ir.PropertyListStmt) phpdoxer.Type {
	nodes, err := phpdoxer.ParseDoc(propertyList.Doc.Raw)
	if err != nil {
		log.Warn(err)
	}

	for _, node := range nodes {
		varNode, ok := node.(*phpdoxer.NodeVar)
		if ok {
			resolveFQN(root, varNode.Type)
			return varNode.Type
		}
	}

	if hintType := parseTypeHint(propertyList); hintType != nil {
		resolveFQN(root, hintType)
		return hintType
	}

	return nil
}
