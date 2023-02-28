package providers

import (
	"fmt"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/doxcontext"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/project/definition"
	"github.com/laytan/elephp/internal/symbol"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/nodeident"
	"github.com/laytan/elephp/pkg/phpdoxer"
)

type CommentsProvider struct{}

func NewComments() *CommentsProvider {
	return &CommentsProvider{}
}

func (p *CommentsProvider) CanDefine(ctx *context.Ctx, kind ir.NodeKind) bool {
	if c, _ := ctx.InComment(); c != nil {
		return true
	}

	return false
}

func (p *CommentsProvider) Define(ctx *context.Ctx) ([]*definition.Definition, error) {
	c, i := ctx.InComment()

	nodes, err := phpdoxer.ParseDoc(string(c.Value))
	if err != nil {
		return nil, fmt.Errorf("[definition.CommentsProvider.Define]: %w", err)
	}

	var nodeAtCursor phpdoxer.Node
	for _, node := range nodes {
		start, end := node.Range()
		if int(i) >= start && int(i) <= end {
			nodeAtCursor = node
			break
		}
	}

	var typ phpdoxer.Type

	switch typedNode := nodeAtCursor.(type) {
	case *phpdoxer.NodeThrows:
		typ = typedNode.Type
	case *phpdoxer.NodeReturn:
		typ = typedNode.Type
	case *phpdoxer.NodeVar:
		typ = typedNode.Type
	case *phpdoxer.NodeParam:
		typ = typedNode.Type
	case *phpdoxer.NodeInheritDoc:
		m, ok := ctx.Current().(*ir.ClassMethodStmt)
		if !ok {
			return nil, fmt.Errorf("got %T node with @inheritdoc, which is not supported", ctx.Current())
		}

		return p.defineInheritDoc(ctx, m)
	}

	var results []*definition.Definition
	if typ == nil {
		return results, nil
	}

	fqnt := fqn.NewTraverser()
	ctx.Root().Walk(fqnt)

	var currFQN *fqn.FQN
	clsNode := ctx.ClassScope()
	if ir.GetNodeKind(clsNode) != ir.KindRoot {
		currFQN = fqnt.ResultFor(&ir.Name{
			Position: ir.GetPosition(clsNode),
			Value:    nodeident.Get(clsNode),
		})
	}

	clsses := doxcontext.ApplyContext(fqnt, currFQN, ir.GetPosition(ctx.Current()), typ)
	for _, cls := range clsses {
		// After ApplyContext, the returned types are all fully qualified.
		if iNode, ok := index.FromContainer().Find(fqn.New(cls.Name)); ok {
			results = append(results, definition.IndexNodeToDef(iNode))
		}
	}

	return results, nil
}

func (p *CommentsProvider) defineInheritDoc(
	ctx *context.Ctx,
	m *ir.ClassMethodStmt,
) ([]*definition.Definition, error) {
	meth := symbol.NewMethod(ctx, m)
	cls, err := symbol.NewClassLikeFromMethod(ctx.Root(), m)
	if err != nil {
		return nil, fmt.Errorf("[CommentsProvider.defineInheritDoc]: %w", err)
	}

	iter := cls.InheritsIter()
	for inhCls, done, err := iter(); !done; inhCls, done, err = iter() {
		if err != nil {
			return nil, fmt.Errorf("[CommentsProvider.defineInheritDoc]: %w", err)
		}

		inhMeth := inhCls.FindMethod(symbol.FilterOverwrittenBy(meth))
		if inhMeth != nil {
			return []*definition.Definition{{
				Path:       inhCls.Path(),
				Position:   inhMeth.Node().Position,
				Identifier: inhMeth.Name(),
			}}, nil
		}
	}

	return nil, fmt.Errorf(
		"[CommentsProvider.defineInheritDoc]: @inheritdoc, but the method has no parent method",
	)
}
