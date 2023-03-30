package expr_test

import (
	"reflect"
	"testing"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/laytan/elephp/internal/expr"
	"github.com/laytan/elephp/pkg/stack"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestDown(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		name  string
		start ast.Vertex
		out   *stack.Stack[*expr.DownResolvement]
	}{
		{
			name: "simple property fetch",
			start: &ast.ExprPropertyFetch{
				Var:  &ast.ExprVariable{Name: &ast.Identifier{Value: []byte("$foo")}},
				Prop: &ast.Identifier{Value: []byte("bar")},
			},
			out: sliceStack([]*expr.DownResolvement{
				{ExprType: expr.TypeVariable, Identifier: "$foo"},
				{ExprType: expr.TypeProperty, Identifier: "bar"},
			}),
		},
		{
			name: "multi-level property fetch",
			start: &ast.ExprPropertyFetch{
				Var: &ast.ExprPropertyFetch{
					Var: &ast.ExprPropertyFetch{
						Var:  &ast.ExprVariable{Name: &ast.Identifier{Value: []byte("$foo")}},
						Prop: &ast.Identifier{Value: []byte("bar")},
					},
					Prop: &ast.Identifier{Value: []byte("foobar")},
				},
				Prop: &ast.Identifier{Value: []byte("foobar2")},
			},
			out: sliceStack([]*expr.DownResolvement{
				{ExprType: expr.TypeVariable, Identifier: "$foo"},
				{ExprType: expr.TypeProperty, Identifier: "bar"},
				{ExprType: expr.TypeProperty, Identifier: "foobar"},
				{ExprType: expr.TypeProperty, Identifier: "foobar2"},
			}),
		},
		{
			name: "method",
			start: &ast.ExprMethodCall{
				Var:    &ast.ExprVariable{Name: &ast.Identifier{Value: []byte("$foo")}},
				Method: &ast.Identifier{Value: []byte("bar")},
			},
			out: sliceStack([]*expr.DownResolvement{
				{ExprType: expr.TypeVariable, Identifier: "$foo"},
				{ExprType: expr.TypeMethod, Identifier: "bar"},
			}),
		},
		{
			name: "method after property",
			start: &ast.ExprMethodCall{
				Var: &ast.ExprPropertyFetch{
					Var:  &ast.ExprVariable{Name: &ast.Identifier{Value: []byte("$foo")}},
					Prop: &ast.Identifier{Value: []byte("foobar")},
				},
				Method: &ast.Identifier{Value: []byte("bar")},
			},
			out: sliceStack([]*expr.DownResolvement{
				{ExprType: expr.TypeVariable, Identifier: "$foo"},
				{ExprType: expr.TypeProperty, Identifier: "foobar"},
				{ExprType: expr.TypeMethod, Identifier: "bar"},
			}),
		},
		{
			name: "method inside property chain",
			start: &ast.ExprPropertyFetch{
				Var: &ast.ExprMethodCall{
					Var:    &ast.ExprVariable{Name: &ast.Identifier{Value: []byte("$foo")}},
					Method: &ast.Identifier{Value: []byte("foobar")},
				},
				Prop: &ast.Identifier{Value: []byte("bar")},
			},
			out: sliceStack([]*expr.DownResolvement{
				{ExprType: expr.TypeVariable, Identifier: "$foo"},
				{ExprType: expr.TypeMethod, Identifier: "foobar"},
				{ExprType: expr.TypeProperty, Identifier: "bar"},
			}),
		},
		{
			name: "static start",
			start: &ast.ExprMethodCall{
				Var: &ast.ExprPropertyFetch{
					Var: &ast.ExprStaticCall{
						Class: &ast.Name{Parts: []ast.Vertex{&ast.NamePart{Value: []byte("Test")}}},
						Call:  &ast.Identifier{Value: []byte("foo")},
					},
					Prop: &ast.Identifier{Value: []byte("foobar")},
				},
				Method: &ast.Identifier{Value: []byte("bar")},
			},
			out: sliceStack([]*expr.DownResolvement{
				{ExprType: expr.TypeName, Identifier: "Test"},
				{ExprType: expr.TypeStaticMethod, Identifier: "foo"},
				{ExprType: expr.TypeProperty, Identifier: "foobar"},
				{ExprType: expr.TypeMethod, Identifier: "bar"},
			}),
		},
		{
			name: "static end",
			start: &ast.ExprStaticCall{
				Class: &ast.ExprVariable{Name: &ast.Identifier{Value: []byte("$foo")}},
				Call:  &ast.Identifier{Value: []byte("foo")},
			},
			out: sliceStack([]*expr.DownResolvement{
				{ExprType: expr.TypeVariable, Identifier: "$foo"},
				{ExprType: expr.TypeStaticMethod, Identifier: "foo"},
			}),
		},
		{
			name: "method call on function return",
			start: &ast.ExprMethodCall{
				Var: &ast.ExprFunctionCall{
					Function: &ast.Name{Parts: []ast.Vertex{&ast.NamePart{Value: []byte("foo")}}},
				},
				Method: &ast.Identifier{Value: []byte("bar")},
			},
			out: sliceStack([]*expr.DownResolvement{
				{ExprType: expr.TypeFunction, Identifier: "foo"},
				{ExprType: expr.TypeMethod, Identifier: "bar"},
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
