package expr

import (
	"log"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/project/definition"
	"github.com/laytan/elephp/internal/project/definition/providers"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/elephp/pkg/symbol"
	"github.com/laytan/elephp/pkg/traversers"
)

var starters = map[ExprType]StartResolver{
	ExprTypeVariable: &variableResolver{},
	ExprTypeName:     &nameResolver{},
	ExprTypeFunction: &functionResolver{},
}

type variableResolver struct{}

func (p *variableResolver) Down(
	node ir.Node,
) (*DownResolvement, ir.Node, bool) {
	propNode, ok := node.(*ir.SimpleVar)
	if !ok {
		return nil, nil, false
	}

	return &DownResolvement{
		ExprType:   ExprTypeVariable,
		Identifier: symbol.GetIdentifier(propNode),
	}, nil, true
}

// TODO: this is so cursed.
// TODO: should not be using the definition package.
// TODO: definition.VarType should just be a recursive call to this expr flow.
func (p *variableResolver) Up(
	scopes Scopes,
	toResolve *DownResolvement,
) (*phpdoxer.TypeClassLike, bool) {
	if toResolve.ExprType != ExprTypeVariable {
		return nil, false
	}

	t := traversers.NewVariable(toResolve.Identifier)
	scopes.Block.Walk(t)
	if t.Result == nil {
		return nil, false
	}

	// TODO: don't create new context, make VariableType accept normal vars.
	content, err := Wrkspc().ContentOf(scopes.Path)
	if err != nil {
		log.Println(err)
		return nil, false
	}

	row, col := position.PosToLoc(content, uint(t.Result.Position.StartPos))
	pos := &position.Position{
		Row:  row,
		Col:  col,
		Path: scopes.Path,
	}
	ctx, err := context.New(pos)
	if err != nil {
		log.Println(err)
		return nil, false
	}

	// TODO: what does privacy do here.
	def, _, err := definition.VariableType(ctx, t.Result)
	if err != nil {
		log.Println(err)
		return nil, false
	}

	fqn := definition.FullyQualify(ctx.Root(), def.Node.Identifier())
	return &phpdoxer.TypeClassLike{
		Name:           fqn.String(),
		FullyQualified: true,
	}, true
}

type nameResolver struct{}

func (p *nameResolver) Down(
	node ir.Node,
) (*DownResolvement, ir.Node, bool) {
	propNode, ok := node.(*ir.Name)
	if !ok {
		return nil, nil, false
	}

	return &DownResolvement{
		ExprType:   ExprTypeName,
		Identifier: symbol.GetIdentifier(propNode),
	}, nil, true
}

func (p *nameResolver) Up(
	scopes Scopes,
	toResolve *DownResolvement,
) (*phpdoxer.TypeClassLike, bool) {
	if toResolve.ExprType != ExprTypeName {
		return nil, false
	}

	fqn := definition.FullyQualify(scopes.Root, toResolve.Identifier)
	return &phpdoxer.TypeClassLike{
		Name:           fqn.String(),
		FullyQualified: true,
	}, true
}

type functionResolver struct{}

func (p *functionResolver) Down(
	node ir.Node,
) (*DownResolvement, ir.Node, bool) {
	propNode, ok := node.(*ir.FunctionCallExpr)
	if !ok {
		return nil, nil, false
	}

	return &DownResolvement{
		ExprType:   ExprTypeFunction,
		Identifier: symbol.GetIdentifier(propNode),
	}, nil, true
}

func (p *functionResolver) Up(
	scopes Scopes,
	toResolve *DownResolvement,
) (*phpdoxer.TypeClassLike, bool) {
	if toResolve.ExprType != ExprTypeFunction {
		return nil, false
	}

	t := traversers.NewFunctionCall(toResolve.Identifier)
	scopes.Block.Walk(t)
	if t.Result == nil {
		log.Printf("could not find function matching function call %s", toResolve.Identifier)
		return nil, false
	}

	def, err := providers.DefineFunction(scopes.Path, scopes.Root, scopes.Block, t.Result)
	if err != nil {
		log.Println(err)
		return nil, false
	}

	if n, ok := defToNode(scopes.Root, def); ok {
		ret := Typer().Returns(scopes.Root, n)
		if clsRet, ok := ret.(*phpdoxer.TypeClassLike); ok {
			return clsRet, true
		}

		return nil, false
	}

	return nil, false
}

func defToNode(root *ir.Root, def *definition.Definition) (ir.Node, bool) {
	t := traversers.NewSymbolToNode(def.Node)
	root.Walk(t)

	if t.Result == nil {
		log.Printf("Symbol to node traverser returned nil for %v in %s", def.Node, def.Path)
		return nil, false
	}

	return t.Result, true
}
