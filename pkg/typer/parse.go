package typer

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	maxUint = ^uint(0)
	maxInt  = int(maxUint >> 1)
	minInt  = -maxInt - 1
)

var (
	ErrParse = errors.New("Could not parse type out of PHPDoc")

	precRegex   = regexp.MustCompile(`^(.*)\((.*)\)(.*)$`)
	precRgxBefG = 1
	precRgxInG  = 2
	precRgxAfG  = 3

	intRegex   = regexp.MustCompile(`^int<([\w-]+), ?([\w-]+)>$`)
	intRgxMinG = 1
	intRgxMaxG = 2

	arrRegex   = regexp.MustCompile(`^([nonempty-]*array)<(\w+),? ?(\w*)>$`)
	arrRgxPreG = 1
	arrRgxKeyG = 2
	arrRgxValG = 3

	typeArrRegex    = regexp.MustCompile(`^([\w\\]+)\[\]$`)
	typeArrRgxTypeG = 1

	// Regex from https://www.php.net/manual/en/language.oop5.basic.php,
	// with added \ because we want to match namespaces too.
	identifierRegex = regexp.MustCompile(`^[a-zA-Z_\x80-\xff\\][a-zA-Z0-9_\x80-\xff\\]*$`)

	keyOfRegex     = regexp.MustCompile(`^key-of<([\w\\]+)::(\w+)>$`)
	keyOfRgxClassG = 1
	keyOfRgxConstG = 2

	valueOfRegex     = regexp.MustCompile(`^value-of<([\w\\]+)::(\w+)>$`)
	valueOfRgxClassG = 1
	valueOfRgxConstG = 2

	valueOfEnumRegex    = regexp.MustCompile(`^value-of<([\w\\]+)>$`)
	valueOfEnumRgxEnumG = 1

	constrainedClassStringRegex = regexp.MustCompile(`^class-string<([\w\\]+)>$`)
	constrClsStrNameG           = 1

	iterableRegex = regexp.MustCompile(
		`^([a-zA-Z_\x80-\xff\\][a-zA-Z0-9_\x80-\xff\\]*)<(\w+),? ?(\w*)>$`,
	)
	iterRgxNameG = 1
	iterRgxKeyG  = 2
	iterRgxValG  = 3
)

// TODO: type aliasses
// TODO: generics
// TODO: conditional types
// TODO: array shapes
// TODO: literals and constants
// TODO: global constants
// TODO: integer masks
// TODO: complex callable

