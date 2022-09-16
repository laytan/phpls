package providers

import (
	"log"

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

func (p *MethodProvider) CanDefine(ctx context.Context, kind ir.NodeKind) bool {
	return kind == ir.KindMethodCallExpr
}

func (p *MethodProvider) Define(ctx context.Context) (*definition.Definition, error) {
	n := ctx.Current().(*ir.MethodCallExpr)

	def, err := definition.WalkToMethod(ctx, n)
	if err != nil {
		log.Println(err)
		return nil, definition.ErrNoDefinitionFound
	}

	return def, nil
}
