package transformer_test

import (
	"bytes"
	"testing"

	"github.com/VKCOM/php-parser/pkg/conf"
	"github.com/VKCOM/php-parser/pkg/errors"
	"github.com/VKCOM/php-parser/pkg/parser"
	"github.com/VKCOM/php-parser/pkg/version"
	"github.com/VKCOM/php-parser/pkg/visitor/printer"
	"github.com/andreyvit/diff"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/pkg/strutil"
	"github.com/laytan/elephp/tools/phpstorm-stubs-versioner/pkg/transformer"
)

type scenario struct {
	name     string
	version  string
	input    string
	expected string
}

func runScenarios(
	t *testing.T,
	scenarios []scenario,
	createTransformer func(*phpversion.PHPVersion) transformer.Transformer,
) {
	t.Helper()

	parserConfig := conf.Config{
		Version: &version.Version{Major: 8, Minor: 1},
		ErrorHandlerFunc: func(e *errors.Error) {
			t.Error(e)
		},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()

			phpv, ok := phpversion.FromString(scenario.version)
			if !ok {
				t.Fatalf("invalid php version %s", scenario.version)
			}

			ast, err := parser.Parse([]byte(scenario.input), parserConfig)
			if err != nil {
				t.Fatal(err)
			}

			trans := createTransformer(phpv)
			trans.Transform(ast)

			out := bytes.NewBufferString("")
			p := printer.NewPrinter(out)
			ast.Accept(p)

			cExpected := strutil.RemoveWhitespace(scenario.expected)
			cOut := strutil.RemoveWhitespace(out.String())
			if cExpected != cOut {
				t.Errorf(
					"Result not as expected:\nWant: %v\nGot : %v\nDiff: %v",
					cExpected,
					cOut,
					diff.CharacterDiff(cExpected, cOut),
				)
			}
		})
	}
}
