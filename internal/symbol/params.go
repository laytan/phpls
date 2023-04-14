package symbol

import (
	"fmt"

	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/phpls/pkg/functional"
	"github.com/laytan/phpls/pkg/phpdoxer"
)

type parametized struct {
	rooter
	named

	*doxed

	node ast.Vertex
}

func (p *parametized) Parameters() ([]*Parameter, error) {
	paramNodes, err := p.paramNodes()
	if err != nil {
		return nil, fmt.Errorf("getting parameter nodes for %v: %w", p.node, err)
	}

	return functional.Map(
		paramNodes,
		func(pNode *ast.Parameter) *Parameter {
			return &Parameter{
				funcOrMeth: p.node,
				node:       pNode,
			}
		},
	), nil
}

func (p *parametized) FindParameter(filters ...FilterFunc[*Parameter]) (*Parameter, error) {
	params, err := p.Parameters()
	if err != nil {
		return nil, fmt.Errorf("retrieving parameters to filter: %w", err)
	}

ParamsRange:
	for _, param := range params {
		for _, filter := range filters {
			if !filter(param) {
				continue ParamsRange
			}
		}

		return param, nil
	}

	return nil, fmt.Errorf("no results: %w", ErrNoParam)
}

func (p *parametized) paramNodes() ([]*ast.Parameter, error) {
	switch typedNode := p.node.(type) {
	case *ast.StmtFunction:
		return functional.Map(
			typedNode.Params,
			func(p ast.Vertex) *ast.Parameter { return p.(*ast.Parameter) },
		), nil
	case *ast.StmtClassMethod:
		return functional.Map(
			typedNode.Params,
			func(p ast.Vertex) *ast.Parameter { return p.(*ast.Parameter) },
		), nil
	default:
		return nil, fmt.Errorf("Node with type %T is invalid inside *parametized", p.node)
	}
}

func FilterParamName(name string) DocFilter {
	if name[0:1] != "$" {
		name = "$" + name
	}

	return func(n phpdoxer.Node) bool {
		tn, ok := n.(*phpdoxer.NodeParam)
		if !ok {
			return false
		}

		return tn.Name == name
	}
}
