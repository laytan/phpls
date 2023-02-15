package symbol

import (
	"log"

	"github.com/VKCOM/noverify/src/ir"
)

// nodeToName returns an *ir.Name when passed an ir.Node of type *ir.Name or
// *ir.Identifier.
func nodeToName(node ir.Node) *ir.Name {
	switch typed := node.(type) {
	case *ir.Name:
		return typed
	case *ir.Identifier:
		return &ir.Name{Position: typed.Position, Value: typed.Value}
	default:
		log.Panicf("[symbol.nodeToName]: expected type %T to be *ir.Name|*ir.Identifier\n", node)
		return nil
	}
}
