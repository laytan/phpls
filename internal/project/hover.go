package project

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irfmt"
	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/elephp/pkg/traversers"
	"github.com/laytan/elephp/pkg/typer"
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
		// log.Infof("%T\n", nap.Nodes[i])
		switch typedNode := nap.Nodes[i].(type) {
		case *ir.ClassStmt, *ir.InterfaceStmt, *ir.TraitStmt, *ir.PropertyListStmt, *ir.FunctionStmt:
			// TODO: show all methods on the class (also properties?).
			cmntsStr = CleanedNodeComments(typedNode)
			signature = NodeSignature(typedNode)

			break Nodes

		case *ir.ClassMethodStmt:
			cmntsStr = CleanedNodeComments(nap.Nodes[i+1])
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

func NodeSignature(node ir.Node) string {
	if node == nil {
		return ""
	}

	out := new(bytes.Buffer)
	p := irfmt.NewPrettyPrinter(out, "")
	p.Print(withoutStmts(node))
	return out.String()
}

// TODO: can this be done in a less verbose way?
func withoutStmts(node ir.Node) ir.Node {
	switch t := node.(type) {
	case *ir.ClassStmt:
		return &ir.ClassStmt{
			Position:             t.Position,
			AttrGroups:           t.AttrGroups,
			Modifiers:            t.Modifiers,
			ClassTkn:             t.ClassTkn,
			ClassName:            t.ClassName,
			OpenCurlyBracketTkn:  t.OpenCurlyBracketTkn,
			CloseCurlyBracketTkn: t.CloseCurlyBracketTkn,
			Class: ir.Class{
				Extends:    t.Extends,
				Implements: t.Implements,
				Doc:        t.Doc,
				Stmts:      nil,
			},
		}

	case *ir.InterfaceStmt:
		return &ir.InterfaceStmt{
			Position:             t.Position,
			AttrGroups:           t.AttrGroups,
			InterfaceTkn:         t.InterfaceTkn,
			InterfaceName:        t.InterfaceName,
			Extends:              t.Extends,
			OpenCurlyBracketTkn:  t.OpenCurlyBracketTkn,
			CloseCurlyBracketTkn: t.CloseCurlyBracketTkn,
			Doc:                  t.Doc,
			Stmts:                nil,
		}

	case *ir.TraitStmt:
		return &ir.TraitStmt{
			Position:             t.Position,
			AttrGroups:           t.AttrGroups,
			TraitTkn:             t.TraitTkn,
			TraitName:            t.TraitName,
			OpenCurlyBracketTkn:  t.OpenCurlyBracketTkn,
			CloseCurlyBracketTkn: t.CloseCurlyBracketTkn,
			Doc:                  t.Doc,
			Stmts:                nil,
		}

	case *ir.FunctionStmt:
		return &ir.FunctionStmt{
			Position:             t.Position,
			AttrGroups:           t.AttrGroups,
			FunctionTkn:          t.FunctionTkn,
			AmpersandTkn:         t.AmpersandTkn,
			FunctionName:         t.FunctionName,
			OpenParenthesisTkn:   t.OpenParenthesisTkn,
			Params:               t.Params,
			SeparatorTkns:        t.SeparatorTkns,
			CloseParenthesisTkn:  t.CloseParenthesisTkn,
			ColonTkn:             t.ColonTkn,
			ReturnType:           t.ReturnType,
			OpenCurlyBracketTkn:  t.OpenCurlyBracketTkn,
			CloseCurlyBracketTkn: t.CloseCurlyBracketTkn,
			ReturnsRef:           t.ReturnsRef,
			Doc:                  t.Doc,
			Stmts:                nil,
		}

	case *ir.ClassMethodStmt:
		return &ir.ClassMethodStmt{
			Position:            t.Position,
			AttrGroups:          t.AttrGroups,
			Modifiers:           t.Modifiers,
			FunctionTkn:         t.FunctionTkn,
			AmpersandTkn:        t.AmpersandTkn,
			MethodName:          t.MethodName,
			OpenParenthesisTkn:  t.OpenParenthesisTkn,
			Params:              t.Params,
			SeparatorTkns:       t.SeparatorTkns,
			CloseParenthesisTkn: t.CloseParenthesisTkn,
			ColonTkn:            t.ColonTkn,
			ReturnType:          t.ReturnType,
			ReturnsRef:          t.ReturnsRef,
			Doc:                 t.Doc,
			Stmt:                nil,
		}

	default:
		return node

	}
}
