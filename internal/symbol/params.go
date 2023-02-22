package symbol

import (
	"errors"
	"fmt"
	"log"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/doxcontext"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/functional"
	"github.com/laytan/elephp/pkg/nodeident"
	"github.com/laytan/elephp/pkg/phpdoxer"
)

var (
	ErrNoParam    = errors.New("parameter has no type")
	ErrNotAMethod = errors.New("parameter does not live on a method")
)

type Parameter struct {
	rooter

	// Either a *ir.FunctionStmt, or a *ir.ClassMethodStmt.
	funcOrMeth ir.Node
	node       *ir.Parameter
}

func NewParameter(root rooter, funcOrMeth ir.Node, param *ir.Parameter) *Parameter {
	return &Parameter{
		rooter:     root,
		funcOrMeth: funcOrMeth,
		node:       param,
	}
}

// Type will get the parameter's type, if Parameter.funcOrMeth is a *Method, it also checks any
// inherited methods for a type definition.
//
// If the funcOrMeth is a method, the containing class of method that had a type definition is
// returned as the 2nd return. This is nil when the funcOrMeth is a function.
func (p *Parameter) Type() (phpdoxer.Type, *ClassLike, error) {
	method, isMethod := p.funcOrMeth.(*ir.ClassMethodStmt)

	typ, err := p.ownType()
	if err != nil && !errors.Is(err, ErrNoParam) {
		return nil, nil, fmt.Errorf("getting own type %v: %w", p.node, err)
	}
	if typ != nil {
		if isMethod {
			cls, err := p.Class()
			if err != nil {
				return nil, nil, fmt.Errorf("getting containing class %v: %w", method, err)
			}

			return typ, cls, nil
		}

		return typ, nil, nil
	}

	if !isMethod {
		return nil, nil, fmt.Errorf("no own type and not a method: %w", ErrNoParam)
	}

	cls, err := p.Class()
	if err != nil {
		return nil, nil, fmt.Errorf("getting containing class %v: %w", method, err)
	}

	startMethod := NewMethod(p.rooter, method)
	iter := cls.InheritsIter()
	for inhCls, done, err := iter(); !done; inhCls, done, err = iter() {
		if err != nil {
			log.Println(fmt.Errorf(
				"error retrieving inherited classes of class %s for method %s for parameter %s: %w",
				cls.Name(),
				nodeident.Get(p.funcOrMeth),
				p.Name(),
				err,
			))
			continue
		}

		meth := inhCls.FindMethod(FilterOverwrittenBy(startMethod))
		if meth == nil {
			continue
		}

		param, err := meth.FindParameter(FilterName[*Parameter](p.Name()))
		if err != nil && !errors.Is(err, ErrNoParam) {
			return nil, nil, fmt.Errorf("find parameter %s in %v: %w", p.Name(), meth.Node(), err)
		}
		if param == nil {
			continue
		}

		typ, err := param.ownType()
		if err != nil {
			if errors.Is(err, ErrNoParam) {
				continue
			}

			return nil, nil, fmt.Errorf("getting type of inherited param %v: %w", param, err)
		}

		return typ, inhCls, nil
	}

	return nil, nil, fmt.Errorf("checked all inherited methods/params: %w", ErrNoParam)
}

// TypeClass resolves and unpacks the raw type returned from Type into the classes it represents.
// See doxcontext.ApplyContext for more.
func (p *Parameter) TypeClass() ([]*phpdoxer.TypeClassLike, error) {
	typ, cls, err := p.Type()
	if err != nil {
		return nil, fmt.Errorf("getting type to apply context: %w", err)
	}

	fqnt := fqn.NewTraverser()
	var currFqn *fqn.FQN
	var node ir.Node

	switch typedCallable := p.funcOrMeth.(type) {
	case *ir.ClassMethodStmt:
		root := cls.Root()
		root.Walk(fqnt)

		currFqn = cls.GetFQN()
		node = typedCallable
	case *ir.FunctionStmt:
		root := p.Root()
		root.Walk(fqnt)

		// If we are not a method this one does not really matter for ApplyContext.
		currFqn = fqn.New(fqn.PartSeperator)
		node = typedCallable
	default:
		return nil, fmt.Errorf("Parameter with invalid callable %T", typedCallable)
	}

	return doxcontext.ApplyContext(fqnt, currFqn, ir.GetPosition(node), typ), nil
}

// Class returns the class this method belongs to or nil if it is a function.
func (p *Parameter) Class() (*ClassLike, error) {
	method, ok := p.funcOrMeth.(*ir.ClassMethodStmt)
	if !ok {
		return nil, ErrNotAMethod
	}

	cls, err := NewClassLikeFromMethod(p.Root(), method)
	if err != nil {
		return nil, fmt.Errorf("retrieving class for method %v: %w", method, err)
	}

	return cls, nil
}

func (p *Parameter) Name() string {
	return nodeident.Get(p.node)
}

func (p *Parameter) Node() *ir.Parameter {
	return p.node
}

// ownType returns the type of this parameter, only retrieved from this method.
func (p *Parameter) ownType() (phpdoxer.Type, error) {
	doxer := NewDoxed(p.funcOrMeth)
	paramDoc := doxer.FindDoc(FilterParamName(p.Name())).(*phpdoxer.NodeParam)
	if paramDoc != nil && paramDoc.Type != nil {
		return paramDoc.Type, nil
	}

	if p.node.VariableType != nil {
		parsedHint, err := TypeHintToDocType(p.node.VariableType)
		if err != nil {
			return nil, fmt.Errorf("parsing type hint for parameter %v: %w", p.node, err)
		}

		return parsedHint, nil
	}

	return nil, ErrNoParam
}

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
