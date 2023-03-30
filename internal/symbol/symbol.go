package symbol

import (
	"log"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/laytan/elephp/pkg/nodeident"
)

func nodeToName(node ast.Vertex) *ast.Name {
	switch typed := node.(type) {
	case *ast.Name:
		return typed
	case *ast.Identifier:
		return &ast.Name{Position: typed.Position, Parts: []ast.Vertex{&ast.NamePart{Value: []byte(nodeident.Get(typed))}}}
	default:
		log.Panicf("[symbol.nodeToName]: expected type %T to be *ast.Name|*ast.Identifier\n", node)
		return nil
	}
}
