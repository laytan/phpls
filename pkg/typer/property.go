package typer

import (
	"log"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/phpdoxer"
)

// TODO: support {@inheritdoc}.
func (typer *typer) Property(root *ir.Root, propertyList *ir.PropertyListStmt) phpdoxer.Type {
	nodes, err := phpdoxer.ParseDoc(propertyList.Doc.Raw)
	if err != nil {
		log.Println(err)
	}

	for _, node := range nodes {
		varNode, ok := node.(*phpdoxer.NodeVar)
		if ok {
			resolveFQN(root, propertyList, varNode.Type)
			return varNode.Type
		}
	}

	if hintType := parseTypeHint(propertyList); hintType != nil {
		resolveFQN(root, propertyList, hintType)
		return hintType
	}

	return nil
}
