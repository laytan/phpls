package phpdoxer_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/laytan/elephp/pkg/phpdoxer"
)

func TestParse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		args             string
		want             phpdoxer.Type
		wantErr          bool
		wantSpecificErr  error
		wantEqualStrings bool
	}{
		{
			name:            "empty",
			wantErr:         true,
			wantSpecificErr: phpdoxer.ErrEmpty,
		},
		{
			name:            "what is this",
			args:            "array<",
			wantErr:         true,
			wantSpecificErr: phpdoxer.ErrUnknown,
		},
		{
			name:             "mixed",
			args:             "mixed",
			want:             &phpdoxer.TypeMixed{},
			wantEqualStrings: true,
		},
		{
			name: "union",
			args: "array|string",
			want: &phpdoxer.TypeUnion{
				Left:  &phpdoxer.TypeArray{},
				Right: &phpdoxer.TypeString{},
			},
			wantEqualStrings: true,
		},
		{
			name: "this",
			args: "$this",
			want: &phpdoxer.TypeClassLike{Name: "$this"},
		},
		{
			name: "static",
			args: "static",
			want: &phpdoxer.TypeClassLike{Name: "static"},
		},
		{
			name: "intersection",
			args: "non-empty-array&positive-int",
			want: &phpdoxer.TypeIntersection{
				Left:  &phpdoxer.TypeArray{NonEmpty: true},
				Right: &phpdoxer.TypeInt{HasPositiveConstraint: true},
			},
			wantEqualStrings: true,
		},
		{
			name: "precedenced union & intersection",
			args: "(negative-int|string)&array",
			want: &phpdoxer.TypeIntersection{
				Left: &phpdoxer.TypePrecedence{
					Type: &phpdoxer.TypeUnion{
						Left:  &phpdoxer.TypeInt{HasNegativeConstraint: true},
						Right: &phpdoxer.TypeString{},
					},
				},
				Right: &phpdoxer.TypeArray{},
			},
			wantEqualStrings: true,
		},
		{
			name: "precedenced union at end & intersection",
			args: "array&(negative-int|string)",
			want: &phpdoxer.TypeIntersection{
				Left: &phpdoxer.TypeArray{},
				Right: &phpdoxer.TypePrecedence{
					Type: &phpdoxer.TypeUnion{
						Left:  &phpdoxer.TypeInt{HasNegativeConstraint: true},
						Right: &phpdoxer.TypeString{},
					},
				},
			},
			wantEqualStrings: true,
		},
		{
			name: "precedenced union in middle & intersection",
			args: "array&(negative-int|string)|int",
			want: &phpdoxer.TypeIntersection{
				Left: &phpdoxer.TypeArray{},
				Right: &phpdoxer.TypeUnion{
					Left: &phpdoxer.TypePrecedence{
						Type: &phpdoxer.TypeUnion{
							Left:  &phpdoxer.TypeInt{HasNegativeConstraint: true},
							Right: &phpdoxer.TypeString{},
						},
					},
					Right: &phpdoxer.TypeInt{},
				},
			},
			wantEqualStrings: true,
		},
		{
			name: "double precedence",
			args: "array&(negative-int|string)|(int|false)",
			// NOTE: this is in a weird order but is technically correct.
			want: &phpdoxer.TypeUnion{
				Left: &phpdoxer.TypeIntersection{
					Left: &phpdoxer.TypeArray{},
					Right: &phpdoxer.TypePrecedence{
						Type: &phpdoxer.TypeUnion{
							Left:  &phpdoxer.TypeInt{HasNegativeConstraint: true},
							Right: &phpdoxer.TypeString{},
						},
					},
				},
				Right: &phpdoxer.TypePrecedence{
					Type: &phpdoxer.TypeUnion{
						Left:  &phpdoxer.TypeInt{},
						Right: &phpdoxer.TypeBool{Accepts: phpdoxer.BoolAcceptsFalse},
					},
				},
			},
			wantEqualStrings: true,
		},
		{
			name:             "complex int",
			args:             "int<3, 5>",
			want:             &phpdoxer.TypeInt{Min: "3", Max: "5"},
			wantEqualStrings: true,
		},
		{
			name:             "complex int min",
			args:             "int<min, 5>",
			want:             &phpdoxer.TypeInt{Min: "min", Max: "5"},
			wantEqualStrings: true,
		},
		{
			name:             "complex int max",
			args:             "int<3, max>",
			want:             &phpdoxer.TypeInt{Min: "3", Max: "max"},
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
			want:             &phpdoxer.TypeInt{Min: "min", Max: "-1"},
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
			want:             &phpdoxer.TypeArray{ItemType: &phpdoxer.TypeString{}},
			wantEqualStrings: true,
		},
		{
			name:             "complex array nonempty",
			args:             "non-empty-array<string>",
			want:             &phpdoxer.TypeArray{NonEmpty: true, ItemType: &phpdoxer.TypeString{}},
			wantEqualStrings: true,
		},
		{
			name: "complex array key value",
			args: "array<string, int>",
			want: &phpdoxer.TypeArray{
				ItemType: &phpdoxer.TypeInt{},
				KeyType:  &phpdoxer.TypeString{},
			},
			wantEqualStrings: true,
		},
		{
			name: "complex array key value nonempty",
			args: "non-empty-array<string, int>",
			want: &phpdoxer.TypeArray{
				NonEmpty: true,
				ItemType: &phpdoxer.TypeInt{},
				KeyType:  &phpdoxer.TypeString{},
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
			want: &phpdoxer.TypeArray{ItemType: &phpdoxer.TypeString{}},
		},
		{
			name: "complex type array class",
			args: `\Test\Class[]`,
			want: &phpdoxer.TypeArray{
				ItemType: &phpdoxer.TypeClassLike{Name: `\Test\Class`, FullyQualified: true},
			},
		},
		{
			name: "complex type array class non-qualified",
			args: `Class[]`,
			want: &phpdoxer.TypeArray{ItemType: &phpdoxer.TypeClassLike{Name: `Class`}},
		},
		{
			name: "string constraints",
			args: `string|class-string|callable-string|numeric-string|non-empty-string|literal-string`,
			want: &phpdoxer.TypeUnion{
				Left: &phpdoxer.TypeString{Constraint: phpdoxer.StringConstraintNone},
				Right: &phpdoxer.TypeUnion{
					Left: &phpdoxer.TypeString{Constraint: phpdoxer.StringConstraintClass},
					Right: &phpdoxer.TypeUnion{
						Left: &phpdoxer.TypeString{Constraint: phpdoxer.StringConstraintCallable},
						Right: &phpdoxer.TypeUnion{
							Left: &phpdoxer.TypeString{
								Constraint: phpdoxer.StringConstraintNumeric,
							},
							Right: &phpdoxer.TypeUnion{
								Left: &phpdoxer.TypeString{
									Constraint: phpdoxer.StringConstraintNonEmpty,
								},
								Right: &phpdoxer.TypeString{
									Constraint: phpdoxer.StringConstraintLiteral,
								},
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
			want: &phpdoxer.TypeString{
				Constraint:  phpdoxer.StringConstraintClass,
				GenericOver: &phpdoxer.TypeClassLike{Name: "Foo"},
			},
			wantEqualStrings: true,
		},
		{
			name: "generic class string",
			args: `class-string<\Test\Foo>`,
			want: &phpdoxer.TypeString{
				Constraint:  phpdoxer.StringConstraintClass,
				GenericOver: &phpdoxer.TypeClassLike{Name: `\Test\Foo`, FullyQualified: true},
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
			want: &phpdoxer.TypeKeyOf{
				Class: &phpdoxer.TypeClassLike{Name: "Foo"},
				Const: "BAR",
			},
			wantEqualStrings: true,
		},
		{
			name: "key-of namespaced",
			args: `key-of<\Foo\FooBar::BAR>`,
			want: &phpdoxer.TypeKeyOf{
				Class: &phpdoxer.TypeClassLike{Name: `\Foo\FooBar`},
				Const: "BAR",
			},
			wantEqualStrings: true,
		},
		{
			name: "value-of",
			args: "value-of<Foo::BAR>",
			want: &phpdoxer.TypeValueOf{
				Class: &phpdoxer.TypeClassLike{Name: "Foo"},
				Const: "BAR",
			},
			wantEqualStrings: true,
		},
		{
			name: "value-of namespaced",
			args: `value-of<\Foo\FooBar::BAR>`,
			want: &phpdoxer.TypeValueOf{
				Class: &phpdoxer.TypeClassLike{Name: `\Foo\FooBar`},
				Const: "BAR",
			},
			wantEqualStrings: true,
		},
		{
			name: "value-of enum",
			args: `value-of<Foo>`,
			want: &phpdoxer.TypeValueOf{
				Class:  &phpdoxer.TypeClassLike{Name: `Foo`},
				IsEnum: true,
			},
			wantEqualStrings: true,
		},
		{
			name:             "iterable",
			args:             "iterable",
			want:             &phpdoxer.TypeIterable{},
			wantEqualStrings: true,
		},
		{
			name:             "iterable with type",
			args:             "iterable<int>",
			want:             &phpdoxer.TypeIterable{ItemType: &phpdoxer.TypeInt{}},
			wantEqualStrings: true,
		},
		{
			name: "iterable with key and value type",
			args: "iterable<string, int>",
			want: &phpdoxer.TypeIterable{
				KeyType:  &phpdoxer.TypeString{},
				ItemType: &phpdoxer.TypeInt{},
			},
			wantEqualStrings: true,
		},
		{
			name: "custom class",
			args: "Test<int, int>",
			want: &phpdoxer.TypeClassLike{
				Name:        "Test",
				GenericOver: []phpdoxer.Type{&phpdoxer.TypeInt{}, &phpdoxer.TypeInt{}},
			},
			wantEqualStrings: true,
		},
		{
			name: "custom class+namespace",
			args: `\Test\Test<int>`,
			want: &phpdoxer.TypeClassLike{
				Name:           `\Test\Test`,
				FullyQualified: true,
				GenericOver:    []phpdoxer.Type{&phpdoxer.TypeInt{}},
			},
			wantEqualStrings: true,
		},
		{
			name: "array shape 1",
			args: `array{'foo': int, "bar": string}`,
			want: &phpdoxer.TypeArrayShape{
				Values: []*phpdoxer.TypeArrayShapeValue{
					{Key: "foo", Type: &phpdoxer.TypeInt{}},
					{Key: "bar", Type: &phpdoxer.TypeString{}},
				},
			},
		},
		{
			name: "array shape 2",
			args: `array{0: int, 1?: int}`,
			want: &phpdoxer.TypeArrayShape{
				Values: []*phpdoxer.TypeArrayShapeValue{
					{Key: "0", Type: &phpdoxer.TypeInt{}},
					{Key: "1", Type: &phpdoxer.TypeInt{}, Optional: true},
				},
			},
		},
		{
			name: "array shape 3",
			args: `array{int, int}`,
			want: &phpdoxer.TypeArrayShape{
				Values: []*phpdoxer.TypeArrayShapeValue{
					{Key: "0", Type: &phpdoxer.TypeInt{}},
					{Key: "1", Type: &phpdoxer.TypeInt{}},
				},
			},
		},
		{
			name: "array shape 4",
			args: `array{foo:int, bar: string}`,
			want: &phpdoxer.TypeArrayShape{
				Values: []*phpdoxer.TypeArrayShapeValue{
					{Key: "foo", Type: &phpdoxer.TypeInt{}},
					{Key: "bar", Type: &phpdoxer.TypeString{}},
				},
			},
		},
		{
			name:             "string literal single quote",
			args:             "'foo'",
			want:             &phpdoxer.TypeStringLiteral{Value: "foo"},
			wantEqualStrings: true,
		},
		{
			name: "string literal double quote",
			args: `"foo"`,
			want: &phpdoxer.TypeStringLiteral{Value: "foo"},
		},
		{
			name: "string literal union",
			args: "'foo'|'bar'",
			want: &phpdoxer.TypeUnion{
				Left:  &phpdoxer.TypeStringLiteral{Value: "foo"},
				Right: &phpdoxer.TypeStringLiteral{Value: "bar"},
			},
			wantEqualStrings: true,
		},
		{
			name: "class constant",
			args: "Foo::SOME_CONSTANT",
			want: &phpdoxer.TypeConstant{
				Class: &phpdoxer.TypeClassLike{Name: "Foo"},
				Const: "SOME_CONSTANT",
			},
			wantEqualStrings: true,
		},
		{
			name: "class constant wildcard",
			args: "Foo::SOME_*",
			want: &phpdoxer.TypeConstant{
				Class: &phpdoxer.TypeClassLike{Name: "Foo"},
				Const: "SOME_*",
			},
			wantEqualStrings: true,
		},
		{
			name: "class constant namespaced",
			args: `Foo\Bar::SOME_CONSTANT`,
			want: &phpdoxer.TypeConstant{
				Class: &phpdoxer.TypeClassLike{Name: `Foo\Bar`},
				Const: "SOME_CONSTANT",
			},
			wantEqualStrings: true,
		},
		{
			name: "global constant",
			args: "SOME_CONSTANT",
			want: &phpdoxer.TypeConstant{
				Const: "SOME_CONSTANT",
			},
			wantEqualStrings: true,
		},
		{
			name:             "global constant wrong (require uppercase)",
			args:             "SOME_CONSTANt",
			want:             &phpdoxer.TypeClassLike{Name: "SOME_CONSTANt"},
			wantEqualStrings: true,
		},
		{
			name:             "callable",
			args:             "callable",
			want:             &phpdoxer.TypeCallable{},
			wantEqualStrings: true,
		},
		{
			name: "callable returns",
			args: "callable(): string",
			want: &phpdoxer.TypeCallable{
				Parameters: []*phpdoxer.CallableParameter{},
				Return:     &phpdoxer.TypeString{},
			},
			wantEqualStrings: true,
		},
		{
			name: "callable parameter",
			args: "callable(int): int",
			want: &phpdoxer.TypeCallable{
				Parameters: []*phpdoxer.CallableParameter{{Type: &phpdoxer.TypeInt{}}},
				Return:     &phpdoxer.TypeInt{},
			},
			wantEqualStrings: true,
		},
		{
			name: "callable double parameter",
			args: "callable(int, string): int",
			want: &phpdoxer.TypeCallable{
				Parameters: []*phpdoxer.CallableParameter{
					{Type: &phpdoxer.TypeInt{}},
					{Type: &phpdoxer.TypeString{}},
				},
				Return: &phpdoxer.TypeInt{},
			},
			wantEqualStrings: true,
		},
		{
			name: "callable double parameter with optional",
			args: "callable(int, int=): int",
			want: &phpdoxer.TypeCallable{
				Parameters: []*phpdoxer.CallableParameter{
					{Type: &phpdoxer.TypeInt{}},
					{Type: &phpdoxer.TypeInt{}, Optional: true},
				},
				Return: &phpdoxer.TypeInt{},
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
			want: &phpdoxer.TypeCallable{
				Parameters: []*phpdoxer.CallableParameter{
					{Type: &phpdoxer.TypeInt{}, Name: "$foo"},
					{Type: &phpdoxer.TypeInt{}, Name: "$bar"},
				},
				Return: &phpdoxer.TypeVoid{},
			},
			wantEqualStrings: true,
		},
		{
			name: "callable with parameter names, no spaces",
			args: "callable(int$foo,int$bar):void",
			want: &phpdoxer.TypeCallable{
				Parameters: []*phpdoxer.CallableParameter{
					{Type: &phpdoxer.TypeInt{}, Name: "$foo"},
					{Type: &phpdoxer.TypeInt{}, Name: "$bar"},
				},
				Return: &phpdoxer.TypeVoid{},
			},
		},
		{
			name: "callable with parameter passed by reference",
			args: "callable(string &$foo): mixed",
			want: &phpdoxer.TypeCallable{
				Parameters: []*phpdoxer.CallableParameter{
					{Type: &phpdoxer.TypeString{}, ByRef: true, Name: "$foo"},
				},
				Return: &phpdoxer.TypeMixed{},
			},
			wantEqualStrings: true,
		},
		{
			name: "callable with variadic args, returning union",
			args: "callable(float ...$floats): (int|null)",
			want: &phpdoxer.TypeCallable{
				Parameters: []*phpdoxer.CallableParameter{
					{Type: &phpdoxer.TypeFloat{}, Name: "$floats", Variadic: true},
				},
				Return: &phpdoxer.TypePrecedence{
					Type: &phpdoxer.TypeUnion{
						Left:  &phpdoxer.TypeInt{},
						Right: &phpdoxer.TypeNull{},
					},
				},
			},
			wantEqualStrings: true,
		},
		{
			name: "callable with variadic args, returning union",
			args: "callable(float...): (int|null)",
			want: &phpdoxer.TypeCallable{
				Parameters: []*phpdoxer.CallableParameter{
					{Type: &phpdoxer.TypeFloat{}, Variadic: true},
				},
				Return: &phpdoxer.TypePrecedence{
					Type: &phpdoxer.TypeUnion{
						Left:  &phpdoxer.TypeInt{},
						Right: &phpdoxer.TypeNull{},
					},
				},
			},
			wantEqualStrings: true,
		},
		{
			name:             "int-mask",
			args:             "int-mask<1, 2, 4>",
			want:             &phpdoxer.TypeIntMask{Values: []int{1, 2, 4}},
			wantEqualStrings: true,
		},
		{
			name: "int-mask-of union",
			args: "int-mask-of<1|2|4>",
			want: &phpdoxer.TypeIntMaskOf{
				Type: &phpdoxer.TypeUnion{
					Left: &phpdoxer.TypeIntLiteral{Value: 1},
					Right: &phpdoxer.TypeUnion{
						Left:  &phpdoxer.TypeIntLiteral{Value: 2},
						Right: &phpdoxer.TypeIntLiteral{Value: 4},
					},
				},
			},
			wantEqualStrings: true,
		},
		{
			name: "int-mask-of class constants",
			args: "int-mask-of<Foo::INT_*>",
			want: &phpdoxer.TypeIntMaskOf{
				Type: &phpdoxer.TypeConstant{
					Class: &phpdoxer.TypeClassLike{Name: "Foo"},
					Const: "INT_*",
				},
			},
			wantEqualStrings: true,
		},
		{
			name: "conditional return with parameter",
			args: "($foo is int ? int : string)",
			want: &phpdoxer.TypeConditionalReturn{
				Condition: &phpdoxer.ConditionalReturnCondition{
					Left:  "$foo",
					Right: &phpdoxer.TypeInt{},
				},
				IfTrue:  &phpdoxer.TypeInt{},
				IfFalse: &phpdoxer.TypeString{},
			},
			wantEqualStrings: true,
		},
		{
			name: "conditional return with generic type",
			args: "(T is array ? array<T> : null)",
			want: &phpdoxer.TypeConditionalReturn{
				Condition: &phpdoxer.ConditionalReturnCondition{
					Left:  "T",
					Right: &phpdoxer.TypeArray{},
				},
				IfTrue:  &phpdoxer.TypeArray{ItemType: &phpdoxer.TypeConstant{Const: "T"}},
				IfFalse: &phpdoxer.TypeNull{},
			},
			wantEqualStrings: true,
		},
		{
			name: "generic template",
			args: `T of \Exception`,
			want: &phpdoxer.TypeGenericTemplate{
				Name: "T",
				Of:   &phpdoxer.TypeClassLike{Name: `\Exception`, FullyQualified: true},
			},
			wantEqualStrings: true,
		},
		{
			name: "generic class",
			args: `\Generator<int, string>`,
			want: &phpdoxer.TypeClassLike{
				FullyQualified: true,
				Name:           `\Generator`,
				GenericOver:    []phpdoxer.Type{&phpdoxer.TypeInt{}, &phpdoxer.TypeString{}},
			},
			wantEqualStrings: true,
		},
		{
			name: "generic class multiple",
			args: `\Generator<int, string, int, Bar>`,
			want: &phpdoxer.TypeClassLike{
				FullyQualified: true,
				Name:           `\Generator`,
				GenericOver: []phpdoxer.Type{
					&phpdoxer.TypeInt{},
					&phpdoxer.TypeString{},
					&phpdoxer.TypeInt{},
					&phpdoxer.TypeClassLike{Name: "Bar"},
				},
			},
			wantEqualStrings: true,
		},
		{
			name: "short nullable",
			args: "?bool",
			want: &phpdoxer.TypeUnion{
				Left:  &phpdoxer.TypeNull{},
				Right: &phpdoxer.TypeBool{Accepts: phpdoxer.BoolAcceptsAll},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := phpdoxer.ParseType(tt.args)
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
