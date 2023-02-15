//nolint:unsafenil // The first arg being false means nil will be returned here.
package phpdoxer

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
	ErrParse   = errors.New("Could not parse type out of PHPDoc")
	ErrUnknown = errors.New("No parsing rules matched the given PHPDoc")
	ErrEmpty   = errors.New("Given PHPDoc string is empty")

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
		`^iterable<(\w+),? ?(\w*)>$`,
	)
	iterRgxKeyG = 1
	iterRgxValG = 2

	arrShapeRegexStep1 = regexp.MustCompile(`^array{(.*)}$`)
	arrShapeRgxValsG   = 1

	// TODO: this will fail for: 'foo\'s' or "foo's" or "foo\"s" etc.
	strLiteralRegex = regexp.MustCompile(`^['"]([^"']*)['"]$`)
	strLitRgxValG   = 1

	classConstRegex = regexp.MustCompile(`^([\w\\]+)::([\*\w]+)$`)
	clsCnstRgxClsG  = 1
	clsCnstRgxCnstG = 2

	globalConstRegex = regexp.MustCompile(`^[_A-Z]*$`)

	callableRegex = regexp.MustCompile(`^callable\((.*)\): ?(.*)$`)
	clbleRgxPrmsG = 1
	clbleRgxRtrnG = 2

	intMaskRegex  = regexp.MustCompile(`^int-mask<([\d,\s]+)>$`)
	intMskRgxVlsG = 1

	intMaskOfRegex  = regexp.MustCompile(`^int-mask-of<(.*)>$`)
	intMskOfRgxTypG = 1

	conditionalReturnRegex = regexp.MustCompile(
		`^\( *(\$?[\w]+) *is *(\S+) *\? *(\S+) *: *(\S+) *\)$`,
	)
	cndRetRgxPrmG   = 1
	cndRetRgxCheckG = 2
	cndRetRgxTrueG  = 3
	cndRetRgxFalseG = 4

	genericTemplateRegex = regexp.MustCompile(`^(\$?[\w]+) *of *([\w\\]+)$`)
	genTemRgxNameG       = 1
	genTemRgxOfG         = 2

	genericClassRegex = regexp.MustCompile(`^([a-zA-Z_\x80-\xff\\][a-zA-Z0-9_\x80-\xff\\]*)<(.*)>$`)
	genClsRgxNameG    = 1
	genClsRgxGenOverG = 2
)

// TODO: type aliasses

