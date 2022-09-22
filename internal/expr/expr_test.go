package expr_test

import (
	"path/filepath"
	"reflect"
	"testing"

	"appliedgo.net/what"
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/config"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/expr"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/project"
	"github.com/laytan/elephp/internal/project/definition"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/annotated"
	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/pkg/stack"
	"github.com/laytan/elephp/pkg/symbol"
	"github.com/laytan/elephp/pkg/typer"
	"github.com/matryer/is"
	"github.com/samber/do"
)

func TestDown(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		name  string
		start ir.Node
		out   *stack.Stack[*expr.DownResolvement]
	}{
		{
			name: "simple property fetch",
			start: &ir.PropertyFetchExpr{
				Variable: &ir.SimpleVar{Name: "$foo"},
				Property: &ir.Identifier{Value: "bar"},
			},
			out: sliceStack([]*expr.DownResolvement{
				{ExprType: expr.ExprTypeVariable, Identifier: "$foo"},
				{ExprType: expr.ExprTypeProperty, Identifier: "bar"},
			}),
		},
		{
			name: "multi-level property fetch",
			start: &ir.PropertyFetchExpr{
				Variable: &ir.PropertyFetchExpr{
					Variable: &ir.PropertyFetchExpr{
						Variable: &ir.SimpleVar{Name: "$foo"},
						Property: &ir.Identifier{Value: "bar"},
					},
					Property: &ir.Identifier{Value: "foobar"},
				},
				Property: &ir.Identifier{Value: "foobar2"},
			},
			out: sliceStack([]*expr.DownResolvement{
				{ExprType: expr.ExprTypeVariable, Identifier: "$foo"},
				{ExprType: expr.ExprTypeProperty, Identifier: "bar"},
				{ExprType: expr.ExprTypeProperty, Identifier: "foobar"},
				{ExprType: expr.ExprTypeProperty, Identifier: "foobar2"},
			}),
		},
		{
			name: "method",
			start: &ir.MethodCallExpr{
				Variable: &ir.SimpleVar{Name: "$foo"},
				Method:   &ir.Identifier{Value: "bar"},
			},
			out: sliceStack([]*expr.DownResolvement{
				{ExprType: expr.ExprTypeVariable, Identifier: "$foo"},
				{ExprType: expr.ExprTypeMethod, Identifier: "bar"},
			}),
		},
		{
			name: "method after property",
			start: &ir.MethodCallExpr{
				Variable: &ir.PropertyFetchExpr{
					Variable: &ir.SimpleVar{Name: "$foo"},
					Property: &ir.Identifier{Value: "foobar"},
				},
				Method: &ir.Identifier{Value: "bar"},
			},
			out: sliceStack([]*expr.DownResolvement{
				{ExprType: expr.ExprTypeVariable, Identifier: "$foo"},
				{ExprType: expr.ExprTypeProperty, Identifier: "foobar"},
				{ExprType: expr.ExprTypeMethod, Identifier: "bar"},
			}),
		},
		{
			name: "method inside property chain",
			start: &ir.PropertyFetchExpr{
				Variable: &ir.MethodCallExpr{
					Variable: &ir.SimpleVar{Name: "$foo"},
					Method:   &ir.Identifier{Value: "foobar"},
				},
				Property: &ir.Identifier{Value: "bar"},
			},
			out: sliceStack([]*expr.DownResolvement{
				{ExprType: expr.ExprTypeVariable, Identifier: "$foo"},
				{ExprType: expr.ExprTypeMethod, Identifier: "foobar"},
				{ExprType: expr.ExprTypeProperty, Identifier: "bar"},
			}),
		},
		{
			name: "static start",
			start: &ir.MethodCallExpr{
				Variable: &ir.PropertyFetchExpr{
					Variable: &ir.StaticCallExpr{
						Class: &ir.Name{Value: "Test"},
						Call:  &ir.Identifier{Value: "foo"},
					},
					Property: &ir.Identifier{Value: "foobar"},
				},
				Method: &ir.Identifier{Value: "bar"},
			},
			out: sliceStack([]*expr.DownResolvement{
				{ExprType: expr.ExprTypeName, Identifier: "Test"},
				{ExprType: expr.ExprTypeStaticMethod, Identifier: "foo"},
				{ExprType: expr.ExprTypeProperty, Identifier: "foobar"},
				{ExprType: expr.ExprTypeMethod, Identifier: "bar"},
			}),
		},
		{
			name: "static end",
			start: &ir.StaticCallExpr{
				Class: &ir.SimpleVar{Name: "$foo"},
				Call:  &ir.Identifier{Value: "foo"},
			},
			out: sliceStack([]*expr.DownResolvement{
				{ExprType: expr.ExprTypeVariable, Identifier: "$foo"},
				{ExprType: expr.ExprTypeStaticMethod, Identifier: "foo"},
			}),
		},
		{
			name: "method call on function return",
			start: &ir.MethodCallExpr{
				Variable: &ir.FunctionCallExpr{
					Function: &ir.Name{Value: "foo"},
				},
				Method: &ir.Identifier{Value: "bar"},
			},
			out: sliceStack([]*expr.DownResolvement{
				{ExprType: expr.ExprTypeFunction, Identifier: "foo"},
				{ExprType: expr.ExprTypeMethod, Identifier: "bar"},
			}),
		},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()

			symbols := stack.New[*expr.DownResolvement]()
			expr.Down(expr.AllResolvers(), symbols, scenario.start)

			if !reflect.DeepEqual(symbols, scenario.out) {
				t.Fail()
				t.Logf("Got: %v\n Want: %v\n", symbols, scenario.out)
			}
		})
	}
}

