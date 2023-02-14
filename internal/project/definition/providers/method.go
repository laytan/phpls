package providers

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/project/definition"
)

// MethodProvider resolves the definition of a method call.
// $this->test(), $this->test->test(), $foo->bar() etc. for example.
type MethodProvider struct{}

func NewMethod() *MethodProvider {
	return &MethodProvider{}
}

func (p *MethodProvider) CanDefine(ctx *context.Ctx, kind ir.NodeKind) bool {
	return kind == ir.KindMethodCallExpr
}

func (p *MethodProvider) Define(ctx *context.Ctx) ([]*definition.Definition, error) {
	return DefineExpr(ctx)
}
