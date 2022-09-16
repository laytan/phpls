package phpdoxer_test

import (
	"reflect"
	"testing"

	"github.com/laytan/elephp/pkg/phpdoxer"
)

func TestParseDoc(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		args    string
		want    []phpdoxer.Node
		wantErr bool
	}{
		{
			name: "simple at return",
			args: `
            /**
             * @return string
             */
            `,
			want: []phpdoxer.Node{
				&phpdoxer.NodeReturn{
					Type: &phpdoxer.TypeString{},
				},
			},
		},
		{
			name: "simple at return single line block comment",
			args: "/** @return string */",
			want: []phpdoxer.Node{
				&phpdoxer.NodeReturn{
					Type: &phpdoxer.TypeString{},
				},
			},
		},
		{
			name: "simple at return single line comment",
			args: "// @return string",
			want: []phpdoxer.Node{
				&phpdoxer.NodeReturn{
					Type: &phpdoxer.TypeString{},
				},
			},
		},
		{
			name: "at return with description single line comment",
			args: "// @return string HelloWorld",
			want: []phpdoxer.Node{
				&phpdoxer.NodeReturn{
					Type:        &phpdoxer.TypeString{},
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
			want: []phpdoxer.Node{
				&phpdoxer.NodeReturn{
					Type:        &phpdoxer.TypeString{},
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
			want: []phpdoxer.Node{
				&phpdoxer.NodeReturn{
					Type:        &phpdoxer.TypeString{},
					Description: "Hello World",
				},
			},
		},
		{
			name: "simple @var",
			args: "// @var string",
			want: []phpdoxer.Node{
				&phpdoxer.NodeVar{
					Type: &phpdoxer.TypeString{},
				},
			},
		},
		{
			name: "block comment @var",
			args: "/** @var string */",
			want: []phpdoxer.Node{
				&phpdoxer.NodeVar{
					Type: &phpdoxer.TypeString{},
				},
			},
		},
		{
			name: "block comment @var multiline",
			args: `
            /**
             * @var string
             */`,
			want: []phpdoxer.Node{
				&phpdoxer.NodeVar{
					Type: &phpdoxer.TypeString{},
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
			want: []phpdoxer.Node{
				&phpdoxer.NodeUnknown{
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
			want: []phpdoxer.Node{
				&phpdoxer.NodeVar{
					Type: &phpdoxer.TypeString{},
				},
				&phpdoxer.NodeReturn{
					Type:        &phpdoxer.TypeString{},
					Description: "The return value is a string",
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := phpdoxer.ParseDoc(tt.args)
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
