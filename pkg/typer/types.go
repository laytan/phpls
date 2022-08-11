package typer

import (
	"fmt"
	"strings"
)

// TODO: support phpstan's constants https://phpstan.org/writing-php-code/phpdoc-types#literals-and-constants.
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
	KindKeyOf
	KindValueOf
	KindArrayShape
	KindArrayShapeValue
	KindStringLiteral
	KindFloatLiteral
	KindIntLiteral
	KindConstant
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

type TypeIterable struct {
	// Is nil when the doc had 'iterable<x,y>' instead of 'Type<x, y>'.
	IterableType Type
	// Is nil when it was created like this: 'iterable<x>'
	KeyType Type
	// Is nil when it was created like this: 'iterable'
	ItemType Type
}

func (t *TypeIterable) String() string {
	if t.IterableType == nil && t.KeyType == nil && t.ItemType == nil {
		return "iterable"
	}

	if t.KeyType == nil && t.IterableType == nil {
		return fmt.Sprintf("iterable<%s>", t.ItemType.String())
	}

	if t.KeyType == nil {
		return fmt.Sprintf("%s<%s>", t.IterableType.String(), t.ItemType.String())
	}

	if t.IterableType == nil {
		return fmt.Sprintf("iterable<%s, %s>", t.KeyType.String(), t.ItemType.String())
	}

	return fmt.Sprintf(
		"%s<%s, %s>",
		t.IterableType.String(),
		t.KeyType.String(),
		t.ItemType.String(),
	)
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

type StringConstraint uint

const (
	StringConstraintNone StringConstraint = iota
	StringConstraintClass
	StringConstraintCallable
	StringConstraintNumeric
	StringConstraintNonEmpty
	stringConstraintLiteral
)

type TypeString struct {
	Constraint StringConstraint

	// Can only be set when Constraint is ConstraintClass.
	GenericOver *TypeClassLike
}

func (t *TypeString) String() string {
	switch t.Constraint {
	case StringConstraintClass:
		if t.GenericOver != nil {
			return fmt.Sprintf("class-string<%s>", t.GenericOver.String())
		}

		return "class-string"
	case StringConstraintCallable:
		return "callable-string"
	case StringConstraintNumeric:
		return "numeric-string"
	case StringConstraintNonEmpty:
		return "non-empty-string"
	case stringConstraintLiteral:
		return "literal-string"
	default:
		return "string"
	}
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

type TypeKeyOf struct {
	Class *TypeClassLike
	Const string
}

func (t *TypeKeyOf) String() string {
	return fmt.Sprintf("key-of<%s::%s>", t.Class.String(), t.Const)
}

func (t *TypeKeyOf) Kind() TypeKind {
	return KindKeyOf
}

type TypeValueOf struct {
	Class *TypeClassLike

	// If IsEnum is true, Const is an empty string.
	Const  string
	IsEnum bool
}

func (t *TypeValueOf) String() string {
	if t.IsEnum {
		return fmt.Sprintf("value-of<%s>", t.Class.String())
	}

	return fmt.Sprintf("value-of<%s::%s>", t.Class.String(), t.Const)
}

func (t *TypeValueOf) Kind() TypeKind {
	return KindValueOf
}

type TypeArrayShapeValue struct {
	// Can be an int disguised as a string
	Key      string
	Type     Type
	Optional bool
}

func (t *TypeArrayShapeValue) String() string {
	key := t.Key
	if key != "" {
		if t.Optional {
			key += "?"
		}

		key += ": "
	}

	return fmt.Sprintf("%s%s", key, t.Type.String())
}

func (t *TypeArrayShapeValue) Kind() TypeKind {
	return KindArrayShapeValue
}

type TypeArrayShape struct {
	Values []*TypeArrayShapeValue
}

func (t *TypeArrayShape) String() string {
	values := make([]string, len(t.Values))
	for i, v := range t.Values {
		values[i] = v.String()
	}

	return fmt.Sprintf("array{%s}", strings.Join(values, ", "))
}

func (t *TypeArrayShape) Kind() TypeKind {
	return KindArrayShape
}

type TypeStringLiteral struct {
	Value string
}

func (t *TypeStringLiteral) String() string {
	return fmt.Sprintf("'%s'", t.Value)
}

func (t *TypeStringLiteral) Kind() TypeKind {
	return KindStringLiteral
}

type TypeIntLiteral struct {
	Value int
}

func (t *TypeIntLiteral) String() string {
	return fmt.Sprintf("%d", t.Value)
}

func (t *TypeIntLiteral) Kind() TypeKind {
	return KindIntLiteral
}

type TypeFloatLiteral struct {
	Value float64
}

func (t *TypeFloatLiteral) String() string {
	return fmt.Sprintf("%f", t.Value)
}

func (t *TypeFloatLiteral) Kind() TypeKind {
	return KindFloatLiteral
}

type TypeConstant struct {
	Class *TypeClassLike
	Const string
}

func (t *TypeConstant) String() string {
	classStr := ""
	if t.Class != nil {
		classStr = t.Class.String() + "::"
	}

	return classStr + t.Const
}

func (t *TypeConstant) Kind() TypeKind {
	return KindConstant
}
