package project

import (
	"errors"
	"fmt"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/traversers"
	"github.com/laytan/elephp/pkg/position"
	log "github.com/sirupsen/logrus"
)

var ErrNoDefinitionFound = errors.New("No definition found for symbol at given position")

func (p *Project) Definition(path string, pos *position.Position) (*position.Position, error) {
	file := p.GetFile(path)
	if file == nil {
		return nil, fmt.Errorf("Error retrieving file content for %s", path)
	}

	ast, err := file.parse(p.parserConfig)
	if err != nil {
		return nil, fmt.Errorf("Error parsing %s for definitions: %w", path, err)
	}

	apos := position.LocToPos(file.content, pos.Row, pos.Col)
	nap := traversers.NewNodeAtPos(apos)
	ast.Walk(nap)

	// Root.
	var scope ir.Node
	for i, node := range nap.Nodes {
		// fmt.Printf("%T\n", node)
		switch typedNode := node.(type) {
		case *ir.Root:
			scope = typedNode

		case *ir.FunctionStmt:
			scope = typedNode

		case *ir.GlobalStmt:
			rootNode, ok := nap.Nodes[0].(*ir.Root)
			if !ok {
				log.Errorln("First node was not the root node, this should not happen")
				return nil, ErrNoDefinitionFound
			}

			globalVar, ok := nap.Nodes[i+1].(*ir.SimpleVar)
			if !ok {
				log.Errorln("Node after the global stmt was not a variable, which we expected")
				return nil, ErrNoDefinitionFound
			}

			assignment := p.globalAssignment(rootNode, globalVar)
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
			assignment := p.assignment(scope, typedNode)
			if assignment == nil {
				return nil, ErrNoDefinitionFound
			}

			pos := ir.GetPosition(assignment)
			_, col := position.PosToLoc(file.content, uint(pos.StartPos))

			return &position.Position{
				Row: uint(pos.StartLine),
				Col: col,
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
			rootNode, ok := nap.Nodes[0].(*ir.Root)
			if !ok {
				log.Errorln("First node was not the root node, this should not happen")
				return nil, ErrNoDefinitionFound
			}

			classLike, destPath := p.classLike(file, rootNode, typedNode)
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
		}
	}

	return nil, ErrNoDefinitionFound
}
