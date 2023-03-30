package symbol

import (
	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/visitor"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/functional"
	"github.com/laytan/elephp/pkg/nodeident"
	"github.com/laytan/elephp/pkg/nodescopes"
)

type inheritType int

const (
	inheritTypeUses inheritType = iota
	inheritTypeExtends
	inheritTypeImplements
)

type (
	InheritsIterFunc func() (clsLike *ClassLike, done bool, newClsLikeErr error)
)

// InheritsIter returns a generator function that generates the
// classes/interfaces/traits that this class/interface/trait inherits from.
//
// The classes are generated in the same order that PHP evaluates them, so,
// if you have a method in the first generated class, it is overwriting any
// other generated classes that define that method.
//
// This order is:
//  1. Current class/trait/interface
//  2. Any used traits
//  3. If the class extends another
//     3a. Extended class
//     3b. Extended class's traits
//     3c. If the class extends another -> back to 3, otherwise 4.
//  4. Any interface implementations
//
// The generator returns true for done when there are no classes left in the chain.
// If there was an error generating the next class, the newClsLikeErr argument will be set.
func (c *ClassLike) InheritsIter() InheritsIterFunc {
	slice := c.inheritor.Uses()
	sliceIndex := 0
	sliceType := inheritTypeUses

	var iter InheritsIterFunc

	return func() (*ClassLike, bool, error) {
		if iter != nil {
			clsLike, _, err := iter()
			if err != nil {
				return nil, false, err
			}

			if clsLike != nil {
				return clsLike, false, nil
			}
		}

		// As long as we don't have values in the slice, go to the next slice type
		// If we are at implements and it has no values, we are done.
		for ok := len(slice) > sliceIndex; !ok; ok = len(slice) > sliceIndex {
			if sliceType == inheritTypeUses {
				// A class can only extend one class, for easier logic, add it
				// to a temporary slice.
				slice = []*ast.Name{}
				if c.inheritor.Extends() != nil {
					slice = append(slice, c.inheritor.Extends())
				}

				sliceType = inheritTypeExtends
				sliceIndex = 0
				continue
			}

			if sliceType == inheritTypeExtends {
				slice = c.inheritor.Implements()

				sliceType = inheritTypeImplements
				sliceIndex = 0
				continue
			}

			if sliceType == inheritTypeImplements {
				return nil, true, nil
			}
		}

		newC, err := NewClassLikeFromName(c.Root(), slice[sliceIndex])
		sliceIndex++

		// If there was an error New'ing the class, return it, we still say
		// false for done because we can still generate the next values.
		if err != nil {
			return nil, false, err
		}

		// Recursively set the iter, this will make sure we go into traits,
		// then into extends, and then into implements, recursively.
		iter = newC.InheritsIter()

		return newC, false, nil
	}
}

type inheritsTraverser struct {
	visitor.Null
	target *fqn.FQN

	uses       []*ast.Name
	extends    *ast.Name
	implements []*ast.Name

	currNamespace string
}

func newInheritsTraverser(target *fqn.FQN) *inheritsTraverser {
	return &inheritsTraverser{
		target: target,
	}
}

func (t *inheritsTraverser) EnterNode(node ast.Vertex) bool {
	switch node.(type) {
	case *ast.Root:
		return true

	case *ast.StmtNamespace:
		t.currNamespace = nodeident.Get(node)
		return t.currNamespace == t.target.Namespace()

	default:
		// Don't go into classes that don't match target.
		if nodescopes.IsClassLike(node.GetType()) {
			if nodeident.Get(node) != t.target.Name() {
				return false
			}
		}

		switch typedNode := node.(type) {
		case *ast.StmtTraitUse:
			t.uses = append(t.uses, functional.MapFilter(typedNode.Traits, nodeToName)...)
		case *ast.StmtClass:
			t.implements = append(t.implements, functional.MapFilter(typedNode.Implements, nodeToName)...)
			if typedNode.Extends != nil {
				t.extends = typedNode.Extends.(*ast.Name)
			}
		case *ast.StmtInterface:
			t.implements = append(t.implements, functional.MapFilter(typedNode.Extends, nodeToName)...)
		}

		// Don't go into scopes that are not necessary.
		if nodescopes.IsNonClassLikeScope(node.GetType()) {
			return false
		}

		return true
	}
}
