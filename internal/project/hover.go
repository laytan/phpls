package project

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/laytan/elephp/internal/symbol"
	"github.com/laytan/elephp/internal/throws"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/elephp/pkg/traversers"
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor/printer"
	"github.com/laytan/php-parser/pkg/visitor/traverser"
)

func (p *Project) Hover(currpos *position.Position) string {
	nodes, root := nodeToHover(p, currpos)
	if len(nodes) == 0 || root == nil {
		return ""
	}

	top := []string{}
	out := []string{}

Nodes:
	for i := len(nodes) - 1; i >= 0; i-- {
		switch typedNode := nodes[i].(type) {
		case *ast.StmtClass, *ast.StmtInterface, *ast.StmtTrait, *ast.StmtPropertyList, *ast.ExprFunctionCall:
			if cmnts := cleanedNodeComments(typedNode); len(cmnts) > 0 {
				out = append(out, cmnts)
			}

			if signature := NodeSignature(typedNode); len(signature) > 0 {
				out = append(out, signature)
			}

			break Nodes

		case *ast.StmtFunction:
			top = append(top, funcOrMethodThrows(root, typedNode, currpos.Path))

			if cmnts := cleanedNodeComments(typedNode); len(cmnts) > 0 {
				out = append(out, cmnts)
			}

			if signature := NodeSignature(typedNode); len(signature) > 0 {
				out = append(out, signature)
			}

			break Nodes

		case *ast.StmtClassMethod:
			top = append(top, funcOrMethodThrows(root, typedNode, currpos.Path))

			if cmnts := cleanedNodeComments(nodes[i+1]); len(cmnts) > 0 {
				out = append(out, cmnts)
			}

			if signature := NodeSignature(typedNode); len(signature) > 0 {
				out = append(out, signature)
			}

			break Nodes

		case *ast.Parameter:
			if signature := NodeSignature(typedNode); len(signature) > 0 {
				out = append(out, signature)
			}

			break Nodes

		case *ast.StmtClassConstList:
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

	outStr := wrapWithPhpMarkdown(strings.Join(out, "\n"))

	if len(top) == 0 {
		return outStr
	}

	return strings.Join(top, "\n") + "\n" + outStr
}

func nodeToHover(p *Project, currpos *position.Position) ([]ast.Vertex, *ast.Root) {
	napper := func(pos *position.Position) ([]ast.Vertex, *ast.Root) {
		content, root := wrkspc.FromContainer().FAllOf(pos.Path)
		apos := position.LocToPos(content, pos.Row, pos.Col)
		nap := traversers.NewNodeAtPos(int(apos))
		napt := traverser.NewTraverser(nap)
		root.Accept(napt)

		return nap.Nodes, root
	}

	poss, err := p.Definition(currpos)
	if err != nil || len(poss) == 0 {
		return napper(currpos)
	}

	pos := poss[0]
	return napper(pos)
}

func NodeSignature(node ast.Vertex) string {
	if node == nil {
		return ""
	}

	out := new(bytes.Buffer)
	p := printer.NewPrinter(out)
	withoutStmts(node).Accept(p)
	return cleanBrackets(out.String())
}

func funcOrMethodThrows(root *ast.Root, node ast.Vertex, path string) string {
	builder := strings.Builder{}
	//nolint:revive // This always returns a nil error.
	builder.WriteString("Throws:")

	thrown := throws.NewResolver(wrkspc.NewRooter(path, root), node).Throws()
	for _, t := range thrown {
		//nolint:revive // This always returns a nil error.
		builder.WriteString(fmt.Sprintf("\n- `%s`", t.String()))
	}

	//nolint:revive // This always returns a nil error.
	builder.WriteString("\n")

	return builder.String()
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

func cleanedNodeComments(node ast.Vertex) string {
	cmnts := symbol.NodeComments(node)
	return strings.Join(cmnts, "\n")
}

func withoutStmts(node ast.Vertex) ast.Vertex {
	switch t := node.(type) {
	case *ast.StmtClass:
		t.Stmts = nil
		return t
	case *ast.StmtInterface:
		t.Stmts = nil
		return t
	case *ast.StmtTrait:
		t.Stmts = nil
		return t
	case *ast.StmtFunction:
		t.Stmts = nil
		return t
	case *ast.StmtClassMethod:
		t.Stmt = nil
		return t
	}

	return node
}
