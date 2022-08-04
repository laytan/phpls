package symbol

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/php-parser/pkg/position"
)

// Symbol is basically an ir.Node with all non-relevant data stripped.
// This is needed to save a bunch of memory for child nodes etc. that won't ever
// be accessed.
type Symbol interface {
	Position() *position.Position
	NodeKind() ir.NodeKind
	Identifier() string
}

type baseSymbol struct {
	position   *position.Position
	identifier string
}

func (b *baseSymbol) FromNode(node ir.Node) {
	b.position = ir.GetPosition(node)
	b.identifier = GetIdentifier(node)
}

func (b *baseSymbol) Position() *position.Position {
	return b.position
}

func (b *baseSymbol) Identifier() string {
	return b.identifier
}

// FunctionStmtSymbol is a stripped version of ir.FunctionStmt.
type FunctionStmtSymbol struct {
	baseSymbol
	Params []ir.Node
}

func NewFunction(stmt *ir.FunctionStmt) *FunctionStmtSymbol {
	f := &FunctionStmtSymbol{Params: stmt.Params}
	f.FromNode(stmt)
	return f
}

func (f *FunctionStmtSymbol) NodeKind() ir.NodeKind {
	return ir.KindFunctionStmt
}

// ClassLikeStmtSymbol is a stripped version of ir.ClassStmt, ir.InterfaceStmt or ir.TraitStmt.
type ClassLikeStmtSymbol struct {
	baseSymbol
	kind ir.NodeKind
}

func NewClassLikeClass(stmt *ir.ClassStmt) *ClassLikeStmtSymbol {
	c := &ClassLikeStmtSymbol{kind: ir.KindClassStmt}
	c.FromNode(stmt)
	return c
}

func NewClassLikeInterface(stmt *ir.InterfaceStmt) *ClassLikeStmtSymbol {
	c := &ClassLikeStmtSymbol{kind: ir.KindInterfaceStmt}
	c.FromNode(stmt)
	return c
}

func NewClassLikeTrait(stmt *ir.TraitStmt) *ClassLikeStmtSymbol {
	c := &ClassLikeStmtSymbol{kind: ir.KindTraitStmt}
	c.FromNode(stmt)
	return c
}

func (c *ClassLikeStmtSymbol) NodeKind() ir.NodeKind {
	return c.kind
}