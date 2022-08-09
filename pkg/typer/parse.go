package typer

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	precRgxBefG = 1
	precRgxInG  = 2
	precRgxAfG  = 3
)

var (
	ErrParse  = errors.New("Could not parse type out of PHPDoc")
	precRegex = regexp.MustCompile(`^(.*)\((.*)\)(.*)$`)
)

// TODO: type aliasses
// TODO: generics
// TODO: conditional types
// TODO: array shapes
// TODO: literals and constants
// TODO: global constants
// TODO: integer masks
// TODO: any parameterized types (int<0, 2>, callable(x)x, array<int>, Type[], etc.)

// TODO: pass in ir.Root and class scope, use class scope with $this and static.
func Parse(class *FQN, value string) (Type, error) {
	if len(value) == 0 {
		return nil, fmt.Errorf("Empty phpdoc given: %w", ErrParse)
	}

	switch value {
	case "mixed":
		return &TypeMixed{}, nil
	case "null":
		return &TypeNull{}, nil
	case "bool", "boolean":
		return &TypeBool{Accepts: BoolAcceptsAll}, nil
	case "true":
		return &TypeBool{Accepts: BoolAcceptsTrue}, nil
	case "false":
		return &TypeBool{Accepts: BoolAcceptsFalse}, nil
	case "float", "double":
		return &TypeFloat{}, nil
	case "int", "integer":
		return &TypeInt{}, nil
	case "positive-int":
		return &TypeInt{HasPositiveConstraint: true}, nil
	case "negative-int":
		return &TypeInt{HasNegativeConstraint: true}, nil
	case "string",
		"class-string",
		"callable-string",
		"numeric-string",
		"non-empty-string",
		"literal-string":
		// TODO: show the constraints in the struct somehow, maybe a constraint field?
		return &TypeString{}, nil
	case "object":
		return &TypeObject{}, nil
	case "scalar":
		return &TypeScalar{}, nil
	case "iterable":
		return &TypeIterable{}, nil
	case "array":
		return &TypeArray{}, nil
	case "non-empty-array":
		return &TypeArray{NonEmpty: true}, nil
	case "callable":
		return &TypeCallable{Return: &TypeVoid{}}, nil
	case "void":
		return &TypeVoid{}, nil
	case "array-key":
		return &TypeArrayKey{}, nil
	case "resource":
		return &TypeResource{}, nil
	case "$this", "static":
		// NOTE: this is not exactly correct, but we would need to know what class
		// NOTE: the method is called from to get this, which we don't.
		return &TypeClassLike{FQN: class}, nil
	case "never", "never-return", "never-returns", "no-return":
		return &TypeNever{}, nil
	}

	prec := precRegex.FindStringSubmatch(value)
	if len(prec) > 0 {
		var symBef string
		var bef Type
		var symAf string
		var af Type
		var err error
		if prec[precRgxBefG] != "" {
			symBef = prec[precRgxBefG][len(prec[precRgxBefG])-1:]
			bef, err = Parse(
				class,
				prec[precRgxBefG][0:len(prec[precRgxBefG])-1],
			)
		}

		if prec[precRgxAfG] != "" {
			symAf = prec[precRgxAfG][:1]
			af, err = Parse(class, prec[precRgxAfG][1:])
		}

		if err != nil {
			return nil, err
		}

		inner, err := Parse(class, prec[precRgxInG])
		if err != nil {
			return nil, err
		}

		var right Type
		right = &TypePrecedence{Type: inner}
		if af != nil {
			switch symAf {
			case "|":
				right = &TypeUnion{Left: right, Right: af}
			case "&":
				right = &TypeIntersection{Left: right, Right: af}
			default:
				return nil, fmt.Errorf(
					"Unexpected type before or after precedence (want | or &) got %s: %w",
					symBef,
					ErrParse,
				)
			}
		}

		if bef != nil {
			switch symBef {
			case "|":
				return &TypeUnion{Left: bef, Right: right}, nil
			case "&":
				return &TypeIntersection{Left: bef, Right: right}, nil
			default:
				return nil, fmt.Errorf(
					"Unexpected type before or after precedence (want | or &) got %s: %w",
					symBef,
					ErrParse,
				)
			}
		}

		return right, nil
	}

	ui := strings.Index(value, "|")
	ii := strings.Index(value, "&")

	if ui != -1 && (ui < ii || ii == -1) {
		left, err := Parse(class, value[:ui])
		if err != nil {
			return nil, err
		}

		right, err := Parse(class, value[ui+1:])
		if err != nil {
			return nil, err
		}

		return &TypeUnion{
			Left:  left,
			Right: right,
		}, nil
	}

	if ii != -1 && (ii < ui || ui == -1) {
		left, err := Parse(class, value[:ii])
		if err != nil {
			return nil, err
		}

		right, err := Parse(class, value[ii+1:])
		if err != nil {
			return nil, err
		}

		return &TypeIntersection{
			Left:  left,
			Right: right,
		}, nil
	}

	return nil, fmt.Errorf("Unsupported type %s: %w", value, ErrParse)
}

func ParseUnion(class *FQN, value []string) (Type, error) {
	if len(value) < 2 {
		return nil, fmt.Errorf("Union needs at least 2 parts: %w", ErrParse)
	}

	ret := &TypeUnion{}
	curr := ret
	for i, part := range value {
		parsed, err := Parse(class, part)
		if err != nil {
			return nil, err
		}

		if curr.Left == nil {
			curr.Left = parsed
			continue
		}

		if curr.Right == nil {
			if i == len(value)-1 {
				curr.Right = parsed
				continue
			}

			newU := &TypeUnion{
				Left: parsed,
			}

			curr.Right = newU
			curr = newU
		}
	}

	return ret, nil
}
