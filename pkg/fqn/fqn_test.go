package fqn_test

import (
	"strconv"
	"testing"

	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/parsing"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/position"
	"github.com/laytan/php-parser/pkg/visitor/traverser"
	"github.com/matryer/is"
)

func TestFQN(t *testing.T) {
	t.Parallel()

	cases := []struct {
		FQN       string
		Name      string
		Namespace string
	}{
		{
			FQN:       "\\Testing\\One\\Two\\Three\\Four",
			Name:      "Four",
			Namespace: "Testing\\One\\Two\\Three",
		},
		{
			FQN:       "\\",
			Name:      "",
			Namespace: "",
		},
		{
			FQN:       "\\Test",
			Name:      "Test",
			Namespace: "",
		},
		{
			FQN:       "\\Test\\Test",
			Name:      "Test",
			Namespace: "Test",
		},
	}

	for i, test := range cases {
		i, test := i, test
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			f := fqn.New(test.FQN)
			is.Equal(f.String(), test.FQN)
			is.Equal(f.Name(), test.Name)
			is.Equal(f.Namespace(), test.Namespace)
		})
	}
}

func TestFQNTraverser(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		code   string
		name   *ast.Name
		expect string
	}{
		"single semicolon": {
			code: `
            <?php
            namespace Test;
            // Some comment.
            `,
			name: &ast.Name{
				Parts: []ast.Vertex{&ast.NamePart{Value: []byte("TestName")}},
				Position: &position.Position{
					StartLine: 3,
					StartPos:  50,
					EndLine:   3,
					EndPos:    51,
				},
			},
			expect: "\\Test\\TestName",
		},
		"no namespaces": {
			code: `
            <?php
            `,
			name: &ast.Name{
				Parts: []ast.Vertex{&ast.NamePart{Value: []byte("TestName")}},
				Position: &position.Position{
					StartLine: 1,
					StartPos:  1,
					EndLine:   1,
					EndPos:    2,
				},
			},
			expect: "\\TestName",
		},
		"single global namespace block": {
			code: `
            <?php
            namespace {
                // Some comment.
            }
            `,
			name: &ast.Name{
				Parts: []ast.Vertex{&ast.NamePart{Value: []byte("TestName")}},
				Position: &position.Position{
					StartLine: 3,
					StartPos:  20,
					EndLine:   3,
					EndPos:    24,
				},
			},
			expect: "\\TestName",
		},
		"single namespace block": {
			code: `
            <?php
            namespace Test {
                // Some comment.
            }
            `,
			name: &ast.Name{
				Parts: []ast.Vertex{&ast.NamePart{Value: []byte("TestName")}},
				Position: &position.Position{
					StartLine: 3,
					StartPos:  55,
					EndLine:   3,
					EndPos:    56,
				},
			},
			expect: "\\Test\\TestName",
		},
		"double semicolon": {
			code: `
            <?php
            namespace Test;
            // Some comment.
            namespace Test2;
            // Some comment 2.
            `,
			name: &ast.Name{
				Parts: []ast.Vertex{&ast.NamePart{Value: []byte("TestName")}},
				Position: &position.Position{
					StartLine: 3,
					StartPos:  55,
					EndLine:   3,
					EndPos:    56,
				},
			},
			expect: "\\Test\\TestName",
		},
		"double block": {
			code: `
            <?php
            namespace Test {
                // Some comment.
            }

            namespace Test2 {
                // Some comment 2.
            }
            `,
			name: &ast.Name{
				Parts: []ast.Vertex{&ast.NamePart{Value: []byte("TestName")}},
				Position: &position.Position{
					StartLine: 7,
					StartPos:  140,
					EndLine:   7,
					EndPos:    141,
				},
			},
			expect: "\\Test2\\TestName",
		},
		"mixed": {
			code: `
            <?php
            namespace {
                // Some comment.
            }
            namespace Test;
            use Test\Aliassed as Alias;
            // some comment 2.
            `,
			name: &ast.Name{
				Parts: []ast.Vertex{&ast.NamePart{Value: []byte("Alias")}},
				Position: &position.Position{
					StartLine: 6,
					StartPos:  130,
					EndLine:   6,
					EndPos:    131,
				},
			},
			expect: "\\Test\\Aliassed",
		},
		"block with use statement": {
			code: `
            <?php
            namespace Test1 {
                use Testing\Test;
                // Some comment.
            }
            `,
			name: &ast.Name{
				Parts: []ast.Vertex{&ast.NamePart{Value: []byte("Test")}},
				Position: &position.Position{
					StartLine: 4,
					StartPos:  60,
					EndLine:   4,
					EndPos:    61,
				},
			},
			expect: "\\Testing\\Test",
		},
	}

	parser := parsing.New(phpversion.EightOne())

	for name, test := range cases {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			root, err := parser.Parse([]byte(test.code))
			is.NoErr(err)

			fqnt := fqn.NewTraverser()
			fqntt := traverser.NewTraverser(fqnt)
			root.Accept(fqntt)

			result := fqnt.ResultFor(test.name)
			is.Equal(result.String(), test.expect)
		})
	}
}
