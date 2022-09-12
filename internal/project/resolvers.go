package project

import (
	"fmt"
	"log"
	"strings"

	"github.com/laytan/elephp/internal/project/definition"
	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/elephp/pkg/traversers"
)

// Returns the position for the namespace statement that matches the given position.
func (p *Project) Namespace(pos *position.Position) *position.Position {
	content, root, err := p.wrksp.AllOf(pos.Path)
	if err != nil {
		log.Println(
			fmt.Errorf(
				"Getting namespace, could not get content/nodes of %s: %w",
				pos.Path,
				err,
			),
		)
	}

	traverser := traversers.NewNamespace(pos.Row)
	root.Walk(traverser)

	if traverser.Result == nil {
		log.Println("Did not find namespace")
		return nil
	}

	row, col := position.PosToLoc(content, uint(traverser.Result.Position.StartPos))

	return &position.Position{
		Row:  row,
		Col:  col,
		Path: pos.Path,
	}
}

// Returns whether the file at given pos needs a use statement for the given fqn.
func (p *Project) NeedsUseStmtFor(pos *position.Position, FQN string) bool {
	root, err := p.wrksp.IROf(pos.Path)
	if err != nil {
		log.Println(
			fmt.Errorf(
				"Checking use statement needed, could not get nodes of %s: %w",
				pos.Path,
				err,
			),
		)
	}

	parts := strings.Split(FQN, `\`)
	className := parts[len(parts)-1]

	// Get how it would be resolved in the current file state.
	actFQN := definition.FullyQualify(root, className)

	// If the resolvement in current state equals the wanted fqn, no use stmt is needed.
	return actFQN.String() != FQN
}

// func (p *Project) method(
// 	root *ir.Root,
// 	classLikeScope ir.Node,
// 	scope ir.Node,
// 	method *ir.MethodCallExpr,
// ) (*ir.ClassMethodStmt, string, error) {
// 	var targetClass *traversers.TrieNode
// 	var targetPrivacy phprivacy.Privacy
// 	switch typedVar := method.Variable.(type) {
// 	case *ir.PropertyFetchExpr:
// 		prop, path, err := p.property(root, classLikeScope, scope, typedVar)
// 		if err != nil {
// 			return nil, "", fmt.Errorf("Method definition: unable to get type of variable that method is called on: %w", err)
// 		}
//
// 		varRoot, err := p.wrksp.IROf(path)
// 		if err != nil {
// 			return nil, "", fmt.Errorf("Method definition: unable to get file content/nodes of %s: %w", path, err)
// 		}
//
// 		propType := p.typer.Property(varRoot, prop)
// 		clsType, ok := propType.(*phpdoxer.TypeClassLike)
// 		if !ok {
// 			return nil, "", fmt.Errorf("Method definition: type of variable that method is called on is %s, expecting a class like type", propType)
// 		}
//
// 		target := p.findClassLikeSymbol(clsType)
// 		if target == nil {
// 			return nil, "", fmt.Errorf("Method definition: could not find definition for %s in symbol trie", clsType.Name)
// 		}
//
// 		targetClass = target
// 		targetPrivacy = phprivacy.PrivacyPublic
//
// 	case *ir.SimpleVar:
// 		classScope, privacy, err := p.variableType(root, classLikeScope, scope, typedVar)
// 		if err != nil {
// 			return nil, "", fmt.Errorf("Method definition: could not find definition for variable that method is called on: %w", err)
// 		}
//
// 		if classScope == nil {
// 			return nil, "", fmt.Errorf("Method definition: could not find class like definition for variable that method is called on")
// 		}
//
// 		targetClass = classScope
// 		targetPrivacy = privacy
//
// 	default:
// 		return nil, "", fmt.Errorf(
// 			"Method definition: unexpected variable node of type %T",
// 			method.Variable,
// 		)
// 	}
//
// 	methodName, ok := method.Method.(*ir.Identifier)
// 	if !ok {
// 		return nil, "", fmt.Errorf(
// 			"Method definition: unexpected variable node of type %T",
// 			method.Method,
// 		)
// 	}
//
// 	classRoot, err := p.wrksp.IROf(targetClass.Path)
// 	if err != nil {
// 		return nil, "", fmt.Errorf(
// 			"Method definition: unable to get file content/nodes of %s: %w",
// 			targetClass.Path,
// 			err,
// 		)
// 	}
//
// 	var result *ir.ClassMethodStmt
// 	var resultPath string
// 	err = p.walkResolveQueue(
// 		classRoot,
// 		targetClass.Symbol,
// 		func(wc *walkContext) (bool, error) {
// 			traverser := traversers.NewMethod(
// 				methodName.Value,
// 				wc.QueueNode.FQN.Name(),
// 				targetPrivacy,
// 			)
// 			wc.IR.Walk(traverser)
//
// 			if traverser.Method != nil {
// 				result = traverser.Method
// 				resultPath = wc.TrieNode.Path
//
// 				return true, nil
// 			}
//
// 			return false, nil
// 		},
// 	)
// 	if err != nil {
// 		return nil, "", fmt.Errorf(
// 			"Error retrieving method definition for %s on class %s: %w",
// 			methodName.Value,
// 			targetClass.Symbol.Identifier(),
// 			err,
// 		)
// 	}
//
// 	return result, resultPath, nil
// }
