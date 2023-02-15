package symbol

import (
	"errors"
	"fmt"
	"log"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/phpdoxer"
)

var ErrNoReturn = errors.New("node has no return type")

type canReturn struct {
	doxed *doxed
	rooter
	node ir.Node

	cache   phpdoxer.Type
	cacheOk bool

	clsCache []*phpdoxer.TypeClassLike // Should start out nil.
}

// Returns gets the return type of this callable.
// Traversing through inherited methods with {@inheritdoc}.
// Returns &phpdoxer.TypeMixed{} by default on errors or no return.
// It returns a boolean indicating whether the default mixed is used.
func (r *canReturn) Returns() (phpdoxer.Type, bool) {
	if r.cache != nil {
		return r.cache, r.cacheOk
	}

	rc, err := r.returnsComment()
	if err == nil {
		r.cache = rc
		r.cacheOk = true
		return r.cache, r.cacheOk
	}

	if !errors.Is(err, ErrNoReturn) {
		log.Println(fmt.Errorf("[callable.canReturn.Returns]: err getting doc return: %w", err))
		r.cache = &phpdoxer.TypeMixed{}
		return r.cache, r.cacheOk
	}

	rh, err := r.returnsHint()
	if err == nil {
		r.cache = rh
		r.cacheOk = true
		return r.cache, r.cacheOk
	}

	if !errors.Is(err, ErrNoReturn) {
		log.Println(fmt.Errorf("[callable.canReturn.Returns]: err getting return hint: %w", err))
	}

	r.cache = &phpdoxer.TypeMixed{}
	return r.cache, r.cacheOk
}

// ReturnsClass resolves and unpacks the raw type returned from Returns into
// the classes it represents.
// See symbol.DocClassNormalize for more.
func (r *canReturn) ReturnsClass(currFqn *fqn.FQN) []*phpdoxer.TypeClassLike {
	if r.clsCache != nil {
		return r.clsCache
	}

	ret, ok := r.Returns()
	if !ok {
		return nil
	}

	fqnt := fqn.NewTraverser()
	r.Root().Walk(fqnt)

	return DocClassNormalize(fqnt, currFqn, ir.GetPosition(r.node), ret)
}

func (r *canReturn) returnsComment() (phpdoxer.Type, error) {
	docNode := r.doxed.FindDoc(FilterDocKind(phpdoxer.KindReturn))
	if docNode != nil {
		return docNode.(*phpdoxer.NodeReturn).Type, nil
	}

	if inhDoc := r.doxed.FindDoc(FilterDocKind(phpdoxer.KindInheritDoc)); inhDoc == nil {
		return nil, ErrNoReturn
	}

	methNode, ok := r.node.(*ir.ClassMethodStmt)
	if !ok {
		return nil, fmt.Errorf(
			"[callable.canReturn.ReturnsComment]: found inheritDoc PHPDoc but the node(%T) is not a method",
			r.node,
		)
	}

	meth := NewMethod(r.rooter, methNode)
	cls, err := NewClassLikeFromMethod(r.Root(), methNode)
	if err != nil {
		return nil, fmt.Errorf(
			"[callable.canReturn.ReturnsComment]: can't get class from method: %w",
			err,
		)
	}

	iter := cls.InheritsIter()
	for inhCls, done, err := iter(); !done; inhCls, done, err = iter() {
		if err != nil {
			log.Println(
				fmt.Errorf("[callable.canReturn.ReturnsComment]: err in inherits iter: %w", err),
			)
			continue
		}

		inhMeth := inhCls.FindMethod(FilterOverwrittenBy(meth))
		if inhMeth == nil {
			continue
		}

		if returnType, ok := inhMeth.Returns(); ok {
			return returnType, nil
		}
	}

	return nil, ErrNoReturn
}

// ReturnsHint parses the return typehint node into a phpdoxer.Type and returns
// it.
func (r *canReturn) returnsHint() (phpdoxer.Type, error) {
	retNode := r.returnsHintNode()
	if retNode == nil {
		return nil, ErrNoReturn
	}

	var isNullable bool
	if nullable, ok := retNode.(*ir.Nullable); ok {
		retNode = nullable.Expr
		isNullable = true
	}

	name, ok := retNode.(*ir.Name)
	if !ok {
		return nil, fmt.Errorf(
			"[callable.canReturn.ReturnsHint]: %T is unexpected for a return type hint, expecting *ir.Name or *ir.Nullable",
			retNode,
		)
	}

	toParse := name.Value
	if isNullable {
		toParse = "null|" + toParse
	}

	t, err := phpdoxer.ParseType(toParse)
	if err != nil {
		return nil, fmt.Errorf(`Error parsing return type hint "%s": %w`, name.Value, err)
	}

	return t, nil
}

func (r *canReturn) returnsHintNode() ir.Node {
	switch tNode := r.node.(type) {
	case *ir.FunctionStmt:
		return tNode.ReturnType
	case *ir.ClassMethodStmt:
		return tNode.ReturnType
	default:
		return nil
	}
}
