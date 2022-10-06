package visitor

import (
	"bytes"
	"fmt"

	"appliedgo.net/what"
	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/token"
	"github.com/VKCOM/php-parser/pkg/visitor"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/phpversion"
)

// AtSinceAtRemoved removes nodes that are @since > or @removed <= the given version.
type AtSinceAtRemoved struct {
	visitor.Null

	version *phpversion.PHPVersion
	logging bool
}

func NewAtSinceAtRemoved(version *phpversion.PHPVersion, logging bool) *AtSinceAtRemoved {
	return &AtSinceAtRemoved{
		version: version,
		logging: logging,
	}
}

func (r *AtSinceAtRemoved) Root(n *ast.Root) {
	n.Stmts = r.filterStmts(n.Stmts)
}

func (r *AtSinceAtRemoved) StmtNamespace(n *ast.StmtNamespace) {
	n.Stmts = r.filterStmts(n.Stmts)
}

func (r *AtSinceAtRemoved) StmtClass(n *ast.StmtClass) {
	n.Stmts = r.filterStmts(n.Stmts)
}

func (r *AtSinceAtRemoved) StmtInterface(n *ast.StmtInterface) {
	n.Stmts = r.filterStmts(n.Stmts)
}

func (r *AtSinceAtRemoved) StmtTrait(n *ast.StmtTrait) {
	n.Stmts = r.filterStmts(n.Stmts)
}

func (r *AtSinceAtRemoved) filterStmts(nodes []ast.Vertex) []ast.Vertex {
	newStmts := make([]ast.Vertex, 0, len(nodes))
	for _, stmt := range nodes {
		ok := true
		switch typedStmt := stmt.(type) {
		case *ast.StmtNamespace:
			stmt.Accept(r)

		case *ast.StmtExpression:
			ok = !r.shouldRemoveExpression(typedStmt)

		case *ast.StmtFunction:
			ok = !r.shouldRemoveFunction(typedStmt)

		case *ast.StmtClass:
			stmt.Accept(r)
			ok = !r.shouldRemoveClass(typedStmt)

		case *ast.StmtInterface:
			stmt.Accept(r)
			ok = !r.shouldRemoveInterface(typedStmt)

		case *ast.StmtTrait:
			stmt.Accept(r)
			ok = !r.shouldRemoveTrait(typedStmt)

		case *ast.StmtClassMethod:
			ok = !r.shouldRemoveMethod(typedStmt)

		case *ast.StmtPropertyList:
			ok = !r.shouldRemovePropertyList(typedStmt)

		case *ast.StmtConstList:
			ok = !r.shouldRemoveConstList(typedStmt)

		case *ast.StmtClassConstList:
			ok = !r.shouldRemoveClassConstList(typedStmt)
		}

		if ok {
			newStmts = append(newStmts, stmt)
			continue
		}

		r.logRemoval(stmt)
	}

	return newStmts
}

func (r *AtSinceAtRemoved) shouldRemoveFunction(n *ast.StmtFunction) bool {
	return r.shouldRemove(n.FunctionTkn.FreeFloating) ||
		r.shouldRemove(r.attrGroupFreefloatings(n.AttrGroups))
}

func (r *AtSinceAtRemoved) shouldRemoveClass(n *ast.StmtClass) bool {
	return r.shouldRemove(n.ClassTkn.FreeFloating) ||
		r.shouldRemove(r.attrGroupFreefloatings(n.AttrGroups)) ||
		r.shouldRemove(r.identifiersFreefloatings(n.Modifiers))
}

func (r *AtSinceAtRemoved) shouldRemoveInterface(n *ast.StmtInterface) bool {
	return r.shouldRemove(n.InterfaceTkn.FreeFloating) ||
		r.shouldRemove(r.attrGroupFreefloatings(n.AttrGroups))
}

func (r *AtSinceAtRemoved) shouldRemoveTrait(n *ast.StmtTrait) bool {
	return r.shouldRemove(n.TraitTkn.FreeFloating) ||
		r.shouldRemove(r.attrGroupFreefloatings(n.AttrGroups))
}

func (r *AtSinceAtRemoved) shouldRemoveMethod(n *ast.StmtClassMethod) bool {
	return r.shouldRemove(n.FunctionTkn.FreeFloating) ||
		r.shouldRemove(r.attrGroupFreefloatings(n.AttrGroups)) ||
		r.shouldRemove(r.identifiersFreefloatings(n.Modifiers))
}

func (r *AtSinceAtRemoved) shouldRemovePropertyList(n *ast.StmtPropertyList) bool {
	return r.shouldRemove(r.attrGroupFreefloatings(n.AttrGroups)) ||
		r.shouldRemove(r.identifiersFreefloatings(n.Modifiers))
}

func (r *AtSinceAtRemoved) shouldRemoveConstList(n *ast.StmtConstList) bool {
	return r.shouldRemove(n.ConstTkn.FreeFloating)
}

func (r *AtSinceAtRemoved) shouldRemoveClassConstList(n *ast.StmtClassConstList) bool {
	return r.shouldRemove(r.identifiersFreefloatings(n.Modifiers)) ||
		r.shouldRemove(n.ConstTkn.FreeFloating)
}

func (r *AtSinceAtRemoved) shouldRemoveExpression(n *ast.StmtExpression) bool {
	switch typedNode := n.Expr.(type) {
	case *ast.ExprFunctionCall:
		return r.shouldRemoveFunctionCall(typedNode)

	case *ast.ExprAssign:
		return r.shouldRemoveAssignment(typedNode)

	default:
		return false
	}
}

// Handles `define()` function calls (constants).
func (r *AtSinceAtRemoved) shouldRemoveFunctionCall(fnCall *ast.ExprFunctionCall) bool {
	if fnName, ok := fnCall.Function.(*ast.Name); ok {
		if len(fnName.Parts) != 1 {
			return false
		}

		if fnNameStr, ok := fnName.Parts[0].(*ast.NamePart); ok {
			if !bytes.Equal(fnNameStr.Value, []byte("define")) {
				return false
			}

			return r.shouldRemove(fnNameStr.StringTkn.FreeFloating)
		}
	}

	return false
}

// Basic handling of superglobals.
func (r *AtSinceAtRemoved) shouldRemoveAssignment(assign *ast.ExprAssign) bool {
	if theVar, ok := assign.Var.(*ast.ExprVariable); ok {
		if theVarName, ok := theVar.Name.(*ast.Identifier); ok {
			return r.shouldRemove(theVarName.IdentifierTkn.FreeFloating)
		}
	}

	return false
}

func (r *AtSinceAtRemoved) shouldRemove(freefloatings []*token.Token) bool {
	for _, t := range freefloatings {
		if t.ID == token.T_DOC_COMMENT || t.ID == token.T_COMMENT {
			nodes, err := phpdoxer.ParseDoc(string(t.Value))
			if err != nil {
				what.Happens(fmt.Errorf("phpdoxer.ParseDoc(%s): %w", string(t.Value), err).Error())
				continue
			}

			for _, docNode := range nodes {
				switch typedDocNode := docNode.(type) {
				case *phpdoxer.NodeSince:
					if typedDocNode.Version.IsHigherThan(r.version) {
						return true
					}

				case *phpdoxer.NodeRemoved:
					if r.version.Equals(typedDocNode.Version) || r.version.IsHigherThan(typedDocNode.Version) {
						return true
					}

				default:
					continue
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
	if r.logging {
		_, _ = fmt.Printf("x") //nolint:forbidigo // For visualization.
	}
}
