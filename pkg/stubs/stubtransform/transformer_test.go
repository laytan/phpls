package stubtransform_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/conf"
	"github.com/VKCOM/php-parser/pkg/errors"
	"github.com/VKCOM/php-parser/pkg/parser"
	"github.com/VKCOM/php-parser/pkg/version"
	"github.com/VKCOM/php-parser/pkg/visitor/printer"
	"github.com/andreyvit/diff"
	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/pkg/strutil"
	"github.com/laytan/elephp/pkg/stubs/stubtransform"
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
	createTransformer func(*phpversion.PHPVersion) ast.Visitor,
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
			ast.Accept(trans)

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

type logger struct {
	b *testing.B
}

func (l *logger) Printf(format string, args ...any) {
}

func BenchmarkTransformer(b *testing.B) {
	stubsDir := filepath.Join(pathutils.Root(), "third_party", "phpstorm-stubs")
	outDir := filepath.Join(os.TempDir(), "elephp-stub-benchmark")

	l := &logger{b: b}
	version := phpversion.EightOne()

	clean := func() {
		err := os.RemoveAll(outDir)
		if err != nil {
			b.Error(err)
		}
	}

	clean()

	for i := 0; i < b.N; i++ {
		func() {
			defer clean()

			w := stubtransform.NewWalker(
				l,
				stubsDir,
				outDir,
				version,
				stubtransform.All(version, nil),
			)
			err := w.Walk()
			if err != nil {
				b.Error(err)
			}
		}()
	}
}
