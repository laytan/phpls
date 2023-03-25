// nolint: gocritic // hugeParam for dump not prio to address.
package ast

import (
	"io"
)

var (
	StatementImpls = []Statement{
		Die{},
		Exit{},
		CallStatement{},
		Class{},
		Function{},
		Die{},
		Echo{},
		Global{},
		Require{},
		Namespace{},
		Foreach{},
		IfStatement{},
		ExprStatement{},
	}
	ClassStatementImpls = []ClassStatement{
		Method{},
		CallStatement{},
		Require{},
	}
	CallableStatementImpls = []CallableStatement{
		Die{},
		Exit{},
		CallStatement{},
		Echo{},
		Global{},
		Require{},
		Foreach{},
		IfStatement{},
		ExprStatement{},
	}
)

type (
	// Any statement that can be at the top level, see `ast.StatementImpls`.
	Statement interface {
		Node
		statement()
	}

	// Any statement that can be in a class body, see `ast.ClassStatementImpls`.
	ClassStatement interface {
		Node
		classStatement()
	}

	// Any statement that can be in a function/method body, see `ast.CallableStatementImpls`.
	CallableStatement interface {
		Node
		callableStatement()
	}
)

type Namespace struct {
	BaseNode

	Name string `parser:"Namespace @Ident Semicolon"`
}

func (n Namespace) statement() {}

func (n Namespace) Dump(w io.Writer, level int) {
	n.BaseDump(w, level, "Namespace")
	level++
	writeStr(w, "Name: %s", n.Name)
	writeEOL(w, level)
}

type Require struct {
	BaseNode

	IsOnce bool `parser:"( @RequireOnce?"`
	Value  Expr `parser:"Require? )! @@ Semicolon"`
}

func (r Require) statement()         {}
func (r Require) classStatement()    {}
func (r Require) callableStatement() {}

func (r Require) Dump(w io.Writer, level int) {
	r.BaseDump(w, level, "Require")
	level++
	writeStr(w, "IsOnce: %v", r.IsOnce)
	writeEOL(w, level)
	writeStr(w, "Value:")
	level++
	writeEOL(w, level)
	r.Value.Dump(w, level)
}

type Global struct {
	BaseNode

	Vars []Var `parser:"Global @@ (Comma @@)* Semicolon"`
}

func (g Global) statement()         {}
func (g Global) callableStatement() {}

func (g Global) Dump(w io.Writer, level int) {
	g.BaseDump(w, level, "Global")
	level++
	writeStr(w, "Vars:")
	level++
	writeEOL(w, level)
	writeList(w, level, g.Vars)
}

type Echo struct {
	BaseNode

	Exprs []Expr `parser:"Echo @@ (Comma @@)* Semicolon"`
}

func (e Echo) statement()         {}
func (e Echo) callableStatement() {}

func (e Echo) Dump(w io.Writer, level int) {
	e.BaseDump(w, level, "Echo")
	level++
	writeStr(w, "Exprs:")
	level++
	writeEOL(w, level)
	writeList(w, level, e.Exprs)
}

type ExprStatement struct {
	BaseNode

	Comment *string `parser:"@BlockComment?"`
	Expr    Expr    `parser:"@@ Semicolon"`
}

func (a ExprStatement) statement()         {}
func (a ExprStatement) callableStatement() {}

func (a ExprStatement) Dump(w io.Writer, level int) {
	a.BaseDump(w, level, "ExprStatement")
	level++
	writeStr(w, "Comment: %s", defaultStr(a.Comment))
	writeEOL(w, level)
	writeStr(w, "Expr:")
	level++
	writeEOL(w, level)
	a.Expr.Dump(w, level)
}

type Class struct {
	BaseNode

	Comment    *string          `parser:"@BlockComment?"`
	Modifiers  Modifiers        `parser:"@@ Class"`
	Name       Name             `parser:"@@"`
	Extends    *Extends         `parser:"@@?"`
	Implements *Implements      `parser:"@@?"`
	Statements []ClassStatement `parser:"LBrace @@* RBrace"`
}

