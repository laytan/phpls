package phpdoxer

import (
	"reflect"
	"testing"
)

func TestParseDoc(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		args    string
		want    []Node
		wantErr bool
	}{
		{
			name: "simple at return",
			args: `
            /**
             * @return string
             */
            `,
			want: []Node{
				&NodeReturn{
					Type: &TypeString{},
				},
			},
		},
		{
			name: "simple at return single line block comment",
			args: "/** @return string */",
			want: []Node{
				&NodeReturn{
					Type: &TypeString{},
				},
			},
		},
		{
			name: "simple at return single line comment",
			args: "// @return string",
			want: []Node{
				&NodeReturn{
					Type: &TypeString{},
				},
			},
		},
		{
			name: "at return with description single line comment",
			args: "// @return string HelloWorld",
			want: []Node{
				&NodeReturn{
					Type:        &TypeString{},
					Description: "HelloWorld",
				},
			},
		},
		{
			name: "at return with description multi line comment",
			args: `
            /**
             * @return string HelloWorld
             */
            `,
			want: []Node{
				&NodeReturn{
					Type:        &TypeString{},
					Description: "HelloWorld",
				},
			},
		},
		{
			name: "at return with description on line below",
			args: `
            /**
             * @return string
             *   Hello World
             *
             */
            `,
			want: []Node{
				&NodeReturn{
					Type:        &TypeString{},
					Description: "Hello World",
				},
			},
		},
		{
			name: "simple @var",
			args: "// @var string",
			want: []Node{
				&NodeVar{
					Type: &TypeString{},
				},
			},
		},
		{
			name: "block comment @var",
			args: "/** @var string */",
			want: []Node{
				&NodeVar{
					Type: &TypeString{},
				},
			},
		},
		{
			name: "block comment @var multiline",
			args: `
            /**
             * @var string
             */`,
			want: []Node{
				&NodeVar{
					Type: &TypeString{},
				},
			},
		},
		{
			name: "unknown",
			args: `
            /**
             * @methody lalosie posie
             */
            `,
			want: []Node{
				&NodeUnknown{
					At:    "methody",
					Value: "lalosie posie",
				},
			},
		},
		{
			name: "multiple @'s",
			args: `
            /**
             * @var string
             *
             * @return string
             *   The return value is a string
             *
             */
            `,
			want: []Node{
				&NodeVar{
					Type: &TypeString{},
				},
				&NodeReturn{
					Type:        &TypeString{},
					Description: "The return value is a string",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseDoc(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDoc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseDoc() = %v, want %v", got, tt.want)
			}
		})
	}
}
