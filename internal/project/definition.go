package project

import (
	"errors"
	"fmt"
	"log"

	"appliedgo.net/what"
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/elephp/pkg/symbol"
	"github.com/laytan/elephp/pkg/traversers"
)

var ErrNoDefinitionFound = errors.New("No definition found for symbol at given position")

// TODO: lots of repetition here.
func (p *Project) Definition(pos *position.Position) (*position.Position, error) {
	content, root, err := p.wrksp.AllOf(pos.Path)
	if err != nil {
		return nil, fmt.Errorf(
			"[ERROR] In definition retrieving file content/nodes for %s: %w",
			pos.Path,
			err,
		)
	}

	apos := position.LocToPos(content, pos.Row, pos.Col)
	nap := traversers.NewNodeAtPos(apos)
	root.Walk(nap)

	for i := len(nap.Nodes) - 1; i >= 0; i-- {
		node := nap.Nodes[i]
		what.Happens("Checking for definition of %T\n", node)

		var scope ir.Node
		var classLikeScope ir.Node
		for j := i - 1; j >= 0; j-- {
			sNode := nap.Nodes[j]
			if scope == nil && (symbol.IsScope(sNode) || ir.GetNodeKind(sNode) == ir.KindRoot) {
				scope = sNode
			}

			if classLikeScope == nil && symbol.IsClassLike(sNode) {
				classLikeScope = sNode
			}
		}

		switch typedNode := node.(type) {
		case *ir.GlobalStmt:
			// TODO: this might index out of bounds
			globalVar, ok := nap.Nodes[i+1].(*ir.SimpleVar)
			if !ok {
				log.Println("Node after the global stmt was not a variable, which we expected")
				return nil, ErrNoDefinitionFound
			}

			assignment := p.globalAssignment(root, globalVar)
			if assignment == nil {
				return nil, ErrNoDefinitionFound
			}

			pos := ir.GetPosition(assignment)
			_, col := position.PosToLoc(content, uint(pos.StartPos))

			return &position.Position{
				Row: uint(pos.StartLine),
				Col: col,
			}, nil

		case *ir.SimpleVar:
			var assignment ir.Node
			switch typedNode.Name {
			case "this":
				if classLikeScope == nil {
					return nil, ErrNoDefinitionFound
				}

				assignment = classLikeScope

			default:
				// UGLLY :(
				if len(nap.Nodes) >= i-2 {
					if _, ok := nap.Nodes[i-1].(*ir.GlobalStmt); ok {
						continue
					}
				}

				if ass := p.assignment(scope, typedNode); ass != nil {
					assignment = ass
				}
			}

			if assignment == nil {
				return nil, ErrNoDefinitionFound
			}

			destPos := ir.GetPosition(assignment)
			_, col := position.PosToLoc(content, uint(destPos.StartPos))

			return &position.Position{
				Row:  uint(destPos.StartLine),
				Col:  col,
				Path: pos.Path,
			}, nil

		case *ir.FunctionCallExpr:
			function, destPath := p.function(scope, typedNode)

			if function == nil {
				return nil, ErrNoDefinitionFound
			}

			if destPath == "" {
				destPath = pos.Path
			}

			destContent, err := p.wrksp.ContentOf(destPath)
			if err != nil {
				log.Println(fmt.Errorf("[ERROR] Definition destination at %s: %w", destPath, err))
				return nil, ErrNoDefinitionFound
			}

			pos := function.Position()
			_, col := position.PosToLoc(destContent, uint(pos.StartPos))

			return &position.Position{
				Row:  uint(pos.StartLine),
				Col:  col,
				Path: destPath,
			}, nil

		case *ir.Name:
			classLike, destPath := p.classLike(root, typedNode)
			if classLike == nil {
				continue
			}

			if destPath == "" {
				destPath = pos.Path
			}

			destContent, err := p.wrksp.ContentOf(destPath)
			if err != nil {
				log.Println(fmt.Errorf("[ERROR] Definition destination at %s: %w", destPath, err))
				return nil, ErrNoDefinitionFound
			}

			pos := classLike.Position()
			_, col := position.PosToLoc(destContent, uint(pos.StartPos))

			return &position.Position{
				Row:  uint(pos.StartLine),
				Col:  col,
				Path: destPath,
			}, nil

		case *ir.MethodCallExpr:
			if len(nap.Nodes) < i {
				log.Println("No nodes found for given position that are more specific than the method call node")
				return nil, ErrNoDefinitionFound
			}

			switch nap.Nodes[i+1].(type) {
			// If one index further is the variable, go to the definition of that variable.
			// when we break here, the next node will be checked and it will match the
			// variable arm of the switch.
			case *ir.SimpleVar:

				// If one index further is an identifier, go to the method definition.
			case *ir.Identifier:
				method, destPath, err := p.method(root, classLikeScope, scope, typedNode)
				if err != nil {
					return nil, err
				}

				if method == nil {
					return nil, ErrNoDefinitionFound
				}

				if destPath == "" {
					destPath = pos.Path
				}

				destContent, err := p.wrksp.ContentOf(destPath)
				if err != nil {
					log.Println(fmt.Errorf("[ERROR] Definition destination at %s: %w", destPath, err))
					return nil, ErrNoDefinitionFound
				}

				pos := ir.GetPosition(method)
				_, col := position.PosToLoc(destContent, uint(pos.StartPos))

				return &position.Position{
					Row:  uint(pos.StartLine),
					Col:  col,
					Path: destPath,
				}, nil
			}

		case *ir.PropertyFetchExpr:
			if len(nap.Nodes) < i {
				log.Println("No nodes found for given position that are more specific than the property fetch node")
				return nil, ErrNoDefinitionFound
			}

			switch nap.Nodes[i+1].(type) {
			// If one index further is the variable, go to the definition of that variable.
			// when we break here, the next node will be checked and it will match the
			// variable arm of the switch.
			case *ir.SimpleVar:

				// If one index further is an identifier, go to the property definition.
			case *ir.Identifier:
				method, destPath, err := p.property(root, classLikeScope, scope, typedNode)
				if err != nil {
					return nil, err
				}

				if method == nil {
					return nil, ErrNoDefinitionFound
				}

				if destPath == "" {
					destPath = pos.Path
				}

				destContent, err := p.wrksp.ContentOf(destPath)
				if err != nil {
					log.Println(fmt.Errorf("[ERROR] Definition destination at %s: %w", destPath, err))
					return nil, ErrNoDefinitionFound
				}

				pos := ir.GetPosition(method)
				_, col := position.PosToLoc(destContent, uint(pos.StartPos))

				return &position.Position{
					Row:  uint(pos.StartLine),
					Col:  col,
					Path: destPath,
				}, nil
			}
		}
	}

	return nil, ErrNoDefinitionFound
}
