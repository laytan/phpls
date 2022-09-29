package expr_test

import (
	"reflect"
	"testing"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/expr"
	"github.com/laytan/elephp/pkg/stack"
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

func sliceStack(resolvements []*expr.DownResolvement) *stack.Stack[*expr.DownResolvement] {
	s := stack.New[*expr.DownResolvement]()
	for i := len(resolvements) - 1; i >= 0; i-- {
		s.Push(resolvements[i])
	}

	return s
}