package traversers

import (
	"log"

	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor"
	"github.com/laytan/phpls/pkg/nodeident"
	"github.com/laytan/phpls/pkg/nodescopes"
)

type Variable struct {
	visitor.Null
	target   string
	Results  []*ast.ExprVariable
	multiple bool
	isFirst  bool
}

func NewVariable(name string, multiple bool) *Variable {
	return &Variable{isFirst: true, target: name, multiple: multiple}
}

func (v *Variable) EnterNode(node ast.Vertex) bool {
	if !v.multiple && len(v.Results) > 0 {
		return false
	}

	defer func() { v.isFirst = false }()
	if v.isFirst {
		return true
	}

	switch tn := node.(type) {
	case *ast.ExprClosure:
		return hasUsageNamed(tn, v.target)
	case *ast.ExprArrowFunction:
		return true
	default:
		return !nodescopes.IsScope(node.GetType())
	}
}

func (v *Variable) ExprVariable(node *ast.ExprVariable) {
	if nodeident.Get(node.Name) == v.target {
		v.Results = append(v.Results, node)
	}
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