func TestResolve(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	root := filepath.Join(pathutils.Root(), "test", "testdata", "expr")

	proj := setup(root, phpversion.EightOne())
	err := proj.ParseWithoutProgress()
	is.NoErr(err)

	scenarios := annotated.Aggregate(t, root)
	for group, gscenarios := range scenarios {
		group, gscenarios := group, gscenarios
		t.Run(group, func(t *testing.T) {
			t.Parallel()
			for name, scenario := range gscenarios {
				name, scenario := name, scenario
				t.Run(name, func(t *testing.T) {
					t.Parallel()
					is := is.New(t)

					if scenario.ShouldSkip {
						t.SkipNow()
					}

					if scenario.In.Path == "" {
						t.Fatalf("invalid test scenario, no in called for '%s'", name)
					}

					if !scenario.IsNoDef && scenario.Out == nil {
						t.Fatalf("invalid test scenario, no out called for '%s'", name)
					}

					ctx, err := context.New(&scenario.In)
					is.NoErr(err)
					ctx.Advance()

					what.Happens("Type: %T", ctx.Current())

					out := expr.Resolve(ctx)
					if out == nil && !scenario.IsNoDef {
						what.Is(ctx.Current())
						t.Fatalf("No result type for given scenario")
					}

					outCtx, err := context.New(scenario.Out)
					is.NoErr(err)
					outFqn := definition.FullyQualify(
						outCtx.Root(),
						symbol.GetIdentifier(outCtx.Current()),
					)

					if outFqn.String() != out.Name {
						t.Fatalf(
							"Returned type %v does not match expected type %v",
							outFqn.String(),
							out.Name,
						)
					}
				})
			}
		})
	}
}

func sliceStack(resolvements []*expr.DownResolvement) *stack.Stack[*expr.DownResolvement] {
	s := stack.New[*expr.DownResolvement]()
	for i := len(resolvements) - 1; i >= 0; i-- {
		s.Push(resolvements[i])
	}

	return s
}

func setup(root string, phpv *phpversion.PHPVersion) *project.Project {
	do.OverrideValue(nil, config.Default())
	do.OverrideValue(nil, index.New(phpv))
	do.OverrideValue(nil, wrkspc.New(phpv, root))
	do.OverrideValue(nil, typer.New())

	return project.New()
}
