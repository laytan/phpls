package typer

import "github.com/VKCOM/noverify/src/ir"

type Typer interface {
	// Call with either a ir.ClassMethodStmt or ir.FunctionStmt.
	Returns(Root *ir.Root, funcOrMeth ir.Node) (Type, error)

	// Call with either a ir.ClassMethodStmt or ir.FunctionStmt.
	Param(Root *ir.Root, funcOrMeth ir.Node, param *ir.Parameter) (Type, error)

	// Scope should be the method/function the variable is used in, if it is used
	// globally, this can be left nil.
	Variable(Root *ir.Root, variable *ir.SimpleVar, scope ir.Node) (Type, error)
}

type Type interface {
	IsScalar() bool

	// Either a Scalar or FQN type.
	Value() ScalarOrFQN
}

type ScalarOrFQN interface {
	String() string
}

type (
	Scalar struct{}
	FQN    struct{}
)

// TODO: Move fully qualified name logic & traverser here.
