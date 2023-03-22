package ast

import (
	"github.com/alecthomas/participle/v2/lexer"
)

var (
	StatementImpls = []Statement{
		Namespace{},
		Assign{},
		Class{},
		Function{},
		CallStatement{},
	}
	ClassStatementImpls    = []ClassStatement{Method{}, CallStatement{}}
	CallableStatementImpls = []CallableStatement{Assign{}, CallStatement{}}
)

type (
	Statement         interface{ statement() }
	ClassStatement    interface{ classStatement() }
	CallableStatement interface{ callableStatement() }
)

type Namespace struct {
	BaseNode
	Name lexer.Token `parser:"Namespace @Ident Semicolon"`
}

func (n Namespace) statement() {}

type Assign struct {
	BaseNode
	Name string `parser:"@Var Assign"`
	// TODO: Expressions.
	Value Expr `parser:"@@ Semicolon"`
}

func (v Assign) statement()         {}
func (v Assign) callableStatement() {}

type Class struct {
	BaseNode

	Modifiers  Modifiers        `parser:"@@ Class"`
	Name       Name             `parser:"@@"`
	Extends    *Extends         `parser:"@@?"`
	Implements *Implements      `parser:"@@?"`
	Statements []ClassStatement `parser:"LBrace @@* RBrace"`
}

func (c Class) statement() {}

type Method struct {
	BaseNode

	Modifiers  Modifiers           `parser:"@@"`
	Name       string              `parser:"Function @Ident"`
	Params     []Parameter         `parser:"LParen @@? ( Comma @@ )* RParen"`
	ReturnType *string             `parser:"(Colon @Ident)?"`
	Statements []CallableStatement `parser:"LBrace @@* RBrace"`
}

func (m Method) classStatement() {}

type Function struct {
	BaseNode

	Name       string              `parser:"Function @Ident"`
	Params     []Parameter         `parser:"LParen @@? ( Comma @@ )* RParen"`
	ReturnType *string             `parser:"(Colon @Ident)?"`
	Statements []CallableStatement `parser:"LBrace @@* RBrace"`
}

func (m Function) statement() {}

type Parameter struct {
	Variadic     bool          `parser:"@Variadic?"` // TODO: only the last parameter can be variadic.
	Reference    bool          `parser:"@Reference?"`
	TypeHint     *string       `parser:"@Ident?"`
	Var          string        `parser:"@Var"`
	DefaultValue *DefaultValue `parser:"@@?"`
}

type DefaultValue struct {
	Assign lexer.Token `parser:"@Assign"`
	// TODO: default value can be scalar values, arrays, null, or a (new ClassName())
	Value []lexer.Token `parser:"( @SimpleString | ( @StringStart @StringContent* @StringEnd ) | @Number )"`
}
