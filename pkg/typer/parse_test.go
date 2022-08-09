package typer

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	type args struct {
		class *FQN
		value string
	}
	tests := []struct {
		name             string
		args             args
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
			args:             args{nil, "mixed"},
			want:             &TypeMixed{},
			wantEqualStrings: true,
		},
		{
			name:             "union",
			args:             args{nil, "array|string"},
			want:             &TypeUnion{Left: &TypeArray{}, Right: &TypeString{}},
			wantEqualStrings: true,
		},
		{
			name: "this",
			args: args{&FQN{value: `\Test`}, "$this"},
			want: &TypeClassLike{FQN: &FQN{value: `\Test`}},
		},
		{
			name: "static",
			args: args{&FQN{value: `\Test`}, "static"},
			want: &TypeClassLike{FQN: &FQN{value: `\Test`}},
		},
		{
			name: "intersection",
			args: args{nil, "non-empty-array&positive-int"},
			want: &TypeIntersection{
				Left:  &TypeArray{NonEmpty: true},
				Right: &TypeInt{HasPositiveConstraint: true},
			},
			wantEqualStrings: true,
		},
		{
			name: "precedenced union & intersection",
			args: args{nil, "(negative-int|string)&array"},
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
			args: args{nil, "array&(negative-int|string)"},
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
			args: args{nil, "array&(negative-int|string)|int"},
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
			args: args{nil, "array&(negative-int|string)|(int|false)"},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args.class, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}

			if tt.wantEqualStrings && got.String() != tt.args.value {
				t.Errorf("Parse().String() = %v, want %v", got.String(), tt.args.value)
			}
		})
	}
}

func TestParseUnion(t *testing.T) {
	type args struct {
		class *FQN
		value []string
	}

	tests := []struct {
		name    string
		args    args
		want    Type
		wantErr bool
	}{
		{
			name:    "empty",
			wantErr: true,
		},
		{
			name: "one item",
			args: args{
				value: []string{"string"},
			},
			wantErr: true,
		},
		{
			name: "basic",
			args: args{
				value: []string{"string", "null"},
			},
			want: &TypeUnion{Left: &TypeString{}, Right: &TypeNull{}},
		},
		{
			name: "complex",
			args: args{
				value: []string{"string", "false&true", "(int&array)&true"},
			},
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
			got, err := ParseUnion(tt.args.class, tt.args.value)
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