func (c Class) statement() {}

func (c Class) Dump(w io.Writer, level int) {
	c.BaseDump(w, level, "Class")
	level++
	writeStr(w, "Comment: %s", defaultStr(c.Comment))
	writeEOL(w, level)
	writeStr(w, "Modifiers:")
	level++
	writeEOL(w, level)
	c.Modifiers.Dump(w, level)
	level--
	writeStr(w, "Name:")
	level++
	writeEOL(w, level)
	c.Name.Dump(w, level)
	level--
	writeStr(w, "Extends:")
	if c.Extends != nil {
		level++
		writeEOL(w, level)
		c.Extends.Dump(w, level)
		level--
	}
	writeStr(w, "Implements:")
	if c.Implements != nil {
		level++
		writeEOL(w, level)
		c.Implements.Dump(w, level)
		level--
	}
	writeStr(w, "Statements:")
	level++
	writeList(w, level, c.Statements)
}

type Method struct {
	BaseNode

	Comment    *string             `parser:"@BlockComment?"`
	Modifiers  Modifiers           `parser:"@@"`
	Name       string              `parser:"Function @Ident"`
	Params     []Parameter         `parser:"LParen @@? ( Comma @@ )* RParen"`
	ReturnType *string             `parser:"(Colon @Ident)?"`
	Statements []CallableStatement `parser:"LBrace @@* RBrace"`
}

func (m Method) classStatement() {}

func (m Method) Dump(w io.Writer, level int) {
	m.BaseDump(w, level, "Method")
	level++
	writeStr(w, "Comment: %s", defaultStr(m.Comment))
	writeStr(w, "Modifiers:")
	level++
	writeEOL(w, level)
	m.Modifiers.Dump(w, level)
	level--
	writeStr(w, "Name: %s", m.Name)
	writeEOL(w, level)
	writeStr(w, "Parameters:")
	level++
	writeEOL(w, level)
	writeList(w, level, m.Params)
	level--
	writeEOL(w, level)
	writeStr(w, "ReturnType: %s", defaultStr(m.ReturnType))
	writeEOL(w, level)
	writeStr(w, "Statements:")
	level++
	writeList(w, level, m.Statements)
}

type Function struct {
	BaseNode

	Comment    *string             `parser:"@BlockComment?"`
	Name       string              `parser:"Function @Ident"`
	Params     []Parameter         `parser:"LParen @@? ( Comma @@ )* RParen"`
	ReturnType *string             `parser:"(Colon @Ident)?"`
	Statements []CallableStatement `parser:"LBrace @@* RBrace"`
}

func (f Function) statement() {}

func (f Function) Dump(w io.Writer, level int) {
	f.BaseDump(w, level, "Function")
	level++
	writeStr(w, "Comment: %s", defaultStr(f.Comment))
	writeStr(w, "Name: %s", f.Name)
	writeEOL(w, level)
	writeStr(w, "Parameters:")
	level++
	writeEOL(w, level)
	writeList(w, level, f.Params)
	level--
	writeEOL(w, level)
	writeStr(w, "ReturnType: %s", defaultStr(f.ReturnType))
	writeEOL(w, level)
	writeStr(w, "Statements:")
	level++
	writeList(w, level, f.Statements)
}

type Parameter struct {
	BaseNode

	Variadic     bool          `parser:"@Variadic?"` // TODO: only the last parameter can be variadic.
	Reference    bool          `parser:"@Reference?"`
	TypeHint     *string       `parser:"@Ident?"`
	Var          string        `parser:"@Var"`
	DefaultValue *DefaultValue `parser:"@@?"`
}

func (p Parameter) Dump(w io.Writer, level int) {
	p.BaseDump(w, level, "Parameter")
	level++
	writeStr(w, "Variadic: %v", p.Variadic)
	writeEOL(w, level)
	writeStr(w, "Reference: %v", p.Reference)
	writeEOL(w, level)
	writeStr(w, "TypeHint: %s", defaultStr(p.TypeHint))
	writeEOL(w, level)
	writeStr(w, "Var: %s", p.Var)
	writeEOL(w, level)
	writeStr(w, "DefaultValue:")
	if p.DefaultValue != nil {
		level++
		writeEOL(w, level)
		p.DefaultValue.Dump(w, level)
	}
}

