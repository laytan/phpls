package symbol

import (
	"errors"
	"fmt"
	"log"

	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor/traverser"
	"github.com/laytan/phpls/internal/doxcontext"
	"github.com/laytan/phpls/pkg/fqn"
	"github.com/laytan/phpls/pkg/phpdoxer"
)

var ErrNoReturn = errors.New("node has no return type")

type canReturn struct {
	*doxed
	rooter

	node ast.Vertex
}

// Returns gets the return type of this callable.
// Traversing through inherited methods with {@inheritdoc}.
//
// If the node is a method, the 2nd argument is the enclosing class
// of the method that had the type definition on it.
func (r *canReturn) Returns() (phpdoxer.Type, *ClassLike, error) {
	rh, err := r.ownReturns()
	if err != nil && !errors.Is(err, ErrNoReturn) {
		return nil, nil, fmt.Errorf("parsing own return of %v: %w", r.node, err)
	}
	if rh != nil {
		if method, ok := r.node.(*ast.StmtClassMethod); ok {
			cls, err := r.class()
			if err != nil {
				return nil, nil, fmt.Errorf("retrieving enclosing class of %v: %w", method, err)
			}

			return rh, cls, nil
		}

		return rh, nil, nil
	}

	rc, cls, err := r.returnsComment()
	if err != nil {
		return nil, nil, fmt.Errorf("inherited of %v doc return comments: %w", r.node, err)
	}

	return rc, cls, nil
}

// ReturnsClass resolves and unpacks the raw type returned from Returns into
// the classes it represents.
// See doxcontext.ApplyContext for more.
func (r *canReturn) ReturnsClass() ([]*phpdoxer.TypeClassLike, error) { //nolint:dupl // Not really duplicated, could extract later.
	ret, cls, err := r.Returns()
	if err != nil {
		return nil, fmt.Errorf("getting return type to apply context: %w", err)
	}

	fqnt := fqn.NewTraverser()
	fqntt := traverser.NewTraverser(fqnt)
	var currFqn *fqn.FQN
	var node ast.Vertex

	switch typedCallable := r.node.(type) {
	case *ast.StmtClassMethod:
		root := cls.Root()
		root.Accept(fqntt)

		currFqn = cls.GetFQN()
		node = typedCallable
	case *ast.StmtFunction:
		root := r.Root()
		root.Accept(fqntt)

		// If we are not a method this one does not really matter for ApplyContext.
		currFqn = fqn.New(fqn.PartSeperator)
		node = typedCallable
	default:
		return nil, fmt.Errorf("*canReturn with invalid callable %T", typedCallable)
	}

	return doxcontext.ApplyContext(fqnt, currFqn, node.GetPosition(), ret), nil
}

// NOTE: it is assumed that ownReturns is checked already.
func (r *canReturn) returnsComment() (phpdoxer.Type, *ClassLike, error) {
	methNode, ok := r.node.(*ast.StmtClassMethod)
	if !ok {
		return nil, nil, fmt.Errorf("checking node is class method: %w", ErrNoReturn)
	}

	meth := NewMethod(r.rooter, methNode)
	cls, err := NewClassLikeFromMethod(r.Root(), methNode)
	if err != nil {
		return nil, nil, fmt.Errorf("getting class from method: %w", err)
	}

	iter := cls.InheritsIter()
	for inhCls, done, err := iter(); !done; inhCls, done, err = iter() {
		if err != nil {
			return nil, nil, fmt.Errorf("inherits of classlike %s: %w", cls.Name(), err)
		}

		inhMeth := inhCls.FindMethod(FilterOverwrittenBy(meth))
		if inhMeth == nil {
			continue
		}

		retType, err := inhMeth.ownReturns()
		if err != nil {
			if errors.Is(err, ErrNoReturn) {
				continue
			}

			return nil, nil, fmt.Errorf("getting inherited method return type: %w", err)
		}

		return retType, inhCls, nil
	}

	return nil, nil, fmt.Errorf("checked all inherited methods: %w", ErrNoReturn)
}

// ownReturns returns the type that this callable returns, without traversing inherited methods.
func (r *canReturn) ownReturns() (phpdoxer.Type, error) {
	docNode := r.FindDoc(FilterDocKind(phpdoxer.KindReturn))
	if docNode != nil {
		return docNode.(*phpdoxer.NodeReturn).Type, nil
	}

	hintType, err := r.returnsHint()
	if err != nil {
		return nil, fmt.Errorf("getting return hint: %w", err)
	}

	return hintType, nil
}

func (r *canReturn) returnsHint() (phpdoxer.Type, error) {
	retNode := r.returnsHintNode()
	if retNode == nil {
		return nil, fmt.Errorf("no return hint: %w", ErrNoReturn)
	}

	return TypeHintToDocType(retNode)
}

func (r *canReturn) returnsHintNode() ast.Vertex {
	switch tNode := r.node.(type) {
	case *ast.StmtFunction:
		return tNode.ReturnType
	case *ast.StmtClassMethod:
		return tNode.ReturnType
	default:
		log.Panicf("*canReturn with invalid node of type %T", tNode)
		// already panicked.
		return nil
	}
}

func (r *canReturn) class() (*ClassLike, error) {
	switch typedNode := r.node.(type) {
	case *ast.StmtClassMethod:
		cls, err := NewClassLikeFromMethod(r.Root(), typedNode)
		if err != nil {
			return nil, fmt.Errorf("retrieving class of method %v: %w", typedNode, err)
		}

		return cls, nil
	default:
		return nil, ErrNotAMethod
	}
}