// TODO: pass in ir.Root and class scope, use class scope with $this and static.
func Parse(value string) (Type, error) {
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
	case "string":
		return &TypeString{}, nil
	case "class-string":
		return &TypeString{Constraint: StringConstraintClass}, nil
	case "callable-string":
		return &TypeString{Constraint: StringConstraintCallable}, nil
	case "numeric-string":
		return &TypeString{Constraint: StringConstraintNumeric}, nil
	case "non-empty-string":
		return &TypeString{Constraint: StringConstraintNonEmpty}, nil
	case "literal-string":
		return &TypeString{Constraint: stringConstraintLiteral}, nil
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
		return &TypeClassLike{Name: value}, nil
	case "never", "never-return", "never-returns", "no-return":
		return &TypeNever{}, nil
	}

	if match, rType, rErr := parseComplexInt(value); match {
		return rType, rErr
	}

	if match, rType, rErr := parseComplexArray(value); match {
		return rType, rErr
	}

	if match, rType, rErr := parseComplexTypeArray(value); match {
		return rType, rErr
	}

	if match, rType, rErr := parseIdentifier(value); match {
		return rType, rErr
	}

	if match, rType, rErr := parseKeyOf(value); match {
		return rType, rErr
	}

	if match, rType, rErr := parseValueOf(value); match {
		return rType, rErr
	}

	if match, rType, rErr := parseValueOfEnum(value); match {
		return rType, rErr
	}

	if match, rType, rErr := parseConstrainedClassString(value); match {
		return rType, rErr
	}

	if match, rType, rErr := parseIterable(value); match {
		return rType, rErr
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
			bef, err = Parse(prec[precRgxBefG][0 : len(prec[precRgxBefG])-1])
		}

		if prec[precRgxAfG] != "" {
			symAf = prec[precRgxAfG][:1]
			af, err = Parse(prec[precRgxAfG][1:])
		}

		if err != nil {
			return nil, err
		}

		inner, err := Parse(prec[precRgxInG])
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
		left, err := Parse(value[:ui])
		if err != nil {
			return nil, err
		}

		right, err := Parse(value[ui+1:])
		if err != nil {
			return nil, err
		}

		return &TypeUnion{
			Left:  left,
			Right: right,
		}, nil
	}

	if ii != -1 && (ii < ui || ui == -1) {
		left, err := Parse(value[:ii])
		if err != nil {
			return nil, err
		}

		right, err := Parse(value[ii+1:])
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

func ParseUnion(value []string) (Type, error) {
	if len(value) < 2 {
		return nil, fmt.Errorf("Union needs at least 2 parts: %w", ErrParse)
	}

	ret := &TypeUnion{}
	curr := ret
	for i, part := range value {
		parsed, err := Parse(part)
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

func parseComplexInt(value string) (bool, Type, error) {
	intMatch := intRegex.FindStringSubmatch(value)
	if len(intMatch) < intRgxMaxG+1 {
		return false, nil, nil
	}

	min := string(intMatch[intRgxMinG])
	minAsInt, err := strconv.Atoi(min)
	if err != nil {
		minAsInt = minInt
		if min != "min" {
			return true, nil, fmt.Errorf(
				"Unexpected minimum for integer type %s, got %s but want either a number or the literal 'min': %w",
				value,
				min,
				ErrParse,
			)
		}
	}

	max := string(intMatch[intRgxMaxG])
	maxAsInt, err := strconv.Atoi(max)
	if err != nil {
		maxAsInt = maxInt
		if max != "max" {
			return true, nil, fmt.Errorf(
				"Unexpected maximum for integer type %s, got %s but want either a number or the literal 'max': %w",
				value,
				max,
				ErrParse,
			)
		}
	}

	if minAsInt > maxAsInt {
		return true, nil, fmt.Errorf(
			"Unexpected min/max range for integer %s type, the given minimum %s is larger than the given maximum %s: %w",
			value,
			min,
			max,
			ErrParse,
		)
	}

	return true, &TypeInt{Min: min, Max: max}, nil
}

func parseComplexArray(value string) (bool, Type, error) {
	arrMatch := arrRegex.FindStringSubmatch(value)
	if len(arrMatch) < arrRgxValG+1 {
		return false, nil, nil
	}

	if arrMatch[arrRgxPreG] != "array" && arrMatch[arrRgxPreG] != "non-empty-array" {
		return false, nil, nil
	}

	nonEmpty := arrMatch[arrRgxPreG] == "non-empty-array"

	// No value means the key is the value.
	if arrMatch[arrRgxValG] == "" {
		itemType, err := Parse(arrMatch[arrRgxKeyG])
		if err != nil {
			return true, nil, fmt.Errorf("Error parsing array item type for %s: %w", value, err)
		}

		return true, &TypeArray{ItemType: itemType, NonEmpty: nonEmpty}, nil
	}

	keyType, err := Parse(arrMatch[arrRgxKeyG])
	if err != nil {
		return true, nil, fmt.Errorf("Error parsing array key type for %s: %w", value, err)
	}

	itemType, err := Parse(arrMatch[arrRgxValG])
	if err != nil {
		return true, nil, fmt.Errorf("Error parsing array value type for %s: %w", value, err)
	}

	return true, &TypeArray{KeyType: keyType, ItemType: itemType, NonEmpty: nonEmpty}, nil
}

func parseComplexTypeArray(value string) (bool, Type, error) {
	arrMatch := typeArrRegex.FindStringSubmatch(value)
	if len(arrMatch) != typeArrRgxTypeG+1 {
		return false, nil, nil
	}

	itemType, err := Parse(arrMatch[typeArrRgxTypeG])
	if err != nil {
		return true, nil, fmt.Errorf("Error parsing array value type for %s: %w", value, err)
	}

	return true, &TypeArray{ItemType: itemType}, nil
}

func parseIdentifier(value string) (bool, Type, error) {
	identifierMatch := identifierRegex.MatchString(value)
	if !identifierMatch {
		return false, nil, nil
	}

	fullyQualified := strings.HasPrefix(value, `\`)

	return true, &TypeClassLike{Name: value, FullyQualified: fullyQualified}, nil
}

func parseKeyOf(value string) (bool, Type, error) {
	keyOfMatch := keyOfRegex.FindStringSubmatch(value)
	if len(keyOfMatch) < keyOfRgxConstG+1 {
		return false, nil, nil
	}

	return true, &TypeKeyOf{
		Class: &TypeClassLike{Name: keyOfMatch[keyOfRgxClassG]},
		Const: keyOfMatch[keyOfRgxConstG],
	}, nil
}

func parseValueOf(value string) (bool, Type, error) {
	valueOfMatch := valueOfRegex.FindStringSubmatch(value)
	if len(valueOfMatch) < valueOfRgxConstG+1 {
		return false, nil, nil
	}

	return true, &TypeValueOf{
		Class: &TypeClassLike{Name: valueOfMatch[valueOfRgxClassG]},
		Const: valueOfMatch[valueOfRgxConstG],
	}, nil
}

func parseValueOfEnum(value string) (bool, Type, error) {
	valueOfEnumMatch := valueOfEnumRegex.FindStringSubmatch(value)
	if len(valueOfEnumMatch) < valueOfEnumRgxEnumG+1 {
		return false, nil, nil
	}

	return true, &TypeValueOf{
		Class:  &TypeClassLike{Name: valueOfEnumMatch[valueOfEnumRgxEnumG]},
		IsEnum: true,
	}, nil
}

// NOTE: iterables have some different rules depending on implementation of interfaces,
// NOTE: see: https://phpstan.org/writing-php-code/phpdoc-types#iterables for details.
// TODO: this is currently not supported and is going to be hard to support.
func parseIterable(value string) (bool, Type, error) {
	iterableMatch := iterableRegex.FindStringSubmatch(value)
	if len(iterableMatch) < iterRgxValG+1 {
		return false, nil, nil
	}

	var iterType Type
	iterTypeRaw := iterableMatch[iterRgxNameG]
	if iterTypeRaw != "iterable" {
		res, err := Parse(iterTypeRaw)
		if err != nil {
			return true, nil, fmt.Errorf(
				"Error parsing iterable type %s of type %s: %w",
				iterTypeRaw,
				value,
				err,
			)
		}

		iterType = res
	}

	keyTypeRaw := iterableMatch[iterRgxKeyG]
	keyType, err := Parse(keyTypeRaw)
	if err != nil {
		return true, nil, fmt.Errorf(
			"Error parsing iterable key type %s of type %s: %w",
			keyTypeRaw,
			value,
			err,
		)
	}

	valTypeRaw := iterableMatch[iterRgxValG]
	if len(valTypeRaw) == 0 {
		return true, &TypeIterable{
			IterableType: iterType,
			ItemType:     keyType,
		}, nil
	}

	valType, err := Parse(valTypeRaw)
	if err != nil {
		return true, nil, fmt.Errorf(
			"Error parsing iterable value type %s of type %s: %w",
			valTypeRaw,
			value,
			err,
		)
	}

	return true, &TypeIterable{
		IterableType: iterType,
		KeyType:      keyType,
		ItemType:     valType,
	}, nil
}

func parseConstrainedClassString(value string) (bool, Type, error) {
	match := constrainedClassStringRegex.FindStringSubmatch(value)
	if len(match) < constrClsStrNameG+1 {
		return false, nil, nil
	}

	fullyQualified := match[constrClsStrNameG][0:1] == `\`

	return true, &TypeString{
		Constraint:  StringConstraintClass,
		GenericOver: &TypeClassLike{Name: match[constrClsStrNameG], FullyQualified: fullyQualified},
	}, nil
}
