package doxcontext

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/php-parser/pkg/position"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/phpdoxer"
)

// ApplyContext applies any context we have about the code (which phpdoxer
// does not have).
//
// Normalizations:
// - "static" -> current class
// - "$this" -> current class
// - "self" -> current class
// - A class in all CAPS -> qualified class
// - Precedence -> recursively unpacked
// - Union -> recursively unpacked
// - Intersection -> recursively unpacked
// - Arrays -> their value type
//
// Precedence, union & intersection can result in multiple classes, so a slice is returned.
// TODO: return fqn.FQN's instead of typeclasslikes.
func ApplyContext(
	fqnt *fqn.Traverser,
	currFqn *fqn.FQN,
	currPos *position.Position,
	doc phpdoxer.Type,
) (res []*phpdoxer.TypeClassLike) {
	switch typed := doc.(type) {
	case *phpdoxer.TypeClassLike:
		switch typed.Name {
		case "self", "static", "$this":
			return []*phpdoxer.TypeClassLike{{
				Name:           currFqn.String(),
				FullyQualified: true,
			}}
		}

		return []*phpdoxer.TypeClassLike{{
			Name: fqnt.ResultFor(&ir.Name{
				Position: currPos,
				Value:    typed.Name,
			}).String(),
			FullyQualified: true,
		}}
	case *phpdoxer.TypeConstant:
		if typed.Class != nil {
			return nil
		}

		// In case the user created a class in all CAPS, we catch here that
		// it is not a constant, but a class.
		res := fqnt.ResultFor(&ir.Name{Position: currPos, Value: typed.Const})
		if res != nil {
			return []*phpdoxer.TypeClassLike{{Name: res.String(), FullyQualified: true}}
		}

	case *phpdoxer.TypePrecedence:
		return ApplyContext(fqnt, currFqn, currPos, typed.Type)
	case *phpdoxer.TypeUnion:
		res = append(res, ApplyContext(fqnt, currFqn, currPos, typed.Left)...)
		res = append(res, ApplyContext(fqnt, currFqn, currPos, typed.Right)...)
		return res
	case *phpdoxer.TypeIntersection:
		res = append(res, ApplyContext(fqnt, currFqn, currPos, typed.Left)...)
		res = append(res, ApplyContext(fqnt, currFqn, currPos, typed.Right)...)
		return res
	case *phpdoxer.TypeArray:
		return ApplyContext(fqnt, currFqn, currPos, typed.ItemType)
	}

	return nil
}
