package phpdoxer_test

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
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
		{
			name: "param with description",
			args: `
            /**
             * @param string $test The test description.
             */
            `,
			want: []phpdoxer.Node{
				&phpdoxer.NodeParam{
					Type:        &phpdoxer.TypeString{},
					Name:        "$test",
					Description: "The test description.",
				},
			},
			wantErr: false,
		},
		{
			name: "param with multi line description",
			args: `
            /**
             * @param string $test The test description.
             *   Use at own risk.
             */
            `,
			want: []phpdoxer.Node{
				&phpdoxer.NodeParam{
					Type:        &phpdoxer.TypeString{},
					Name:        "$test",
					Description: "The test description.\nUse at own risk.",
				},
			},
			wantErr: false,
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

			// Use reflect to zero out the range of the node.
			// This way we only test the parsing and type assignment and
			// don't have to add start and end indexes which are very hard to
			// manually do in each test case.
			for _, gn := range got {
				f := reflect.ValueOf(gn).Elem()
				f.FieldByName("StartPos").SetInt(0)
				f.FieldByName("EndPos").SetInt(0)
			}

			if !reflect.DeepEqual(got, tt.want) {
				spew.Dump(got[0].Range())
				t.Errorf("ParseDoc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDoc_String(t *testing.T) {
	t.Parallel()

	type fields struct {
		Top         string
		Indentation string
		Nodes       []phpdoxer.Node
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "empty",
			fields: fields{
				Top:         "",
				Indentation: "",
				Nodes:       []phpdoxer.Node{},
			},
			want: "",
		},
		{
			name: "no top 1 param",
			fields: fields{
				Top:         "",
				Indentation: " ",
				Nodes: []phpdoxer.Node{
					&phpdoxer.NodeParam{
						Type: &phpdoxer.TypeString{},
						Name: "$test",
					},
				},
			},
			want: "/**\n * @param string $test\n */",
		},
		{
			name: "1 line top 1 param",
			fields: fields{
				Top:         "Hello world",
				Indentation: " ",
				Nodes: []phpdoxer.Node{
					&phpdoxer.NodeParam{
						Type: &phpdoxer.TypeString{},
						Name: "$test",
					},
				},
			},
			want: "/**\n * Hello world\n *\n * @param string $test\n */",
		},
		{
			name: "2 line top, 1 param",
			fields: fields{
				Top:         "Hello\nWorld!",
				Indentation: " ",
				Nodes: []phpdoxer.Node{
					&phpdoxer.NodeParam{
						Type: &phpdoxer.TypeString{},
						Name: "$test",
					},
				},
			},
			want: "/**\n * Hello\n * World!\n *\n * @param string $test\n */",
		},
		{
			name: "1 top 2 param",
			fields: fields{
				Top:         "Hello World!",
				Indentation: " ",
				Nodes: []phpdoxer.Node{
					&phpdoxer.NodeParam{
						Type: &phpdoxer.TypeString{},
						Name: "$test",
					},
					&phpdoxer.NodeParam{
						Type: &phpdoxer.TypeString{},
						Name: "$test2",
					},
				},
			},
			want: "/**\n * Hello World!\n *\n * @param string $test\n * @param string $test2\n */",
		},
		{
			name: "top, no param",
			fields: fields{
				Top:         "Hello World!",
				Indentation: " ",
				Nodes:       []phpdoxer.Node{},
			},
			want: "/**\n * Hello World!\n */",
		},
		{
			name: "2 top, no param",
			fields: fields{
				Top:         "Hello\nWorld!",
				Indentation: " ",
				Nodes:       []phpdoxer.Node{},
			},
			want: "/**\n * Hello\n * World!\n */",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := &phpdoxer.Doc{
				Top:         tt.fields.Top,
				Indentation: tt.fields.Indentation,
				Nodes:       tt.fields.Nodes,
			}
			if got := d.String(); got != tt.want {
				edits := myers.ComputeEdits(span.URIFromPath("expected.txt"), tt.want, got)
				t.Errorf(
					"Expected output and actual output don't match:\n%v",
					gotextdiff.ToUnified("expected.txt", "out.txt", tt.want, edits),
				)
			}
		})
	}
}
