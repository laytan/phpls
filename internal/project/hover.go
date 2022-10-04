package project

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irfmt"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/elephp/pkg/traversers"
	"github.com/laytan/elephp/pkg/typer"
)

func (p *Project) Hover(currpos *position.Position) string {
	poss, err := p.Definition(currpos)
	if err != nil || len(poss) == 0 {
		log.Println(fmt.Errorf("Hover error getting definition: %w", err))
		return ""
	}

	pos := poss[0]

	content, root, err := wrkspc.FromContainer().AllOf(pos.Path)
	if err != nil {
		log.Println(
			fmt.Errorf("Hover error getting content/parsing of %s: %w", pos.Path, err),
		)

		return ""
	}

	apos := position.LocToPos(content, pos.Row, pos.Col)
	nap := traversers.NewNodeAtPos(apos)
	root.Walk(nap)

	out := []string{}

Nodes:
	for i := len(nap.Nodes) - 1; i >= 0; i-- {
		switch typedNode := nap.Nodes[i].(type) {
		case *ir.ClassStmt, *ir.InterfaceStmt, *ir.TraitStmt, *ir.PropertyListStmt, *ir.FunctionStmt, *ir.FunctionCallExpr:
			// TODO: show all methods on the class (also properties?).
			if cmnts := cleanedNodeComments(typedNode); len(cmnts) > 0 {
				out = append(out, cmnts)
			}

			if signature := NodeSignature(typedNode); len(signature) > 0 {
				out = append(out, signature)
			}

			break Nodes

		case *ir.ClassMethodStmt:
			if cmnts := cleanedNodeComments(nap.Nodes[i+1]); len(cmnts) > 0 {
				out = append(out, cmnts)
			}

			if signature := NodeSignature(typedNode); len(signature) > 0 {
				out = append(out, signature)
			}

			break Nodes

		case *ir.Parameter:
			if signature := NodeSignature(typedNode); len(signature) > 0 {
				out = append(out, signature)
			}

			break Nodes

		case *ir.ClassConstListStmt:
			if cmnts := cleanedNodeComments(typedNode); len(cmnts) > 0 {
				out = append(out, cmnts)
			}

			if signature := NodeSignature(typedNode); len(signature) > 0 {
				out = append(out, signature)
			}

			break Nodes
		}
	}

	if len(out) == 0 {
		return ""
	}

	return wrapWithPhpMarkdown(strings.Join(out, "\n"))
}

func NodeSignature(node ir.Node) string {
	if node == nil {
		return ""
	}

	out := new(bytes.Buffer)
	p := irfmt.NewPrettyPrinter(out, "")
	p.Print(withoutStmts(node))
	return cleanBrackets(out.String())
}

func wrapWithPhpMarkdown(content string) string {
	return fmt.Sprintf("```php\n<?php\n%s\n```", content)
}

func cleanBrackets(signature string) string {
	signature = strings.TrimSpace(signature)
	if strings.HasSuffix(signature, "{\n\n}") {
		signature = strings.TrimSuffix(signature, "{\n\n}")
		signature = strings.TrimSpace(signature)
		signature += " {}"
	}

	return signature
}

func cleanedNodeComments(node ir.Node) string {
	cmnts := typer.NodeComments(node)
	return strings.Join(cmnts, "\n")
}

// TODO: can this be done in a less verbose way?
// Shallow copies the node excluding its Stmts property.
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
