package ast

import "github.com/laytan/elephp/pkg/functional"

var (
	ExprNoArrAccImpls []ExprNoArrAcc
	ExprImpls         []Expr
)

func init() {
	ExprNoArrAccImpls = []ExprNoArrAcc{SimpleString{}, Call{}, Var{}}
	ExprImpls = append([]Expr{ArrAccessGroup{}}, noArrAccToExpr(ExprNoArrAccImpls)...)
}

func noArrAccToExpr(es []ExprNoArrAcc) []Expr {
	return functional.Map(es, func(e ExprNoArrAcc) Expr { return e.(Expr) })
}

type (
	Expr         interface{ expr() }
	ExprNoArrAcc interface{ exprNoArr() }
)

type Var struct {
	Value string `parser:"@Var"`
}

func (v Var) expr()      {}
func (v Var) exprNoArr() {}

type ArrAccessGroup struct {
	// Any expression, but not the `ArrAccessGroup` (this struct).
	Value   *ExprNoArrAcc `parser:"@@"`
	Indices []Expr        `parser:"(LBracket @@ RBracket )+"`
}

func (v ArrAccessGroup) expr() {}

type SimpleString struct {
	Value string `parser:"@SimpleString"`
}

func (s SimpleString) expr()      {}
func (s SimpleString) exprNoArr() {}

type IfTernary struct {
	// TODO: add condition and add to union type, probably need something similar to ArrAccessGroup to prevent left recursion.
    // I believe there are some rules where nested ternaries need parentheses, can probably use those rules to our advantage.
	IfTrue  Expr `parser:"QuestionMark @@"`
	IfFalse Expr `parser:"Colon @@"`
}

func (i IfTernary) expr() {}
