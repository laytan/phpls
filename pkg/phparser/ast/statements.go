// nolint: gocritic // hugeParam for dump not prio to address.
package ast

import (
	"io"

	"github.com/alecthomas/participle/v2/lexer"
)

var (
	StatementImpls = []Statement{
		Die{},
		Exit{},
		CallStatement{},
		Assign{},
		Class{},
		Function{},
		Die{},
		Require{},
		Namespace{},
		IfStatement{},
	}
	ClassStatementImpls = []ClassStatement{
		Method{},
		CallStatement{},
		Require{},
		IfStatement{},
	}
	CallableStatementImpls = []CallableStatement{
		Die{},
		Exit{},
		CallStatement{},
		Assign{},
		Require{},
		IfStatement{},
	}
	AssignableImpls = []Assignable{
		Var{},
		Call{},
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

	// Anything that can prefix an assigment, see `ast.AssignableImpls`.
	Assignable interface {
		Node
		assignable()
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

// type RequireOnce struct {
//     BaseNode
//
//     Value Expr `parser:"RequireOnce @@ Semicolon"`
// }
//
// func (r RequireOnce) statement()         {}
// func (r RequireOnce) classStatement()    {}
// func (r RequireOnce) callableStatement() {}
//
// func (r RequireOnce) Dump(w io.Writer, level int) {
// 	r.BaseDump(w, level, "Require")
// 	level++
// 	writeStr(w, "Value:")
// 	level++
// 	writeEOL(w, level)
// 	r.Value.Dump(w, level)
// }

type Assign struct {
	BaseNode

	Comment  *string     `parser:"@BlockComment?"`
	Var      Assignable  `parser:"@@"`
	Operator lexer.Token `parser:"@( Assign | ConcatAssign )"`
	Value    Expr        `parser:"@@ Semicolon"`
}

func (a Assign) statement()         {}
func (a Assign) callableStatement() {}

func (a Assign) Dump(w io.Writer, level int) {
	a.BaseDump(w, level, "Assign")
	level++
	writeStr(w, "Comment: %s", defaultStr(a.Comment))
	writeEOL(w, level)
	writeStr(w, "Var:")
	level++
	writeEOL(w, level)
	a.Var.Dump(w, level)
	level--
	writeEOL(w, level)
	writeStr(w, "Value:")
	level++
	writeEOL(w, level)
	a.Value.Dump(w, level)
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
func (i IfStatement) classStatement()    {}
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
	Statements []Statement `parser:"LBrace @@* RBrace"` // TODO: this could be statements, but also could be CallableStatments.
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

	Statements []Statement `parser:"Else LBrace @@* RBrace"`
}

func (e Else) Dump(w io.Writer, level int) {
	e.BaseDump(w, level, "Else")
	level++
	writeStr(w, "Statements:")
	level++
	writeEOL(w, level)
	writeList(w, level, e.Statements)
}
