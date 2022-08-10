package typer

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name             string
		args             string
		want             Type
		wantErr          bool
		wantEqualStrings bool
	}{
		{
			name:    "empty",
			wantErr: true,
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}

			if tt.wantEqualStrings && got.String() != tt.args {
				t.Errorf("Parse().String() = %v, want %v", got.String(), tt.args)
			}
		})
	}
}

func TestParseUnion(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    Type
		wantErr bool
	}{
		{
			name:    "empty",
			wantErr: true,
		},
		{
			name:    "one item",
			args:    []string{"string"},
			wantErr: true,
		},
		{
			name: "basic",
			args: []string{"string", "null"},
			want: &TypeUnion{Left: &TypeString{}, Right: &TypeNull{}},
		},
		{
			name: "complex",
			args: []string{"string", "false&true", "(int&array)&true"},
			want: &TypeUnion{
				Left: &TypeString{},
				Right: &TypeUnion{
					Left: &TypeIntersection{
						Left:  &TypeBool{Accepts: BoolAcceptsFalse},
						Right: &TypeBool{Accepts: BoolAcceptsTrue},
					},
					Right: &TypeIntersection{
						Left: &TypePrecedence{
							Type: &TypeIntersection{
								Left:  &TypeInt{},
								Right: &TypeArray{},
							},
						},
						Right: &TypeBool{Accepts: BoolAcceptsTrue},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseUnion(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseUnion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseUnion() = %v, want %v", got, tt.want)
			}
		})
	}
}
