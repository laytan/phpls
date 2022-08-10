package typer

import (
	"fmt"
	"strings"
)

// TODO: support phpstan's literals and constants https://phpstan.org/writing-php-code/phpdoc-types#literals-and-constants.
// TODO: support phpstan's generics: https://phpstan.org/writing-php-code/phpdoc-types#generics.
// TODO: support phpstan's conditional return types: https://phpstan.org/writing-php-code/phpdoc-types#conditional-return-types.

type TypeKind uint

const (
	KindMixed TypeKind = iota
	KindNull
	KindClassLike
	KindArray
	KindIterable
	KindCallale
	KindBool
	KindFloat
	KindInt
	KindString
	KindObject
	KindNever
	KindScalar
	KindVoid
	KindArrayKey
	KindResource
	KindPrecedence
	KindUnion
	KindIntersection
)

type Type interface {
	String() string
	Kind() TypeKind
}

type TypeMixed struct{}

func (t *TypeMixed) String() string {
	return "mixed"
}

func (t *TypeMixed) Kind() TypeKind {
	return KindMixed
}

type TypeNull struct{}

func (t *TypeNull) String() string {
	return "null"
}

func (t *TypeNull) Kind() TypeKind {
	return KindNull
}

type TypeClassLike struct {
	Name           string
	FullyQualified bool
}

func (t *TypeClassLike) String() string {
	return t.Name
}

func (t *TypeClassLike) Kind() TypeKind {
	return KindClassLike
}

// TODO: support phpstan's array shapes: https://phpstan.org/writing-php-code/phpdoc-types#array-shapes.
type TypeArray struct {
	KeyType  Type
	ItemType Type
	NonEmpty bool
}

func (t *TypeArray) String() string {
	nonEmptyPrefix := ""
	if t.NonEmpty {
		nonEmptyPrefix = "non-empty-"
	}

	if t.ItemType == nil && t.KeyType == nil {
		return fmt.Sprintf("%sarray", nonEmptyPrefix)
	}

	if t.KeyType == nil {
		return fmt.Sprintf("%sarray<%s>", nonEmptyPrefix, t.ItemType.String())
	}

	return fmt.Sprintf("%sarray<%s, %s>", nonEmptyPrefix, t.KeyType.String(), t.ItemType.String())
}

func (t *TypeArray) Kind() TypeKind {
	return KindArray
}

// TODO: support phpstan's iterable signatures: array signatures: https://phpstan.org/writing-php-code/phpdoc-types#iterables.
type TypeIterable struct {
	ItemType Type
}

func (t *TypeIterable) String() string {
	return fmt.Sprintf("iterable<%s>", t.ItemType.String())
}

func (t *TypeIterable) Kind() TypeKind {
	return KindIterable
}

type CallableParameter struct {
	IsOptional bool
	Type       Type
}

func (c *CallableParameter) String() string {
	res := c.Type.String()
	if c.IsOptional {
		res += "="
	}

	return res
}

// TODO: support variadic parameters.
type TypeCallable struct {
	Parameters []CallableParameter
	Return     Type
}

func (t *TypeCallable) String() string {
	params := make([]string, len(t.Parameters))
	for i, param := range t.Parameters {
		params[i] = param.String()
	}

	return fmt.Sprintf("callable(%s): %s", strings.Join(params, ", "), t.Return.String())
}

func (t *TypeCallable) Kind() TypeKind {
	return KindCallale
}

type BoolAccepts uint

const (
	BoolAcceptsFalse BoolAccepts = iota
	BoolAcceptsTrue
	BoolAcceptsAll
)

type TypeBool struct {
	Accepts BoolAccepts
}

func (t *TypeBool) String() string {
	switch t.Accepts {
	case BoolAcceptsFalse:
		return "false"
	case BoolAcceptsTrue:
		return "true"
	default:
		return "bool"
	}
}

func (t *TypeBool) Kind() TypeKind {
	return KindBool
}

type TypeFloat struct{}

func (t *TypeFloat) String() string {
	return "float"
}

func (t *TypeFloat) Kind() TypeKind {
	return KindFloat
}

type TypeInt struct {
	// Either 'min' or an int.
	Min string
	// Either 'max' or an int.
	Max string

	// Whether it is a 'positive-int'.
	HasPositiveConstraint bool
	// Whether it is a 'negative-int'.
	HasNegativeConstraint bool
}

func (t *TypeInt) String() string {
	if t.HasNegativeConstraint {
		return "negative-int"
	}

	if t.HasPositiveConstraint {
		return "positive-int"
	}

	if len(t.Min) > 0 || len(t.Max) > 0 {
		min := t.Min
		if min == "" {
			min = "min"
		}

		max := t.Max
		if max == "" {
			max = "max"
		}

		return fmt.Sprintf("int<%s, %s>", min, max)
	}

	return "int"
}

func (t *TypeInt) Kind() TypeKind {
	return KindInt
}

// TODO: support phpstan's advanced string types: https://phpstan.org/writing-php-code/phpdoc-types#other-advanced-string-types.
// TODO: support phpstan's class-string: https://phpstan.org/writing-php-code/phpdoc-types#class-string.
type TypeString struct{}

func (t *TypeString) String() string {
	return "string"
}

func (t *TypeString) Kind() TypeKind {
	return KindString
}

type TypeObject struct{}

func (t *TypeObject) String() string {
	return "object"
}

func (t *TypeObject) Kind() TypeKind {
	return KindObject
}

type TypeNever struct{}

func (t *TypeNever) String() string {
	return "never"
}

func (t *TypeNever) Kind() TypeKind {
	return KindNever
}

// Scalar is an int, float, bool or string.
type TypeScalar struct{}

func (t *TypeScalar) String() string {
	return "scalar"
}

func (t *TypeScalar) Kind() TypeKind {
	return KindScalar
}

type TypeVoid struct{}

func (t *TypeVoid) String() string {
	return "void"
}

func (t *TypeVoid) Kind() TypeKind {
	return KindVoid
}

// TypeArrayKey is a hashable value (usable as an array key).
type TypeArrayKey struct{}

func (t *TypeArrayKey) String() string {
	return "array-key"
}

func (t *TypeArrayKey) Kind() TypeKind {
	return KindArrayKey
}

type TypeResource struct{}

func (t *TypeResource) String() string {
	return "resource"
}

func (t *TypeResource) Kind() TypeKind {
	return KindResource
}

type TypePrecedence struct {
	Type Type
}

func (t *TypePrecedence) String() string {
	return fmt.Sprintf("(%s)", t.Type.String())
}

func (t *TypePrecedence) Kind() TypeKind {
	return KindPrecedence
}

type TypeUnion struct {
	Left  Type
	Right Type
}

func (t *TypeUnion) String() string {
	return fmt.Sprintf("%s|%s", t.Left.String(), t.Right.String())
}

func (t *TypeUnion) Kind() TypeKind {
	return KindUnion
}

type TypeIntersection struct {
	Left  Type
	Right Type
}

func (t *TypeIntersection) String() string {
	return fmt.Sprintf("%s&%s", t.Left.String(), t.Right.String())
}

func (t *TypeIntersection) Kind() TypeKind {
	return KindIntersection
}