type DefaultValue struct {
	BaseNode

	// TODO: default value can be scalar values, arrays, null, or a (new ClassName())
	Value Expr `parser:"Assign @@"`
}

func (d DefaultValue) Dump(w io.Writer, level int) {
	d.BaseDump(w, level, "DefaultValue")
	level++
	writeStr(w, "Value:")
	level++
	writeEOL(w, level)
	d.Value.Dump(w, level)
}

type IfStatement struct {
	BaseNode

	If      IfPart   `parser:"If @@"`
	ElseIfs []IfPart `parser:"(ElseIf @@)*"`
	Else    *Else    `parser:"@@?"`
}

func (i IfStatement) statement()         {}
func (i IfStatement) callableStatement() {}

func (i IfStatement) Dump(w io.Writer, level int) {
	i.BaseDump(w, level, "IfStatement")
	level++
	writeStr(w, "If:")
	level++
	writeEOL(w, level)
	i.If.Dump(w, level)
	level--
	writeEOL(w, level)
	writeStr(w, "ElseIfs:")
	level++
	writeEOL(w, level)
	writeList(w, level, i.ElseIfs)
	level--
	writeEOL(w, level)
	writeStr(w, "Else:")
	if i.Else != nil {
		level++
		writeEOL(w, level)
		i.Else.Dump(w, level)
	} else {
		writeEOL(w, level)
	}
}

type IfPart struct {
	BaseNode

	Condition  Expr        `parser:"LParen @@ RParen"`
	Statements []Statement `parser:"( LBrace @@* RBrace ) | ( Colon @@* EndIf Semicolon )"` // TODO: this could be statements, but also could be CallableStatments.
}

func (i IfPart) Dump(w io.Writer, level int) {
	i.BaseDump(w, level, "IfPart")
	level++
	writeStr(w, "Condition:")
	level++
	writeEOL(w, level)
	i.Condition.Dump(w, level)
	level--
	writeEOL(w, level)
	writeStr(w, "Statements:")
	level++
	writeEOL(w, level)
	writeList(w, level, i.Statements)
}

type Else struct {
	BaseNode

	Statements []Statement `parser:"Else ( LBrace @@* RBrace ) | ( Colon @@* (?= ElseIf Colon ) | (?= EndIf Semicolon ) )"`
}

func (e Else) Dump(w io.Writer, level int) {
	e.BaseDump(w, level, "Else")
	level++
	writeStr(w, "Statements:")
	level++
	writeEOL(w, level)
	writeList(w, level, e.Statements)
}

type Foreach struct {
	BaseNode

	Collection  Expr                `parser:"ForEach LParen @@ As"`
	Key         *Expr               `parser:"(@@ Arrow)?"`
	ByReference bool                `parser:"Reference?"`
	Value       Expr                `parser:"@@ RParen"`
	Statements  []CallableStatement `parser:"( LBrace @@* RBrace ) | ( Colon @@* EndForEach Semicolon )"`
}

func (f Foreach) statement()         {}
func (f Foreach) callableStatement() {}

func (f Foreach) Dump(w io.Writer, level int) {
	f.BaseDump(w, level, "Foreach")
	level++
	writeStr(w, "Collection:")
	level++
	writeEOL(w, level)
	f.Collection.Dump(w, level)
	level--
	if f.Key != nil {
		writeStr(w, "Key:")
		level++
		writeEOL(w, level)
		f.Key.Dump(w, level)
		level--
	}
	writeStr(w, "ByReference: %v", f.ByReference)
	writeEOL(w, level)
	writeStr(w, "Value:")
	level++
	writeEOL(w, level)
	f.Value.Dump(w, level)
	level--
	writeEOL(w, level)
	writeStr(w, "Statements:")
	level++
	writeEOL(w, level)
	writeList(w, level, f.Statements)
}
