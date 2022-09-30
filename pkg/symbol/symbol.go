package symbol

import (
	"log"

	"appliedgo.net/what"
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

func New(node ir.Node) Symbol {
	switch typedNode := node.(type) {
	case *ir.FunctionStmt:
		return NewFunction(typedNode)
	case *ir.ClassMethodStmt:
		return NewMethod(typedNode)
	case *ir.ClassStmt:
		return NewClassLikeClass(typedNode)
	case *ir.TraitStmt:
		return NewClassLikeTrait(typedNode)
	case *ir.InterfaceStmt:
		return NewClassLikeInterface(typedNode)
	case *ir.SimpleVar:
		return NewAssignment(typedNode)
	case *ir.PropertyListStmt:
		return NewProperty(typedNode)
	case *ir.FunctionCallExpr:
		if fn, ok := typedNode.Function.(*ir.Name); ok && fn.Value == "define" {
			return NewGlobalConstant(typedNode)
		}

		log.Printf("symbol.New called with unsupported type %T", node)
		return nil

	default:
		log.Printf("symbol.New called with unsupported type %T", node)
		return nil
	}
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
}

func NewFunction(stmt *ir.FunctionStmt) *FunctionStmtSymbol {
	f := &FunctionStmtSymbol{}
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

// AssignmentSymbol represents a variable assignment.
type AssignmentSymbol struct {
	baseSymbol
}

func NewAssignment(stmt *ir.SimpleVar) *AssignmentSymbol {
	a := &AssignmentSymbol{}
	a.FromNode(stmt)
	return a
}

func (a *AssignmentSymbol) NodeKind() ir.NodeKind {
	return ir.KindSimpleVar
}

type PropertySymbol struct {
	baseSymbol
}

func NewProperty(stmt *ir.PropertyListStmt) *PropertySymbol {
	a := &PropertySymbol{}
	a.FromNode(stmt)
	return a
}

func (a *PropertySymbol) NodeKind() ir.NodeKind {
	return ir.KindPropertyListStmt
}

type MethodSymbol struct {
	baseSymbol
}

func NewMethod(stmt *ir.ClassMethodStmt) *MethodSymbol {
	m := &MethodSymbol{}
	m.FromNode(stmt)
	return m
}

func (a *MethodSymbol) NodeKind() ir.NodeKind {
	return ir.KindClassMethodStmt
}

type GlobalConstantSymbol struct {
	baseSymbol
}

func NewGlobalConstant(defineCall *ir.FunctionCallExpr) *GlobalConstantSymbol {
	c := &GlobalConstantSymbol{}
	c.FromNode(defineCall)

	if firstArg, ok := defineCall.Args[0].(*ir.Argument); ok {
		if ident, ok := firstArg.Expr.(*ir.String); ok {
			c.identifier = ident.Value
			return c
		}
	}

	log.Println(
		"Invalid constant declaration, first argument of define() call is not a string (constant identifier)",
	)
	what.Is(defineCall)
	c.identifier = "ELEPHP_INVALID_CONST"

	return c
}

func (g *GlobalConstantSymbol) NodeKind() ir.NodeKind {
	return ir.KindConstantStmt
}
