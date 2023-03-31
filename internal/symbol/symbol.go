package symbol

import (
	"log"

	"github.com/laytan/elephp/pkg/nodeident"
	"github.com/laytan/php-parser/pkg/ast"
)

func nodeToName(node ast.Vertex) ast.Vertex {
	switch typed := node.(type) {
	case *ast.Name, *ast.NameFullyQualified, *ast.NameRelative:
		return typed
	case *ast.Identifier:
		return &ast.Name{Position: typed.Position, Parts: []ast.Vertex{&ast.NamePart{Value: []byte(nodeident.Get(typed))}}}
	default:
		log.Panicf("[symbol.nodeToName]: expected type %T to be *ast.Name|*ast.Identifier\n", node)
		return nil
	}
}
