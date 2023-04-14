package symbol

import (
	"errors"
	"fmt"

	"github.com/laytan/phpls/internal/doxcontext"
	"github.com/laytan/phpls/pkg/fqn"
	"github.com/laytan/phpls/pkg/nodeident"
	"github.com/laytan/phpls/pkg/phpdoxer"
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor/traverser"
)

var ErrNoVarType = errors.New("variable has no @var node")

type Variable struct {
	rooter

	*doxed

	node *ast.ExprVariable
}

func NewVariable(root rooter, node *ast.ExprVariable) *Variable {
	return &Variable{
		rooter: root,
		doxed:  NewDoxed(node),
		node:   node,
	}
}

// Type checks for @var comments and returns the type.
// NOTE: this does not check any assignment or try to be smart.
func (v *Variable) Type() (phpdoxer.Type, error) {
	doc := v.FindDoc(FilterDocKind(phpdoxer.KindVar))
	if doc != nil {
		return doc.(*phpdoxer.NodeVar).Type, nil
	}

	return nil, fmt.Errorf("checking %s for @var: %w", v.Name(), ErrNoVarType)
}

// currFqn can be nil if the variable is not in a class.
func (v *Variable) TypeCls(currFqn *fqn.FQN) ([]*phpdoxer.TypeClassLike, error) {
	doc, err := v.Type()
	if err != nil {
		return nil, fmt.Errorf("getting type of %s var to apply context to: %w", v.Name(), err)
	}

	fqnt := fqn.NewTraverser()
	tv := traverser.NewTraverser(fqnt)
	v.Root().Accept(tv)

	return doxcontext.ApplyContext(fqnt, currFqn, v.node.Position, doc), nil
}

func (v *Variable) Name() string {
	return nodeident.Get(v.node)
}
