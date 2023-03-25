// nolint: gocritic // hugeParam for dump not prio to address.
package ast

import (
	"io"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/laytan/elephp/pkg/phparser/token"
)

type (
	Value interface {
		Node
		value()
	}
	Operation interface {
		Node
		operation()
	}
)

var (
	ValueImpls = []Value{
		Number{},
		Var{},
		SimpleString{},
		ComplexString{},
		Bool{},
		Call{},
		Group{},
		Not{},
		New{},
		ErrorSuppress{},
		Constant{},
	}
	OperationImpls = []Operation{
		CommonOperation{},
		IndexOperation{},
		MethodCallOperation{},
		PropertyFetchOperation{},
		IfOperation{},
	}
)

type Number struct {
	BaseNode

	Value string `parser:"@Number"`
}

var _ Value = Number{}

func (n Number) value() {}

func (n Number) Dump(w io.Writer, level int) {
	n.BaseDump(w, level, "Number")
	level++
	writeStr(w, "Value: %s", n.Value)
	writeEOL(w, level)
}

type Bool struct {
	BaseNode

	Value string `parser:"@( True | False )"`
}

var _ Value = Bool{}

func (b Bool) value() {}

func (b Bool) Dump(w io.Writer, level int) {
	b.BaseDump(w, level, "Bool")
	level++
	writeStr(w, "Value: %s", b.Value)
	writeEOL(w, level)
}

type Var struct {
	BaseNode

	Value string `parser:"@Var"`
}

var _ Value = Var{}

func (v Var) value()             {}
func (v Var) complexStrContent() {}

func (v Var) Dump(w io.Writer, level int) {
	v.BaseDump(w, level, "Var")
	level++
	writeStr(w, "Value: %s", v.Value)
	writeEOL(w, level)
}

type Constant struct {
	BaseNode

	Value string `parser:"@Ident"`
}

var _ Value = Constant{}

func (c Constant) value() {}

func (c Constant) Dump(w io.Writer, level int) {
	c.BaseDump(w, level, "Constant")
	level++
	writeStr(w, "Value: %s", c.Value)
	writeEOL(w, level)
}

type Call struct {
	BaseNode

	Name       string `parser:"@Ident"`
	Parameters []Expr `parser:"LParen @@? ( Comma @@ )* RParen"`
}

var _ Node = Call{}

func (c Call) value() {}

func (c Call) Dump(w io.Writer, level int) {
	c.BaseDump(w, level, "Call")
	level++
	writeStr(w, "Name: %s", c.Name)
	writeEOL(w, level)
	writeStr(w, "Parameters:")
	level++
	writeEOL(w, level)
	writeList(w, level, c.Parameters)
}

type Group struct {
	BaseNode

	Value Expr `parser:"LParen @@ RParen"`
}

var _ Node = Group{}

func (g Group) value() {}

func (g Group) Dump(w io.Writer, level int) {
	g.BaseDump(w, level, "Group")
	level++
	writeStr(w, "Value:")
	level++
	writeEOL(w, level)
	g.Value.Dump(w, level)
}

type New struct {
	BaseNode

	Value Expr `parser:"New @@"`
}

func (n New) value() {}

func (n New) Dump(w io.Writer, level int) {
	n.BaseDump(w, level, "New")
	level++
	writeStr(w, "Value:")
	writeEOL(w, level)
	n.Value.Dump(w, level)
}

type Not struct {
	BaseNode

	Value Expr `parser:"Not @@"`
}

var _ Node = Not{}

func (n Not) value() {}

func (n Not) Dump(w io.Writer, level int) {
	n.BaseDump(w, level, "Not")
	level++
	writeStr(w, "Value:")
	level++
	writeEOL(w, level)
	n.Value.Dump(w, level)
}

type ErrorSuppress struct {
	BaseNode

	Value Expr `parser:"ErrorSuppress @@"`
}

var _ Node = ErrorSuppress{}

func (e ErrorSuppress) value() {}

func (e ErrorSuppress) Dump(w io.Writer, level int) {
	e.BaseDump(w, level, "ErrorSuppress")
	level++
	writeStr(w, "Value")
	level++
	writeEOL(w, level)
	e.Value.Dump(w, level)
}

type Expr struct {
	BaseNode

	Left       Value       `parser:"@@"`
	Operations []Operation `parser:"@@*"`
}

var _ Node = Expr{}

func (e Expr) complexStrContent() {}

func (e Expr) Dump(w io.Writer, level int) {
	e.BaseDump(w, level, "Expr")
	level++
	writeStr(w, "Left:")
	level++
	writeEOL(w, level)
	e.Left.Dump(w, level)
	level--
	writeEOL(w, level)
	writeStr(w, "Operations:")
	level++
	writeEOL(w, level)
	writeList(w, level, e.Operations)
}

type IndexOperation struct {
	BaseNode

	Index *Expr `parser:"LBracket @@? RBracket"`
}

var _ Node = IndexOperation{}

func (i IndexOperation) operation() {}

func (i IndexOperation) Dump(w io.Writer, level int) {
	i.BaseDump(w, level, "IndexOperation")
	level++
	if i.Index != nil {
		writeStr(w, "Index:")
		level++
		writeEOL(w, level)
		i.Index.Dump(w, level)
	}
}

type MethodCallOperation struct {
	BaseNode

	Fetch Call `parser:"ClassAccess @@"`
}

var _ Node = MethodCallOperation{}

func (m MethodCallOperation) operation() {}

func (m MethodCallOperation) Dump(w io.Writer, level int) {
	m.BaseDump(w, level, "MethodCallOperation")
	level++
	writeStr(w, "Fetch:")
	level++
	writeEOL(w, level)
	m.Fetch.Dump(w, level)
}

type PropertyFetchOperation struct {
	BaseNode

	Fetch string `parser:"ClassAccess @Ident"`
}

var _ Node = PropertyFetchOperation{}

func (p PropertyFetchOperation) operation() {}

func (p PropertyFetchOperation) Dump(w io.Writer, level int) {
	p.BaseDump(w, level, "PropertyFetchOperation")
	level++
	writeStr(w, "Fetch: %s", p.Fetch)
	writeEOL(w, level)
}

type IfOperation struct {
	BaseNode

	IfTrue Expr `parser:"QuestionMark @@"`
	Else   Expr `parser:"Colon @@"`
}

func (p IfOperation) operation() {}

func (p IfOperation) Dump(w io.Writer, level int) {
	p.BaseDump(w, level, "IfOperation")
	level++
	writeStr(w, "IfTrue:")
	level++
	writeEOL(w, level)
	p.IfTrue.Dump(w, level)
	level--
	writeEOL(w, level)
	writeStr(w, "Else:")
	level++
	writeEOL(w, level)
	p.Else.Dump(w, level)
}

type CommonOperation struct {
	BaseNode

	Operator lexer.Token `parser:"@( Equals | StrictEquals | NotEquals | StrictNotEquals | Plus | Minus | Divide | Times | Concat | And | Or | BinaryOr | Assign | ConcatAssign )"`
	Right    Value       `parser:"@@"`
}

var _ Node = CommonOperation{}

func (i CommonOperation) operation() {}

func (i CommonOperation) Dump(w io.Writer, level int) {
	i.BaseDump(w, level, "CommonOperation")
	level++
	writeStr(w, "Operator: %s", token.Type(i.Operator.Type))
	writeEOL(w, level)
	writeStr(w, "Right:")
	level++
	writeEOL(w, level)
	i.Right.Dump(w, level)
}
