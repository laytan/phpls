// nolint: gocritic // hugeParam for dump not prio to address.
package ast

import "io"

type CallExprOrStmt interface {
	Node
	callExprOrStmt()
}

// type Call struct {
// 	BaseNode
// 	Name       string `parser:"@Ident"`
// 	Parameters []Expr `parser:"LParen @@? ( Comma @@ )* RParen"`
// }
//
// func (c Call) expr()           {}
// func (c Call) exprNoArr()      {}
// func (c Call) callExprOrStmt() {}
// func (c Call) exprNoConcat()   {}
// func (c Call) exprNoEquals()   {}
// func (c Call) assignable()     {}

type CallStatement struct {
	BaseNode

	Call Call `parser:"@@ Semicolon"`
}

var _ Node = CallStatement{}

// TODO: this is alot.
func (c CallStatement) statement()         {}
func (c CallStatement) classStatement()    {}
func (c CallStatement) callableStatement() {}
func (c CallStatement) callExprOrStmt()    {}

func (c CallStatement) Dump(w io.Writer, level int) {
	c.BaseDump(w, level, "CallStatement")
	level++
	writeStr(w, "Call:")
	level++
	writeEOL(w, level)
	c.Call.Dump(w, level)
}

type Die struct {
	BaseNode

	AsCall     bool   `parser:"Die @LParen?"`
	Parameters []Expr `parser:"@@? ( Comma @@ )* RParen? Semicolon"`
}

var _ Node = Die{}

func (d Die) statement()         {}
func (d Die) classStatement()    {}
func (d Die) callableStatement() {}
func (d Die) callExprOrStmt()    {}

func (d Die) Dump(w io.Writer, level int) {
	d.BaseDump(w, level, "Die")
	level++
	writeStr(w, "AsCall: %v", d.AsCall)
	writeEOL(w, level)
	writeStr(w, "Parameters:")
	level++
	writeEOL(w, level)
	writeList(w, level, d.Parameters)
}

type Exit struct {
	BaseNode

	AsCall     bool   `parser:"Exit @LParen?"`
	Parameters []Expr `parser:"@@? ( Comma @@ )* RParen? Semicolon"`
}

func (e Exit) statement()         {}
func (e Exit) classStatement()    {}
func (e Exit) callableStatement() {}
func (e Exit) callExprOrStmt()    {}

var _ Node = Exit{}

func (e Exit) Dump(w io.Writer, level int) {
	e.BaseDump(w, level, "Exit")
	level++
	writeStr(w, "AsCall: %v", e.AsCall)
	writeEOL(w, level)
	writeStr(w, "Parameters:")
	level++
	writeEOL(w, level)
	writeList(w, level, e.Parameters)
}
