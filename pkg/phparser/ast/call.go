package ast

type Call struct {
	Name       string `parser:"@Ident"`
	Parameters []Expr `parser:"LParen @@? ( Comma @@ )* RParen"`
}

func (c Call) expr()      {}
func (c Call) exprNoArr() {}

type CallStatement struct {
	Name       string `parser:"@Ident"`
	Parameters []Expr `parser:"LParen @@? ( Comma @@ )* RParen"`
	Semicolon  string `parser:"@Semicolon"`
}

func (m CallStatement) statement()         {}
func (m CallStatement) classStatement()    {}
func (m CallStatement) callableStatement() {}
