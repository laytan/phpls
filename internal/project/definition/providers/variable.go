package providers

import (
	"log"

	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor/traverser"
	"github.com/laytan/phpls/internal/context"
	"github.com/laytan/phpls/internal/project/definition"
	"github.com/laytan/phpls/pkg/nodeident"
	"github.com/laytan/phpls/pkg/traversers"
)

// VariableProvider resolves the definition of a simple variable.
// It looks for it being assigned, as a parameter, or a global.
type VariableProvider struct{}

func NewVariable() *VariableProvider {
	return &VariableProvider{}
}

func (p *VariableProvider) CanDefine(ctx *context.Ctx, kind ast.Type) bool {
	return kind == ast.TypeExprVariable
}

func (p *VariableProvider) Define(ctx *context.Ctx) ([]*definition.Definition, error) {
	varN := ctx.Current().(*ast.ExprVariable)
	varName := nodeident.Get(ctx.Current())

	// If scope is a closure, and that closure has a use statement with this variable.
	// Check the next scope for the definition.
	scope := ctx.Scope()
	nextScope := func() {
		ctx.Advance()
		scope = ctx.Scope()
	}
	for ; scope.GetType() == ast.TypeExprClosure; nextScope() {
		cs := scope.(*ast.ExprClosure)
		if !hasUsageNamed(cs, varName) {
			break
		}
	}

	t := traversers.NewAssignment(varN)
	tt := traverser.NewTraverser(t)
	ctx.Scope().Accept(tt)

	if t.Assignment == nil {
		return nil, definition.ErrNoDefinitionFound
	}

	return []*definition.Definition{{
		Path:       ctx.Start().Path,
		Position:   t.Assignment.Position,
		Identifier: nodeident.Get(t.Assignment),
	}}, nil
}

func hasUsageNamed(node *ast.ExprClosure, name string) bool {
	for _, u := range node.Uses {
		switch tu := u.(type) {
		case *ast.ExprClosureUse:
			switch tv := tu.Var.(type) {
			case *ast.ExprVariable:
				if nodeident.Get(tv.Name) == name {
					return true
				}
			default:
				log.Panicf("unexpected uses variable %T", tu.Var)
			}
		default:
			log.Panicf("unexpected uses node %T", u)
		}
	}
	return false
}
