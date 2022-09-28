package typer_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irconv"
	"github.com/VKCOM/php-parser/pkg/conf"
	pErrors "github.com/VKCOM/php-parser/pkg/errors"
	"github.com/VKCOM/php-parser/pkg/parser"
	"github.com/VKCOM/php-parser/pkg/version"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/resolvequeue"
	"github.com/laytan/elephp/pkg/typer"
)

func Test_typer_Returns(t *testing.T) {
	t.Parallel()
	type args struct {
		source             string
		funcOrMethodGetter func(*ir.Root) ir.Node
	}

	tests := []struct {
		name string
		args args
		want phpdoxer.Type
	}{
		{
			name: "doc comment return global function",
			args: args{
				source: `
                <?php
                /**
                 * @return string
                 */
                function foo() {}
                `,
				funcOrMethodGetter: func(r *ir.Root) ir.Node {
					return r.Stmts[1]
				},
			},
			want: &phpdoxer.TypeString{},
		},
		{
			name: "return typehint",
			args: args{
				source: `
		              <?php
		              function foo(): string {}
		              `,
				funcOrMethodGetter: func(r *ir.Root) ir.Node {
					return r.Stmts[1]
				},
			},
			want: &phpdoxer.TypeString{},
		},
		{
			name: "doc comment both, should return commented",
			args: args{
				source: `
                <?php
                /**
                 * @return non-empty-string
                 */
                function foo(): string {}
                `,
				funcOrMethodGetter: func(r *ir.Root) ir.Node {
					return r.Stmts[1]
				},
			},
			want: &phpdoxer.TypeString{Constraint: phpdoxer.StringConstraintNonEmpty},
		},
		{
			name: "doc comment class turned into FQN",
			args: args{
				source: `
                <?php
                /**
                 * @return FooBar
                 */
                function foo() {}
                `,
				funcOrMethodGetter: func(r *ir.Root) ir.Node {
					return r.Stmts[1]
				},
			},
			want: &phpdoxer.TypeClassLike{FullyQualified: true, Name: `\FooBar`},
		},
		{
			name: "type hint turned into FQN",
			args: args{
				source: `
                <?php
                namespace Testing;

                function foo(): Bar {}
                `,
				funcOrMethodGetter: func(r *ir.Root) ir.Node {
					return r.Stmts[2]
				},
			},
			want: &phpdoxer.TypeClassLike{FullyQualified: true, Name: `\Testing\Bar`},
		},
		{
			name: "type hint turned into FQN with use",
			args: args{
				source: `
                <?php
                namespace Testing;

                use Testing\Testerdetest\Bar;

                function foo(): Bar {}
                `,
				funcOrMethodGetter: func(r *ir.Root) ir.Node {
					return r.Stmts[3]
				},
			},
			want: &phpdoxer.TypeClassLike{FullyQualified: true, Name: `\Testing\Testerdetest\Bar`},
		},
		{
			name: "method typehint",
			args: args{
				source: `
                <?php
                namespace Testing;

                class Test {
                    public function test(): \Bar {}
                }
                `,
				funcOrMethodGetter: func(r *ir.Root) ir.Node {
					return r.Stmts[2].(*ir.ClassStmt).Stmts[0].(*ir.ClassMethodStmt)
				},
			},
			want: &phpdoxer.TypeClassLike{FullyQualified: true, Name: `\Bar`},
		},
		{
			name: "method doc comment",
			args: args{
				source: `
                <?php
                namespace Testing;

                class Test {
                    /**
                     * @return \Bar
                     */
                    public function test() {}
                }
                `,
				funcOrMethodGetter: func(r *ir.Root) ir.Node {
					return r.Stmts[2].(*ir.ClassStmt).Stmts[0].(*ir.ClassMethodStmt)
				},
			},
			want: &phpdoxer.TypeClassLike{FullyQualified: true, Name: `\Bar`},
		},
		{
			name: "Returned union null is removed",
			args: args{
				source: `
                <?php
                namespace Testing;

                class Test {
                    /**
                     * @return \Bar|null
                     */
                    public function test() {}
                }
                `,
				funcOrMethodGetter: func(r *ir.Root) ir.Node {
					return r.Stmts[2].(*ir.ClassStmt).Stmts[0].(*ir.ClassMethodStmt)
				},
			},
			want: &phpdoxer.TypeClassLike{FullyQualified: true, Name: `\Bar`},
		},
	}

	tr := typer.New()
	parseConfig := conf.Config{
		Version: &version.Version{
			Major: 8,
			Minor: 0,
		},
		ErrorHandlerFunc: func(e *pErrors.Error) {
			t.Error(e)
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ast, err := parser.Parse([]byte(tt.args.source), parseConfig)
			if err != nil {
				t.Error(err)
			}

			// dump := dumper.NewDumper(os.Stdout)
			// ast.Accept(dump)

			root := irconv.ConvertNode(ast).(*ir.Root)

			// TODO: create test cases that will use this.
			rootRetriever := func(n *resolvequeue.Node) (*ir.Root, error) {
				return nil, errors.New("current tests don't need this")
			}

			got := tr.Returns(root, tt.args.funcOrMethodGetter(root), rootRetriever)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("typer.Returns() = %v, want %v", got, tt.want)
			}
		})
	}
}
