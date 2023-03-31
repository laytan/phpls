package symbol

import (
	"errors"
	"fmt"

	"github.com/laytan/elephp/internal/doxcontext"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/nodeident"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor/traverser"
)

var (
	ErrNoParam    = errors.New("parameter has no type")
	ErrNotAMethod = errors.New("does not live on a method")
)

type Parameter struct {
	rooter

	// Either a *ast.StmtFunction, or a *ast.StmtClassMethod.
	funcOrMeth ast.Vertex
	node       *ast.Parameter
}

func NewParameter(root rooter, funcOrMeth ast.Vertex, param *ast.Parameter) *Parameter {
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
	method, isMethod := p.funcOrMeth.(*ast.StmtClassMethod)

	typ, err := p.ownType()
	if err != nil && !errors.Is(err, ErrNoParam) {
		return nil, nil, fmt.Errorf("getting own type %v: %w", p.node, err)
	}
	if typ != nil {
		if isMethod {
			cls, err := p.class()
			if err != nil {
				return nil, nil, fmt.Errorf("getting containing class %v: %w", method, err)
			}

			return typ, cls, nil
		}

		return typ, nil, nil //nolint:unsafenil // First return is the result, rest is documented.
	}

	if !isMethod {
		return nil, nil, fmt.Errorf("no own type and not a method: %w", ErrNoParam)
	}

	cls, err := p.class()
	if err != nil {
		return nil, nil, fmt.Errorf("getting containing class %v: %w", method, err)
	}

	startMethod := NewMethod(p.rooter, method)
	iter := cls.InheritsIter()
	for inhCls, done, err := iter(); !done; inhCls, done, err = iter() {
		if err != nil {
			return nil, nil, fmt.Errorf(
				"get inherited of class %s for parameter %s: %w",
				cls.Name(),
				p.Name(),
				err,
			)
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
func (p *Parameter) TypeClass() ([]*phpdoxer.TypeClassLike, error) { //nolint:dupl // Not really duplicated, could extract later.
	typ, cls, err := p.Type()
	if err != nil {
		return nil, fmt.Errorf("getting type to apply context: %w", err)
	}

	fqnt := fqn.NewTraverser()
	fqntt := traverser.NewTraverser(fqnt)
	var currFqn *fqn.FQN
	var node ast.Vertex

	switch typedCallable := p.funcOrMeth.(type) {
	case *ast.StmtClassMethod:
		root := cls.Root()
		root.Accept(fqntt)

		currFqn = cls.GetFQN()
		node = typedCallable
	case *ast.StmtFunction:
		root := p.Root()
		root.Accept(fqntt)

		// If we are not a method this one does not really matter for ApplyContext.
		currFqn = fqn.New(fqn.PartSeperator)
		node = typedCallable
	default:
		return nil, fmt.Errorf("Parameter with invalid callable %T", typedCallable)
	}

	return doxcontext.ApplyContext(fqnt, currFqn, node.GetPosition(), typ), nil
}

// class returns the class this method belongs to or nil if it is a function.
func (p *Parameter) class() (*ClassLike, error) {
	method, ok := p.funcOrMeth.(*ast.StmtClassMethod)
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

func (p *Parameter) Node() *ast.Parameter {
	return p.node
}

// ownType returns the type of this parameter, only retrieved from this method.
func (p *Parameter) ownType() (phpdoxer.Type, error) {
	doxer := NewDoxed(p.funcOrMeth)
	paramDoc := doxer.FindDoc(FilterParamName(p.Name()))
	if paramDoc != nil && paramDoc.(*phpdoxer.NodeParam).Type != nil {
		return paramDoc.(*phpdoxer.NodeParam).Type, nil
	}

	if p.node.Type != nil {
		parsedHint, err := TypeHintToDocType(p.node.Type)
		if err != nil {
			return nil, fmt.Errorf("parsing type hint for parameter %v: %w", p.node, err)
		}

		return parsedHint, nil
	}

	return nil, ErrNoParam
}
