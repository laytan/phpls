package symbol

import (
	"fmt"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/functional"
	"github.com/laytan/elephp/pkg/phpdoxer"
)

type parametized struct {
	rooter
	named

	*doxed

	node ir.Node
}

func (p *parametized) Parameters() ([]*Parameter, error) {
	paramNodes, err := p.paramNodes()
	if err != nil {
		return nil, fmt.Errorf("getting parameter nodes for %v: %w", p.node, err)
	}

	return functional.Map(
		paramNodes,
		func(pNode *ir.Parameter) *Parameter {
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

func (p *parametized) paramNodes() ([]*ir.Parameter, error) {
	switch typedNode := p.node.(type) {
	case *ir.FunctionStmt:
		return functional.Map(
			typedNode.Params,
			func(p ir.Node) *ir.Parameter { return p.(*ir.Parameter) },
		), nil
	case *ir.ClassMethodStmt:
		return functional.Map(
			typedNode.Params,
			func(p ir.Node) *ir.Parameter { return p.(*ir.Parameter) },
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
