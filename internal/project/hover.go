package project

import (
	"fmt"
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/elephp/pkg/traversers"
	"github.com/laytan/elephp/pkg/typer"
	log "github.com/sirupsen/logrus"
)

func (p *Project) Hover(pos *position.Position) (string, error) {
	path := pos.Path

	file := p.GetFile(path)
	if file == nil {
		return "", fmt.Errorf("Error retrieving file content for %s", path)
	}

	ast := p.ParseFileCached(file)
	if ast == nil {
		return "", fmt.Errorf("Error parsing %s for definitions", path)
	}

	apos := position.LocToPos(file.content, pos.Row, pos.Col)
	nap := traversers.NewNodeAtPos(apos)
	ast.Walk(nap)

	cmntsStr := ""
	signature := ""

Nodes:
	// TODO: this shows hover for the method if we are anywhere in the method, this is not what we want.
	for i := len(nap.Nodes) - 1; i >= 0; i-- {
		log.Infof("%T\n", nap.Nodes[i])
		switch typedNode := nap.Nodes[i].(type) {
		case *ir.ClassStmt, *ir.InterfaceStmt, *ir.TraitStmt:
			// TODO: show all methods on the class (also properties?).
			cmntsStr = CleanedNodeComments(typedNode)
			signature = NodeSignature(typedNode)

			break Nodes

		case *ir.ClassMethodStmt:
			cmntsStr = CleanedNodeComments(nap.Nodes[i+1])
			signature = NodeSignature(typedNode)

			break Nodes

		case *ir.FunctionStmt:
			// TODO: docs aren't on this node when the node has attributes (check in_array for example).
			cmntsStr = CleanedNodeComments(typedNode)
			signature = NodeSignature(typedNode)

			break Nodes

		case *ir.Parameter:
			// TODO: get @param comment from method statement too, something like:
			// if mtd, ok := nap.Nodes[i-1].(*ir.ClassMethodStmt); ok {
			// 	param := typer.New().Param(nap.Nodes[0].(*ir.Root), mtd, typedNode)
			// 	if param != nil {
			// 		cmntsStr = "/**\n * @param $" + typedNode.Variable.Name + " " + param.String() + "\n */"
			// 	}
			// }

			signature = NodeSignature(typedNode)

			break Nodes

		case *ir.PropertyListStmt:
			cmntsStr = CleanedNodeComments(typedNode)
			signature = NodeSignature(typedNode)

			break Nodes

		}
	}

	signature = strings.TrimSpace(signature)

	out := "```php\n<?php\n"

	if cmntsStr != "" {
		out += cmntsStr + "\n"
	}

	if signature != "" {
		out += signature + "\n"
	}

	out += "```"

	return out, nil
}

func CleanedNodeComments(node ir.Node) string {
	cmnts := typer.NodeComments(node)
	return strings.Join(cmnts, "\n")
}

// TODO: support attributes.
// NOTE: this is basically turning the node into the source string again,
// NOTE: can we not just get the bounds of the signature and parse that out of the source string?
// NOTE: it would be more complicated though.
func NodeSignature(node ir.Node) string {
	if node == nil {
		return ""
	}

	switch typedNode := node.(type) {
	case *ir.ArrayExpr:
		return "[]"

	case *ir.Identifier:
		return typedNode.Value

	case *ir.Name:
		// TODO: fully qualify the classlike name.
		return typedNode.Value

	case *ir.ClassStmt:
		signature := "class " + typedNode.ClassName.Value

		if typedNode.Extends != nil {
			signature += " extends " + typedNode.Extends.ClassName.Value
		}

		if typedNode.Implements != nil {
			signature += " implements "
			for _, name := range typedNode.Implements.InterfaceNames {
				signature += NodeSignature(name) + ", "
			}

			// Remove the last ", "
			signature = signature[:len(signature)-2]
		}

		signature += " {}"

		return signature

	case *ir.InterfaceStmt:
		signature := "interface " + typedNode.InterfaceName.Value

		if typedNode.Extends != nil && len(typedNode.Extends.InterfaceNames) > 0 {
			signature += " extends "
			for _, name := range typedNode.Extends.InterfaceNames {
				signature += NodeSignature(name) + ", "
			}

			signature = signature[:len(signature)-2]
		}

		signature += " {}"

		return signature

	case *ir.TraitStmt:
		return "trait " + typedNode.TraitName.Value + " {}"

	case *ir.ClassMethodStmt:
		signature := ""
		for _, modifier := range typedNode.Modifiers {
			signature += modifier.Value + " "
		}

		signature += "function " + typedNode.MethodName.Value + "("

		for _, param := range typedNode.Params {
			signature += NodeSignature(param)
			signature += ", "
		}

		if len(typedNode.Params) != 0 {
			signature = signature[:len(signature)-2]
		}

		signature += ")"

		if typedNode.ReturnType != nil {
			if typedNode.ReturnsRef {
				signature += "&"
			}

			signature += ": " + NodeSignature(typedNode.ReturnType)
		}

		signature += " {}"
		return signature

	case *ir.FunctionStmt:
		signature := ""

		signature += "function " + typedNode.FunctionName.Value + "("

		for _, param := range typedNode.Params {
			signature += NodeSignature(param)
			signature += ", "
		}

		if len(typedNode.Params) != 0 {
			signature = signature[:len(signature)-2]
		}

		signature += ")"

		if typedNode.ReturnType != nil {
			if typedNode.ReturnsRef {
				signature += "&"
			}

			signature += ": " + NodeSignature(typedNode.ReturnType)
		}

		signature += " {}"
		return signature

	case *ir.Parameter:
		signature := ""

		if typedNode.VariableType != nil {
			signature += NodeSignature(typedNode.VariableType) + " "
		}

		if typedNode.ByRef {
			signature += "&"
		}

		signature += "$" + typedNode.Variable.Name

		if typedNode.DefaultValue != nil {
			signature += " = " + NodeSignature(typedNode.DefaultValue)
		}

		return signature

	case *ir.PropertyListStmt:
		signature := ""
		for _, modifier := range typedNode.Modifiers {
			signature += modifier.Value + " "
		}

		if typedNode.Type != nil {
			signature += NodeSignature(typedNode.Type) + " "
		}

		for _, prop := range typedNode.Properties {
			signature += NodeSignature(prop)
		}

		return signature

	case *ir.PropertyStmt:
		return NodeSignature(typedNode.Variable) + " " + NodeSignature(typedNode.Expr)

	case *ir.SimpleVar:
		return typedNode.Name

	case *ir.Nullable:
		return "?" + NodeSignature(typedNode.Expr)

	case *ir.ConstFetchExpr:
		return NodeSignature(typedNode.Constant)

	default:
		log.Warnf("Node to string called with node of type %T which has no case", typedNode)
		return ""
	}
}
