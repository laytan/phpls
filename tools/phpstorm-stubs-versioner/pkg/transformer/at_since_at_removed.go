package transformer

import (
	"fmt"
	"log"

	"github.com/VKCOM/noverify/src/ir/irconv"
	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/token"
	"github.com/VKCOM/php-parser/pkg/visitor"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/pkg/symbol"
)

// AtSinceAtRemoved removes nodes that are @since > or @removed <= the given version.
type AtSinceAtRemoved struct {
	visitor.Null
	version *phpversion.PHPVersion
}

func (r *AtSinceAtRemoved) Transform(ast ast.Vertex, version string) ast.Vertex {
	// TODO: don't do this here.
	phpv, ok := phpversion.FromString(version)
	if !ok {
		panic("not ok version")
	}

	r.version = phpv
	ast.Accept(r)
	return ast
}

func (r *AtSinceAtRemoved) Root(n *ast.Root) {
	newStmts := make([]ast.Vertex, 0, len(n.Stmts))
	for _, stmt := range n.Stmts {
		var ok bool
		switch typedStmt := stmt.(type) {
		case *ast.StmtFunction:
			ok = !r.shouldRemoveFunction(typedStmt)

		case *ast.StmtClass:
			ok = !r.shouldRemoveClass(typedStmt)
			if ok {
				stmt.Accept(r)
			}

		default:
			ok = true
		}

		if ok {
			newStmts = append(newStmts, stmt)
			continue
		}

		r.logRemoval(stmt)
	}

	n.Stmts = newStmts
}

func (r *AtSinceAtRemoved) shouldRemoveFunction(n *ast.StmtFunction) bool {
	return r.shouldRemove(n.FunctionTkn.FreeFloating) ||
		r.shouldRemove(r.attrGroupFreefloatings(n.AttrGroups))
}

func (r *AtSinceAtRemoved) shouldRemoveClass(n *ast.StmtClass) bool {
	return r.shouldRemove(n.ClassTkn.FreeFloating) ||
		r.shouldRemove(r.attrGroupFreefloatings(n.AttrGroups))
}

func (r *AtSinceAtRemoved) StmtClass(n *ast.StmtClass) {
	newStmts := make([]ast.Vertex, 0, len(n.Stmts))
	for _, stmt := range n.Stmts {
		var ok bool
		switch typedStmt := stmt.(type) {
		case *ast.StmtClassMethod:
			ok = !r.shouldRemoveMethod(typedStmt)

		case *ast.StmtClassConstList:
			ok = !r.shouldRemoveConstList(typedStmt)

		default:
			ok = true
		}

		if ok {
			newStmts = append(newStmts, stmt)
			continue
		}

		r.logRemoval(stmt)
	}

	n.Stmts = newStmts
}

func (r *AtSinceAtRemoved) shouldRemoveMethod(n *ast.StmtClassMethod) bool {
	return r.shouldRemove(n.FunctionTkn.FreeFloating) ||
		r.shouldRemove(r.attrGroupFreefloatings(n.AttrGroups)) ||
		r.shouldRemove(r.identifiersFreefloatings(n.Modifiers))
}

func (r *AtSinceAtRemoved) shouldRemoveConstList(n *ast.StmtClassConstList) bool {
	return r.shouldRemove(r.identifiersFreefloatings(n.Modifiers))
}

func (r *AtSinceAtRemoved) shouldRemove(freefloatings []*token.Token) bool {
	for _, t := range freefloatings {
		if t.ID == token.T_DOC_COMMENT || t.ID == token.T_COMMENT {
			nodes, err := phpdoxer.ParseDoc(string(t.Value))
			if err != nil {
				log.Println(fmt.Errorf("phpdoxer.ParseDoc(%s): %w", string(t.Value), err))
			}

			for _, docNode := range nodes {
				if sinceNode, ok := docNode.(*phpdoxer.NodeSince); ok {
					if sinceNode.Version.IsHigherThan(r.version) {
						return true
					}
				}
			}
		}
	}

	return false
}

func (r *AtSinceAtRemoved) attrGroupFreefloatings(n []ast.Vertex) []*token.Token {
	freefloatings := []*token.Token{}
	for _, atr := range n {
		if atrGroup, ok := atr.(*ast.AttributeGroup); ok {
			freefloatings = append(freefloatings, atrGroup.OpenAttributeTkn.FreeFloating...)
		}
	}

	return freefloatings
}

func (r *AtSinceAtRemoved) identifiersFreefloatings(n []ast.Vertex) []*token.Token {
	freefloatings := []*token.Token{}
	for _, atr := range n {
		if ident, ok := atr.(*ast.Identifier); ok {
			freefloatings = append(freefloatings, ident.IdentifierTkn.FreeFloating...)
		}
	}

	return freefloatings
}

func (r *AtSinceAtRemoved) logRemoval(n ast.Vertex) {
	irNode := irconv.ConvertNode(n)
	ident := symbol.GetIdentifier(irNode)
	log.Printf("Removing %T %s because @since or @removed indicates it", n, ident)
}