// NOTE: const and classlike types are basically the same format, to parse those
// NOTE: correctly we need context of the actual project, which we don't have here.
// NOTE: for elephp, the internal package 'doxcontext' is there to
// NOTE: apply context and transform these cases.
func ParseType(value string) (Type, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return nil, ErrEmpty
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
		return &TypeString{Constraint: StringConstraintLiteral}, nil
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
		return &TypeCallable{}, nil
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

	if match, rType, rErr := parseClassConst(value); match {
		return rType, rErr
	}

	if match, rType, rErr := parseGlobalConst(value); match {
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

	if match, rType, rErr := parseArrayShape(value); match {
		return rType, rErr
	}

	if match, rType, rErr := parseIntLiteral(value); match {
		return rType, rErr
	}

	if match, rType, rErr := parseFloatLiteral(value); match {
		return rType, rErr
	}

	if match, rType, rErr := parseStringLiteral(value); match {
		return rType, rErr
	}

	if match, rType, rErr := parseCallable(value); match {
		return rType, rErr
	}

	if match, rType, rErr := parseIntMask(value); match {
		return rType, rErr
	}

	if match, rType, rErr := parseIntMaskOf(value); match {
		return rType, rErr
	}

	if match, rType, rErr := parseConditionalReturn(value); match {
		return rType, rErr
	}

	if match, rType, rErr := parseGenericTemplate(value); match {
		return rType, rErr
	}

	if match, rType, rErr := parseGenericClass(value); match {
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
			bef, err = ParseType(prec[precRgxBefG][0 : len(prec[precRgxBefG])-1])
		}

		if prec[precRgxAfG] != "" {
			symAf = prec[precRgxAfG][:1]
			af, err = ParseType(prec[precRgxAfG][1:])
		}

		if err != nil {
			return nil, err
		}

		inner, err := ParseType(prec[precRgxInG])
		if err != nil {
			return nil, err
		}

		var right Type
		right = &TypePrecedence{Type: inner}
		if af != nil {
			switch symAf {
			case unionSymbol:
				right = &TypeUnion{Left: right, Right: af}
			case intersectionSymbol:
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
			case unionSymbol:
				return &TypeUnion{Left: bef, Right: right}, nil
			case intersectionSymbol:
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

	ui := strings.Index(value, unionSymbol)
	ii := strings.Index(value, intersectionSymbol)

	if ui != -1 && (ui < ii || ii == -1) {
		left, err := ParseType(value[:ui])
		if err != nil {
			return nil, err
		}

		right, err := ParseType(value[ui+1:])
		if err != nil {
			return nil, err
		}

		return &TypeUnion{
			Left:  left,
			Right: right,
		}, nil
	}

	if ii != -1 && (ii < ui || ui == -1) {
		left, err := ParseType(value[:ii])
		if err != nil {
			return nil, err
		}

		right, err := ParseType(value[ii+1:])
		if err != nil {
			return nil, err
		}

		return &TypeIntersection{
			Left:  left,
			Right: right,
		}, nil
	}

	return nil, fmt.Errorf("%s: %w", value, ErrUnknown)
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
		itemType, err := ParseType(arrMatch[arrRgxKeyG])
		if err != nil {
			return true, nil, fmt.Errorf("Error parsing array item type for %s: %w", value, err)
		}

		return true, &TypeArray{ItemType: itemType, NonEmpty: nonEmpty}, nil
	}

	keyType, err := ParseType(arrMatch[arrRgxKeyG])
	if err != nil {
		return true, nil, fmt.Errorf("Error parsing array key type for %s: %w", value, err)
	}

	itemType, err := ParseType(arrMatch[arrRgxValG])
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

	itemType, err := ParseType(arrMatch[typeArrRgxTypeG])
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

	fullyQualified := strings.HasPrefix(value, namespaceSeperator)

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

func parseIterable(value string) (bool, Type, error) {
	iterableMatch := iterableRegex.FindStringSubmatch(value)
	if len(iterableMatch) < iterRgxValG+1 {
		return false, nil, nil
	}

	keyTypeRaw := iterableMatch[iterRgxKeyG]
	keyType, err := ParseType(keyTypeRaw)
	if err != nil {
		return true, nil, fmt.Errorf(
			"Error parsing iterable key type %s of type %s: %w",
			keyTypeRaw,
			value,
			err,
		)
	}

	valTypeRaw := iterableMatch[iterRgxValG]
	if valTypeRaw == "" {
		return true, &TypeIterable{
			ItemType: keyType,
		}, nil
	}

	valType, err := ParseType(valTypeRaw)
	if err != nil {
		return true, nil, fmt.Errorf(
			"Error parsing iterable value type %s of type %s: %w",
			valTypeRaw,
			value,
			err,
		)
	}

	return true, &TypeIterable{
		KeyType:  keyType,
		ItemType: valType,
	}, nil
}

func parseConstrainedClassString(value string) (bool, Type, error) {
	match := constrainedClassStringRegex.FindStringSubmatch(value)
	if len(match) < constrClsStrNameG+1 {
		return false, nil, nil
	}

	fullyQualified := match[constrClsStrNameG][0:1] == namespaceSeperator

	return true, &TypeString{
		Constraint:  StringConstraintClass,
		GenericOver: &TypeClassLike{Name: match[constrClsStrNameG], FullyQualified: fullyQualified},
	}, nil
}

// PERF: this can probably be heavily optimized, running a lot of replaces, splits, and string checks here.
func parseArrayShape(value string) (bool, Type, error) {
	match := arrShapeRegexStep1.FindStringSubmatch(value)
	if len(match) < arrShapeRgxValsG+1 {
		return false, nil, nil
	}

	values := match[arrShapeRgxValsG]
	values = strings.ReplaceAll(values, "'", "")
	values = strings.ReplaceAll(values, `"`, "")
	values = strings.ReplaceAll(values, " ", "")

	indiVals := strings.Split(values, typeSeperator)
	vals := make([]*TypeArrayShapeValue, 0, len(indiVals))
	for i, val := range indiVals {
		keyval := strings.Split(val, ":")
		if len(keyval) == 1 {
			valType, err := ParseType(keyval[0])
			if err != nil {
				return true, nil, fmt.Errorf(
					"Error parsing array shape value %s of shape %s: %w",
					keyval[0],
					value,
					err,
				)
			}

			vals = append(vals, &TypeArrayShapeValue{Key: fmt.Sprintf("%d", i), Type: valType})
			continue
		}

		key := keyval[0]
		optional := strings.HasSuffix(key, "?")
		key = strings.TrimSuffix(key, "?")

		val := keyval[1]
		valType, err := ParseType(val)
		if err != nil {
			return true, nil, fmt.Errorf(
				"Error parsing array shape value %s of shape %s: %w",
				val,
				value,
				err,
			)
		}

		vals = append(vals, &TypeArrayShapeValue{Key: key, Type: valType, Optional: optional})
	}

	return true, &TypeArrayShape{Values: vals}, nil
}

func parseStringLiteral(value string) (bool, Type, error) {
	match := strLiteralRegex.FindStringSubmatch(value)
	if len(match) < strLitRgxValG+1 {
		return false, nil, nil
	}

	return true, &TypeStringLiteral{Value: match[strLitRgxValG]}, nil
}

func parseIntLiteral(value string) (bool, Type, error) {
	val, err := strconv.Atoi(value)
	if err != nil {
		return false, nil, nil
	}

	return true, &TypeIntLiteral{Value: val}, nil
}

func parseFloatLiteral(value string) (bool, Type, error) {
	val, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return false, nil, nil
	}

	return true, &TypeFloatLiteral{Value: val}, nil
}

func parseClassConst(value string) (bool, Type, error) {
	match := classConstRegex.FindStringSubmatch(value)
	if len(match) < clsCnstRgxCnstG+1 {
		return false, nil, nil
	}

	fullyQualified := match[clsCnstRgxClsG][0:1] == namespaceSeperator
	return true, &TypeConstant{
		Class: &TypeClassLike{Name: match[clsCnstRgxClsG], FullyQualified: fullyQualified},
		Const: match[clsCnstRgxCnstG],
	}, nil
}

func parseGlobalConst(value string) (bool, Type, error) {
	match := globalConstRegex.MatchString(value)
	if !match {
		return false, nil, nil
	}

	return true, &TypeConstant{Const: value}, nil
}

func parseCallable(value string) (bool, Type, error) {
	match := callableRegex.FindStringSubmatch(value)
	if len(match) < clbleRgxRtrnG+1 {
		return false, nil, nil
	}

	returnType, err := ParseType(match[clbleRgxRtrnG])
	if err != nil {
		return true, nil, fmt.Errorf(
			"Error parsing callable return type %s of %s: %w",
			match[clbleRgxRtrnG],
			value,
			err,
		)
	}

	params := []*CallableParameter{}
	for _, param := range strings.Split(match[clbleRgxPrmsG], typeSeperator) {
		if strings.TrimSpace(param) == "" {
			continue
		}

		variadic := strings.Contains(param, "...")
		byRef := strings.Contains(param, byRefSymbol)
		optional := strings.Contains(param, "=")

		param = strings.ReplaceAll(param, " ", "")
		param = strings.ReplaceAll(param, byRefSymbol, "")
		param = strings.ReplaceAll(param, ".", "")
		param = strings.ReplaceAll(param, "=", "")

		parts := strings.Split(param, "$")
		name := ""
		if len(parts) > 1 {
			name += "$" + parts[1]
		}

		paramType, err := ParseType(parts[0])
		if err != nil {
			return true, nil, fmt.Errorf(
				"Error parsing parameter %s of callable %s: %w",
				param,
				value,
				err,
			)
		}

		params = append(params, &CallableParameter{
			Optional: optional,
			Variadic: variadic,
			ByRef:    byRef,
			Type:     paramType,
			Name:     name,
		})
	}

	return true, &TypeCallable{
		Parameters: params,
		Return:     returnType,
	}, nil
}

func parseIntMask(value string) (bool, Type, error) {
	match := intMaskRegex.FindStringSubmatch(value)
	if len(match) < intMskRgxVlsG+1 {
		return false, nil, nil
	}

	valuesRaw := strings.Split(match[intMskRgxVlsG], typeSeperator)
	values := make([]int, 0, len(valuesRaw))
	for _, v := range valuesRaw {
		v = strings.TrimSpace(v)
		intVal, err := strconv.Atoi(v)
		if err != nil {
			return true, nil, fmt.Errorf(
				"Error parsing int-mask, %s of %s is not an int: %w",
				v,
				value,
				err,
			)
		}

		values = append(values, intVal)
	}

	return true, &TypeIntMask{
		Values: values,
	}, nil
}

func parseIntMaskOf(value string) (bool, Type, error) {
	match := intMaskOfRegex.FindStringSubmatch(value)
	if len(match) < intMskOfRgxTypG+1 {
		return false, nil, nil
	}

	typeVal, err := ParseType(match[intMskOfRgxTypG])
	if err != nil {
		return true, nil, fmt.Errorf(
			"Error parsing type %s of %s: %w",
			match[intMskOfRgxTypG],
			value,
			err,
		)
	}

	return true, &TypeIntMaskOf{Type: typeVal}, nil
}

// NOTE: be careful because a generic like T used in the check, if true
// NOTE: or if false part could come out as a TypeConstant instead of a TypeClassLike.
func parseConditionalReturn(value string) (bool, Type, error) {
	match := conditionalReturnRegex.FindStringSubmatch(value)
	if len(match) < cndRetRgxFalseG+1 {
		return false, nil, nil
	}

	right, err := ParseType(match[cndRetRgxCheckG])
	if err != nil {
		return true, nil, fmt.Errorf(
			"Error parsing check type %s of conditional return %s: %w",
			match[cndRetRgxCheckG],
			value,
			err,
		)
	}

	ifTrue, err := ParseType(match[cndRetRgxTrueG])
	if err != nil {
		return true, nil, fmt.Errorf(
			"Error parsing true type %s of conditional return %s: %w",
			match[cndRetRgxTrueG],
			value,
			err,
		)
	}

	ifFalse, err := ParseType(match[cndRetRgxFalseG])
	if err != nil {
		return true, nil, fmt.Errorf(
			"Error parsing false type %s of conditional return %s: %w",
			match[cndRetRgxFalseG],
			value,
			err,
		)
	}

	return true, &TypeConditionalReturn{
		Condition: &ConditionalReturnCondition{
			Left:  match[cndRetRgxPrmG],
			Right: right,
		},
		IfTrue:  ifTrue,
		IfFalse: ifFalse,
	}, nil
}

func parseGenericTemplate(value string) (bool, Type, error) {
	match := genericTemplateRegex.FindStringSubmatch(value)
	if len(match) < genTemRgxOfG+1 {
		return false, nil, nil
	}

	fullyQualified := match[genTemRgxOfG][0:1] == namespaceSeperator

	return true, &TypeGenericTemplate{
		Name: match[genTemRgxNameG],
		Of:   &TypeClassLike{Name: match[genTemRgxOfG], FullyQualified: fullyQualified},
	}, nil
}

func parseGenericClass(value string) (bool, Type, error) {
	match := genericClassRegex.FindStringSubmatch(value)
	if len(match) < genClsRgxGenOverG+1 {
		return false, nil, nil
	}

	name := match[genClsRgxNameG]
	fullyQualified := name[0:1] == namespaceSeperator

	rawGenOver := strings.Split(match[genClsRgxGenOverG], typeSeperator)
	genOver := make([]Type, 0, len(rawGenOver))
	for _, v := range rawGenOver {
		parsed, err := ParseType(v)
		if err != nil {
			return true, nil, fmt.Errorf("Error parsing generic type %s of %s: %w", v, value, err)
		}

		genOver = append(genOver, parsed)
	}

	return true, &TypeClassLike{
		Name:           name,
		FullyQualified: fullyQualified,
		GenericOver:    genOver,
	}, nil
}
