package phpdoxer

import (
	"errors"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name             string
		args             string
		want             Type
		wantErr          bool
		wantSpecificErr  error
		wantEqualStrings bool
	}{
		{
			name:            "empty",
			wantErr:         true,
			wantSpecificErr: ErrEmpty,
		},
		{
			name:            "what is this",
			args:            "array<",
			wantErr:         true,
			wantSpecificErr: ErrUnknown,
		},
		{
			name:             "mixed",
			args:             "mixed",
			want:             &TypeMixed{},
			wantEqualStrings: true,
		},
		{
			name:             "union",
			args:             "array|string",
			want:             &TypeUnion{Left: &TypeArray{}, Right: &TypeString{}},
			wantEqualStrings: true,
		},
		{
			name: "this",
			args: "$this",
			want: &TypeClassLike{Name: "$this"},
		},
		{
			name: "static",
			args: "static",
			want: &TypeClassLike{Name: "static"},
		},
		{
			name: "intersection",
			args: "non-empty-array&positive-int",
			want: &TypeIntersection{
				Left:  &TypeArray{NonEmpty: true},
				Right: &TypeInt{HasPositiveConstraint: true},
			},
			wantEqualStrings: true,
		},
		{
			name: "precedenced union & intersection",
			args: "(negative-int|string)&array",
			want: &TypeIntersection{
				Left: &TypePrecedence{
					Type: &TypeUnion{
						Left:  &TypeInt{HasNegativeConstraint: true},
						Right: &TypeString{},
					},
				},
				Right: &TypeArray{},
			},
			wantEqualStrings: true,
		},
		{
			name: "precedenced union at end & intersection",
			args: "array&(negative-int|string)",
			want: &TypeIntersection{
				Left: &TypeArray{},
				Right: &TypePrecedence{
					Type: &TypeUnion{
						Left:  &TypeInt{HasNegativeConstraint: true},
						Right: &TypeString{},
					},
				},
			},
			wantEqualStrings: true,
		},
		{
			name: "precedenced union in middle & intersection",
			args: "array&(negative-int|string)|int",
			want: &TypeIntersection{
				Left: &TypeArray{},
				Right: &TypeUnion{
					Left: &TypePrecedence{
						Type: &TypeUnion{
							Left:  &TypeInt{HasNegativeConstraint: true},
							Right: &TypeString{},
						},
					},
					Right: &TypeInt{},
				},
			},
			wantEqualStrings: true,
		},
		{
			name: "double precedence",
			args: "array&(negative-int|string)|(int|false)",
			// NOTE: this is in a weird order but is technically correct.
			want: &TypeUnion{
				Left: &TypeIntersection{
					Left: &TypeArray{},
					Right: &TypePrecedence{
						Type: &TypeUnion{
							Left:  &TypeInt{HasNegativeConstraint: true},
							Right: &TypeString{},
						},
					},
				},
				Right: &TypePrecedence{
					Type: &TypeUnion{
						Left:  &TypeInt{},
						Right: &TypeBool{Accepts: BoolAcceptsFalse},
					},
				},
			},
			wantEqualStrings: true,
		},
		{
			name:             "complex int",
			args:             "int<3, 5>",
			want:             &TypeInt{Min: "3", Max: "5"},
			wantEqualStrings: true,
		},
		{
			name:             "complex int min",
			args:             "int<min, 5>",
			want:             &TypeInt{Min: "min", Max: "5"},
			wantEqualStrings: true,
		},
		{
			name:             "complex int max",
			args:             "int<3, max>",
			want:             &TypeInt{Min: "3", Max: "max"},
			wantEqualStrings: true,
		},
		{
			name:    "int, invalid min: 'max'",
			args:    "int<max, 0>",
			wantErr: true,
		},
		{
			name:    "int, invalid max: 'min'",
			args:    "int<min, min>",
			wantErr: true,
		},
		{
			name:             "int, using negative",
			args:             "int<min, -1>",
			want:             &TypeInt{Min: "min", Max: "-1"},
			wantEqualStrings: true,
		},
		{
			name:    "int, min after max",
			args:    "int<0, -1>",
			wantErr: true,
		},
		{
			name:             "complex array",
			args:             "array<string>",
			want:             &TypeArray{ItemType: &TypeString{}},
			wantEqualStrings: true,
		},
		{
			name:             "complex array nonempty",
			args:             "non-empty-array<string>",
			want:             &TypeArray{NonEmpty: true, ItemType: &TypeString{}},
			wantEqualStrings: true,
		},
		{
			name:             "complex array key value",
			args:             "array<string, int>",
			want:             &TypeArray{ItemType: &TypeInt{}, KeyType: &TypeString{}},
			wantEqualStrings: true,
		},
		{
			name: "complex array key value nonempty",
			args: "non-empty-array<string, int>",
			want: &TypeArray{
				NonEmpty: true,
				ItemType: &TypeInt{},
				KeyType:  &TypeString{},
			},
			wantEqualStrings: true,
		},
		{
			name:    "complex array weird prefix",
			args:    "non-array<string>",
			wantErr: true,
		},
		{
			name: "complex type array",
			args: "string[]",
			want: &TypeArray{ItemType: &TypeString{}},
		},
		{
			name: "complex type array class",
			args: `\Test\Class[]`,
			want: &TypeArray{ItemType: &TypeClassLike{Name: `\Test\Class`, FullyQualified: true}},
		},
		{
			name: "complex type array class non-qualified",
			args: `Class[]`,
			want: &TypeArray{ItemType: &TypeClassLike{Name: `Class`}},
		},
		{
			name: "string constraints",
			args: `string|class-string|callable-string|numeric-string|non-empty-string|literal-string`,
			want: &TypeUnion{
				Left: &TypeString{Constraint: StringConstraintNone},
				Right: &TypeUnion{
					Left: &TypeString{Constraint: StringConstraintClass},
					Right: &TypeUnion{
						Left: &TypeString{Constraint: StringConstraintCallable},
						Right: &TypeUnion{
							Left: &TypeString{Constraint: StringConstraintNumeric},
							Right: &TypeUnion{
								Left:  &TypeString{Constraint: StringConstraintNonEmpty},
								Right: &TypeString{Constraint: stringConstraintLiteral},
							},
						},
					},
				},
			},
			wantEqualStrings: true,
		},
		{
			name: "generic class string",
			args: "class-string<Foo>",
			want: &TypeString{
				Constraint:  StringConstraintClass,
				GenericOver: &TypeClassLike{Name: "Foo"},
			},
			wantEqualStrings: true,
		},
		{
			name: "generic class string",
			args: `class-string<\Test\Foo>`,
			want: &TypeString{
				Constraint:  StringConstraintClass,
				GenericOver: &TypeClassLike{Name: `\Test\Foo`, FullyQualified: true},
			},
			wantEqualStrings: true,
		},
		{
			name:    "weird string",
			args:    "blabla-string",
			wantErr: true,
		},
		{
			name: "key-of",
			args: "key-of<Foo::BAR>",
			want: &TypeKeyOf{
				Class: &TypeClassLike{Name: "Foo"},
				Const: "BAR",
			},
			wantEqualStrings: true,
		},
		{
			name: "key-of namespaced",
			args: `key-of<\Foo\FooBar::BAR>`,
			want: &TypeKeyOf{
				Class: &TypeClassLike{Name: `\Foo\FooBar`},
				Const: "BAR",
			},
			wantEqualStrings: true,
		},
		{
			name: "value-of",
			args: "value-of<Foo::BAR>",
			want: &TypeValueOf{
				Class: &TypeClassLike{Name: "Foo"},
				Const: "BAR",
			},
			wantEqualStrings: true,
		},
		{
			name: "value-of namespaced",
			args: `value-of<\Foo\FooBar::BAR>`,
			want: &TypeValueOf{
				Class: &TypeClassLike{Name: `\Foo\FooBar`},
				Const: "BAR",
			},
			wantEqualStrings: true,
		},
		{
			name: "value-of enum",
			args: `value-of<Foo>`,
			want: &TypeValueOf{
				Class:  &TypeClassLike{Name: `Foo`},
				IsEnum: true,
			},
			wantEqualStrings: true,
		},
		{
			name:             "iterable",
			args:             "iterable",
			want:             &TypeIterable{},
			wantEqualStrings: true,
		},
		{
			name:             "iterable with type",
			args:             "iterable<int>",
			want:             &TypeIterable{ItemType: &TypeInt{}},
			wantEqualStrings: true,
		},
		{
			name:             "iterable with key and value type",
			args:             "iterable<string, int>",
			want:             &TypeIterable{KeyType: &TypeString{}, ItemType: &TypeInt{}},
			wantEqualStrings: true,
		},
		{
			name: "custom class",
			args: "Test<int, int>",
			want: &TypeClassLike{
				Name:        "Test",
				GenericOver: []Type{&TypeInt{}, &TypeInt{}},
			},
			wantEqualStrings: true,
		},
		{
			name: "custom class+namespace",
			args: `\Test\Test<int>`,
			want: &TypeClassLike{
				Name:           `\Test\Test`,
				FullyQualified: true,
				GenericOver:    []Type{&TypeInt{}},
			},
			wantEqualStrings: true,
		},
		{
			name: "array shape 1",
			args: `array{'foo': int, "bar": string}`,
			want: &TypeArrayShape{
				Values: []*TypeArrayShapeValue{
					{Key: "foo", Type: &TypeInt{}},
					{Key: "bar", Type: &TypeString{}},
				},
			},
		},
		{
			name: "array shape 2",
			args: `array{0: int, 1?: int}`,
			want: &TypeArrayShape{
				Values: []*TypeArrayShapeValue{
					{Key: "0", Type: &TypeInt{}},
					{Key: "1", Type: &TypeInt{}, Optional: true},
				},
			},
		},
		{
			name: "array shape 3",
			args: `array{int, int}`,
			want: &TypeArrayShape{
				Values: []*TypeArrayShapeValue{
					{Key: "0", Type: &TypeInt{}},
					{Key: "1", Type: &TypeInt{}},
				},
			},
		},
		{
			name: "array shape 4",
			args: `array{foo:int, bar: string}`,
			want: &TypeArrayShape{
				Values: []*TypeArrayShapeValue{
					{Key: "foo", Type: &TypeInt{}},
					{Key: "bar", Type: &TypeString{}},
				},
			},
		},
		{
			name:             "string literal single quote",
			args:             "'foo'",
			want:             &TypeStringLiteral{Value: "foo"},
			wantEqualStrings: true,
		},
		{
			name: "string literal double quote",
			args: `"foo"`,
			want: &TypeStringLiteral{Value: "foo"},
		},
		{
			name: "string literal union",
			args: "'foo'|'bar'",
			want: &TypeUnion{
				Left:  &TypeStringLiteral{Value: "foo"},
				Right: &TypeStringLiteral{Value: "bar"},
			},
			wantEqualStrings: true,
		},
		{
			name: "class constant",
			args: "Foo::SOME_CONSTANT",
			want: &TypeConstant{
				Class: &TypeClassLike{Name: "Foo"},
				Const: "SOME_CONSTANT",
			},
			wantEqualStrings: true,
		},
		{
			name: "class constant wildcard",
			args: "Foo::SOME_*",
			want: &TypeConstant{
				Class: &TypeClassLike{Name: "Foo"},
				Const: "SOME_*",
			},
			wantEqualStrings: true,
		},
		{
			name: "class constant namespaced",
			args: `Foo\Bar::SOME_CONSTANT`,
			want: &TypeConstant{
				Class: &TypeClassLike{Name: `Foo\Bar`},
				Const: "SOME_CONSTANT",
			},
			wantEqualStrings: true,
		},
		{
			name: "global constant",
			args: "SOME_CONSTANT",
			want: &TypeConstant{
				Const: "SOME_CONSTANT",
			},
			wantEqualStrings: true,
		},
		{
			name:             "global constant wrong (require uppercase)",
			args:             "SOME_CONSTANt",
			want:             &TypeClassLike{Name: "SOME_CONSTANt"},
			wantEqualStrings: true,
		},
		{
			name:             "callable",
			args:             "callable",
			want:             &TypeCallable{},
			wantEqualStrings: true,
		},
		{
			name: "callable returns",
			args: "callable(): string",
			want: &TypeCallable{
				Parameters: []*CallableParameter{},
				Return:     &TypeString{},
			},
			wantEqualStrings: true,
		},
		{
			name: "callable parameter",
			args: "callable(int): int",
			want: &TypeCallable{
				Parameters: []*CallableParameter{{Type: &TypeInt{}}},
				Return:     &TypeInt{},
			},
			wantEqualStrings: true,
		},
		{
			name: "callable double parameter",
			args: "callable(int, string): int",
			want: &TypeCallable{
				Parameters: []*CallableParameter{
					{Type: &TypeInt{}},
					{Type: &TypeString{}},
				},
				Return: &TypeInt{},
			},
			wantEqualStrings: true,
		},
		{
			name: "callable double parameter with optional",
			args: "callable(int, int=): int",
			want: &TypeCallable{
				Parameters: []*CallableParameter{
					{Type: &TypeInt{}},
					{Type: &TypeInt{}, Optional: true},
				},
				Return: &TypeInt{},
			},
			wantEqualStrings: true,
		},
		{
			name:    "callable without return (not allowed)",
			args:    "callable(int, int=)",
			wantErr: true,
		},
		{
			name: "callable with parameter names",
			args: "callable(int $foo, int $bar): void",
			want: &TypeCallable{
				Parameters: []*CallableParameter{
					{Type: &TypeInt{}, Name: "$foo"},
					{Type: &TypeInt{}, Name: "$bar"},
				},
				Return: &TypeVoid{},
			},
			wantEqualStrings: true,
		},
		{
			name: "callable with parameter names, no spaces",
			args: "callable(int$foo,int$bar):void",
			want: &TypeCallable{
				Parameters: []*CallableParameter{
					{Type: &TypeInt{}, Name: "$foo"},
					{Type: &TypeInt{}, Name: "$bar"},
				},
				Return: &TypeVoid{},
			},
		},
		{
			name: "callable with parameter passed by reference",
			args: "callable(string &$foo): mixed",
			want: &TypeCallable{
				Parameters: []*CallableParameter{
					{Type: &TypeString{}, ByRef: true, Name: "$foo"},
				},
				Return: &TypeMixed{},
			},
			wantEqualStrings: true,
		},
		{
			name: "callable with variadic args, returning union",
			args: "callable(float ...$floats): (int|null)",
			want: &TypeCallable{
				Parameters: []*CallableParameter{
					{Type: &TypeFloat{}, Name: "$floats", Variadic: true},
				},
				Return: &TypePrecedence{Type: &TypeUnion{Left: &TypeInt{}, Right: &TypeNull{}}},
			},
			wantEqualStrings: true,
		},
		{
			name: "callable with variadic args, returning union",
			args: "callable(float...): (int|null)",
			want: &TypeCallable{
				Parameters: []*CallableParameter{
					{Type: &TypeFloat{}, Variadic: true},
				},
				Return: &TypePrecedence{Type: &TypeUnion{Left: &TypeInt{}, Right: &TypeNull{}}},
			},
			wantEqualStrings: true,
		},
		{
			name:             "int-mask",
			args:             "int-mask<1, 2, 4>",
			want:             &TypeIntMask{Values: []int{1, 2, 4}},
			wantEqualStrings: true,
		},
		{
			name: "int-mask-of union",
			args: "int-mask-of<1|2|4>",
			want: &TypeIntMaskOf{
				Type: &TypeUnion{
					Left: &TypeIntLiteral{Value: 1},
					Right: &TypeUnion{
						Left:  &TypeIntLiteral{Value: 2},
						Right: &TypeIntLiteral{Value: 4},
					},
				},
			},
			wantEqualStrings: true,
		},
		{
			name: "int-mask-of class constants",
			args: "int-mask-of<Foo::INT_*>",
			want: &TypeIntMaskOf{
				Type: &TypeConstant{
					Class: &TypeClassLike{Name: "Foo"},
					Const: "INT_*",
				},
			},
			wantEqualStrings: true,
		},
		{
			name: "conditional return with parameter",
			args: "($foo is int ? int : string)",
			want: &TypeConditionalReturn{
				Condition: &ConditionalReturnCondition{
					Left:  "$foo",
					Right: &TypeInt{},
				},
				IfTrue:  &TypeInt{},
				IfFalse: &TypeString{},
			},
			wantEqualStrings: true,
		},
		{
			name: "conditional return with generic type",
			args: "(T is array ? array<T> : null)",
			want: &TypeConditionalReturn{
				Condition: &ConditionalReturnCondition{
					Left:  "T",
					Right: &TypeArray{},
				},
				IfTrue:  &TypeArray{ItemType: &TypeConstant{Const: "T"}},
				IfFalse: &TypeNull{},
			},
			wantEqualStrings: true,
		},
		{
			name: "generic template",
			args: `T of \Exception`,
			want: &TypeGenericTemplate{
				Name: "T",
				Of:   &TypeClassLike{Name: `\Exception`, FullyQualified: true},
			},
			wantEqualStrings: true,
		},
		{
			name: "generic class",
			args: `\Generator<int, string>`,
			want: &TypeClassLike{
				FullyQualified: true,
				Name:           `\Generator`,
				GenericOver:    []Type{&TypeInt{}, &TypeString{}},
			},
			wantEqualStrings: true,
		},
		{
			name: "generic class multiple",
			args: `\Generator<int, string, int, Bar>`,
			want: &TypeClassLike{
				FullyQualified: true,
				Name:           `\Generator`,
				GenericOver: []Type{
					&TypeInt{},
					&TypeString{},
					&TypeInt{},
					&TypeClassLike{Name: "Bar"},
				},
			},
			wantEqualStrings: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseType(tt.args)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if tt.wantSpecificErr != nil {
					if !errors.Is(err, tt.wantSpecificErr) {
						t.Errorf("Parse() error %v, wantErr %v", err, tt.wantSpecificErr)
					}
				}
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %#v, want %v", got, tt.want)
			}

			if tt.wantEqualStrings && got.String() != tt.args {
				t.Errorf("Parse().String() = %v, want %v", got.String(), tt.args)
			}
		})
	}
}