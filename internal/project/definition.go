package project

import (
	"errors"
	"fmt"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/elephp/pkg/symbol"
	"github.com/laytan/elephp/pkg/traversers"
	log "github.com/sirupsen/logrus"
)

var ErrNoDefinitionFound = errors.New("No definition found for symbol at given position")

func (p *Project) Definition(pos *position.Position) (*position.Position, error) {
	path := pos.Path

	file := p.GetFile(path)
	if file == nil {
		return nil, fmt.Errorf("Error retrieving file content for %s", path)
	}

	ast := p.ParseFileCached(file)
	if ast == nil {
		return nil, fmt.Errorf("Error parsing %s for definitions", path)
	}

	apos := position.LocToPos(file.content, pos.Row, pos.Col)
	nap := traversers.NewNodeAtPos(apos)
	ast.Walk(nap)

	// Root.
	var scope ir.Node
	var classLikeScope ir.Node
	for i, node := range nap.Nodes {
		if symbol.IsScope(node) {
			scope = node
		}

		if symbol.IsClassLike(node) {
			classLikeScope = node
		}

		// fmt.Printf("%T\n", node)
		switch typedNode := node.(type) {
		case *ir.Root:
			scope = typedNode

		case *ir.FunctionStmt:
			scope = typedNode

		case *ir.GlobalStmt:
			// TODO: this might index out of bounds
			globalVar, ok := nap.Nodes[i+1].(*ir.SimpleVar)
			if !ok {
				log.Errorln("Node after the global stmt was not a variable, which we expected")
				return nil, ErrNoDefinitionFound
			}

			assignment := p.globalAssignment(ast, globalVar)
			if assignment == nil {
				return nil, ErrNoDefinitionFound
			}

			pos := ir.GetPosition(assignment)
			_, col := position.PosToLoc(file.content, uint(pos.StartPos))

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
				if ass := p.assignment(scope, typedNode); ass != nil {
					assignment = ass
				}
			}

			if assignment == nil {
				return nil, ErrNoDefinitionFound
			}

			pos := ir.GetPosition(assignment)
			_, col := position.PosToLoc(file.content, uint(pos.StartPos))

			return &position.Position{
				Row:  uint(pos.StartLine),
				Col:  col,
				Path: path,
			}, nil

		case *ir.FunctionCallExpr:
			function, destPath := p.function(scope, typedNode)

			if function == nil {
				return nil, ErrNoDefinitionFound
			}

			if destPath == "" {
				destPath = path
			}

			destFile := p.GetFile(destPath)
			if destFile == nil {
				log.Errorf("Destination at %s is not in parsed files cache\n", path)
				return nil, ErrNoDefinitionFound
			}

			pos := function.Position()
			_, col := position.PosToLoc(destFile.content, uint(pos.StartPos))

			return &position.Position{
				Row:  uint(pos.StartLine),
				Col:  col,
				Path: destPath,
			}, nil

		case *ir.Name:
			classLike, destPath := p.classLike(ast, typedNode)
			if classLike == nil {
				return nil, ErrNoDefinitionFound
			}

			if destPath == "" {
				destPath = path
			}

			destFile := p.GetFile(destPath)
			if destFile == nil {
				log.Errorf("Destination at %s is not in parsed files cache\n", destPath)
				return nil, ErrNoDefinitionFound
			}

			pos := classLike.Position()
			_, col := position.PosToLoc(destFile.content, uint(pos.StartPos))

			return &position.Position{
				Row:  uint(pos.StartLine),
				Col:  col,
				Path: destPath,
			}, nil

		case *ir.MethodCallExpr:
			if len(nap.Nodes) < i {
				log.Errorln("No nodes found for given position that are more specific than the method call node")
				return nil, ErrNoDefinitionFound
			}

			switch nap.Nodes[i+1].(type) {
			// If one index further is the variable, go to the definition of that variable.
			// when we break here, the next node will be checked and it will match the
			// variable arm of the switch.
			case *ir.SimpleVar:

				// If one index further is an identifier, go to the method definition.
			case *ir.Identifier:
				method, destPath, err := p.method(ast, classLikeScope, scope, typedNode)
				if err != nil {
					return nil, err
				}

				if method == nil {
					return nil, ErrNoDefinitionFound
				}

				if destPath == "" {
					destPath = path
				}

				file := p.GetFile(destPath)
				if file == nil {
					return nil, ErrNoDefinitionFound
				}

				pos := ir.GetPosition(method)
				_, col := position.PosToLoc(file.content, uint(pos.StartPos))

				return &position.Position{
					Row:  uint(pos.StartLine),
					Col:  col,
					Path: destPath,
				}, nil
			}

		case *ir.PropertyFetchExpr:
			if len(nap.Nodes) < i {
				log.Errorln("No nodes found for given position that are more specific than the property fetch node")
				return nil, ErrNoDefinitionFound
			}

			switch nap.Nodes[i+1].(type) {
			// If one index further is the variable, go to the definition of that variable.
			// when we break here, the next node will be checked and it will match the
			// variable arm of the switch.
			case *ir.SimpleVar:

				// If one index further is an identifier, go to the property definition.
			case *ir.Identifier:
				method, destPath, err := p.property(ast, classLikeScope, scope, typedNode)
				if err != nil {
					return nil, err
				}

				if method == nil {
					return nil, ErrNoDefinitionFound
				}

				if destPath == "" {
					destPath = path
				}

				file := p.GetFile(destPath)
				if file == nil {
					return nil, ErrNoDefinitionFound
				}

				pos := ir.GetPosition(method)
				_, col := position.PosToLoc(file.content, uint(pos.StartPos))

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
