package providers

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/project/definition"
	"github.com/laytan/elephp/pkg/symbol"
)

// ThisProvider resolves the definition of the current class scope for '$this'
// variables.
type ThisProvider struct{}

func NewThis() *ThisProvider {
	return &ThisProvider{}
}

func (p *ThisProvider) CanDefine(ctx context.Context, kind ir.NodeKind) bool {
	if kind != ir.KindSimpleVar {
		return false
	}

	n := ctx.Current().(*ir.SimpleVar)
	return n.Name == "this"
}

func (p *ThisProvider) Define(ctx context.Context) (*definition.Definition, error) {
	if ir.GetNodeKind(ctx.Current()) == ir.KindRoot {
		return nil, definition.ErrNoDefinitionFound
	}

	return &definition.Definition{
		Path: ctx.Start().Path,
		Node: symbol.New(ctx.ClassScope()),
	}, nil
}
