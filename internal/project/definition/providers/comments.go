package providers

import (
	"fmt"
	"log"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/common"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/project/definition"
	"github.com/laytan/elephp/pkg/phpdoxer"
)

type CommentsProvider struct{}

func NewComments() *CommentsProvider {
	return &CommentsProvider{}
}

func (p *CommentsProvider) CanDefine(ctx context.Context, kind ir.NodeKind) bool {
	if c, _ := ctx.InComment(); c != nil {
		return true
	}

	return false
}

func (p *CommentsProvider) Define(ctx context.Context) ([]*definition.Definition, error) {
	c, i := ctx.InComment()

	nodes, err := phpdoxer.ParseDoc(string(c.Value))
	if err != nil {
		return nil, fmt.Errorf("[project.definition.providers.CommentsProvider.Define]: %w", err)
	}

	var nodeAtCursor phpdoxer.Node
	for _, node := range nodes {
		start, end := node.Range()
		if int(i) >= start && int(i) <= end {
			nodeAtCursor = node
			break
		}
	}

	var clsLike *phpdoxer.TypeClassLike

	switch typedNode := nodeAtCursor.(type) {
	case *phpdoxer.NodeThrows:
		if c, ok := typedNode.Type.(*phpdoxer.TypeClassLike); ok {
			clsLike = c
		}
	case *phpdoxer.NodeReturn:
		if c, ok := typedNode.Type.(*phpdoxer.TypeClassLike); ok {
			clsLike = c
		}
	case *phpdoxer.NodeVar:
		if c, ok := typedNode.Type.(*phpdoxer.TypeClassLike); ok {
			clsLike = c
		}
	case *phpdoxer.NodeParam:
		if c, ok := typedNode.Type.(*phpdoxer.TypeClassLike); ok {
			clsLike = c
		}
	case *phpdoxer.NodeInheritDoc:
		log.Println("[project.definition.providers.CommentsProvider.Define]: TODO: inherit doc comment definition")
	}

	results := []*definition.Definition{}

	if clsLike == nil {
		return results, nil
	}

	if res, ok := common.FindFullyQualifiedName(ctx.Root(), &ir.Name{
		Value:    clsLike.Name,
		Position: c.Position,
	}); ok {
		results = append(results, definition.IndexNodeToDef(res))
	}

	return results, nil
}
