package transformer_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/VKCOM/php-parser/pkg/conf"
	"github.com/VKCOM/php-parser/pkg/errors"
	"github.com/VKCOM/php-parser/pkg/parser"
	"github.com/VKCOM/php-parser/pkg/version"
	"github.com/VKCOM/php-parser/pkg/visitor/printer"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/tools/phpstorm-stubs-versioner/pkg/transformer"
)

func TestElementAvailableAttribute(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		name     string
		version  string
		input    string
		expected string
	}{
		{
			name:    "function that should be removed",
			version: "7.0",
			input: `
                <?php
                #[PhpStormStubsElementAvailable(from: '8.0')]
                function test(string $string, &$result): bool {}
            `,
			expected: "<?php",
		},
		{
			name:    "method that should be removed",
			version: "7.0",
			input: `
                <?php
                class Test {
                    #[PhpStormStubsElementAvailable(to: '6.0')]
                    public function test() {}
                }
            `,
			expected: `
                <?php
                class Test {
                }
            `,
		},
		{
			name:    "multiple parameters should remove the trailing comma",
			version: "7.0",
			input: `
                <?php
                function test(
                    $test,
                    #[PhpStormStubsElementAvailable(from: '8.0')] $query
                ) {}
            `,
			expected: `
                <?php
                function test(
                    $test
                ) {}
            `,
		},
		{
			name:    "more than 2 parameters should remove the trailing comma",
			version: "7.0",
			input: `
                <?php
                function test(
                    $test,
                    #[PhpStormStubsElementAvailable(from: '8.0')] $test2,
                    #[PhpStormStubsElementAvailable(from: '8.0')] $test3,
                    $test4
                ) {}
            `,
			expected: `
                <?php
                function test(
                    $test,
                    $test4
                ) {}
            `,
		},
		{
			name:    "multiple parameters with the same name should keep the PHPDoc",
			version: "7.0",
			input: `
                <?php
                /**
                 * @param string $query A string containing the queries to be executed. Multiple queries must be separated by a semicolon.
                 */
                function mysqli_multi_query(
                    #[PhpStormStubsElementAvailable(from: '5.3', to: '7.0')] string $query,
                    #[PhpStormStubsElementAvailable(from: '7.1', to: '7.4')] string $query = null
                ) {}
            `,
			expected: `
                <?php
                /**
                 * @param string $query A string containing the queries to be executed. Multiple queries must be separated by a semicolon.
                 */
                function mysqli_multi_query(
                    #[PhpStormStubsElementAvailable(from: '5.3', to: '7.0')] string $query
                ) {}
            `,
		},
		{
			name:    "multiple parameters with the same name that are all removed should still remove the PHPDoc",
			version: "7.0",
			input: `
                <?php
                /**
                 * @param string $query A string containing the queries to be executed. Multiple queries must be separated by a semicolon.
                 */
                function mysqli_multi_query(
                    #[PhpStormStubsElementAvailable(from: '7.1')] string $query,
                    #[PhpStormStubsElementAvailable(from: '7.1')] string $query = null
                ) {}
            `,
			expected: `
                <?php
                
                function mysqli_multi_query(
                ) {}
            `,
		},
		{
			name:    "basic method remove parameter",
			version: "7.0",
			input: `
                <?php
                class Test {
                    /**
                     * @param int $test
                     */
                    public function test(#[PhpStormStubsElementAvailable(to: '6.0')] $test) {}
                }
            `,
			expected: `
                <?php
                class Test {
                    
                    public function test() {}
                }
            `,
		},
		{
			name:    "keep top of documentation function",
			version: "7.0",
			input: `
                <?php
                /**
                 * Hello World!
                 *
                 * @param string $test
                 * @param string $no
                 */
                function test(
                    string $test,
                    #[PhpStormStubsElementAvailable(from: '8.0')] string $no
                ) {}
            `,
			expected: `
                <?php
                /**
                 * Hello World!
                 *
                 * @param string $test
                 */
                function test(
                    string $test
                ) {}
            `,
		},
		{
			name:    "keep param descriptions of function",
			version: "7.0",
			input: `
                <?php
                /**
                 * Hello World!
                 *
                 * @param string $test The test description.
                 * @param string $no The no description.
                 */
                function test(
                    string $test,
                    #[PhpStormStubsElementAvailable(from: '8.0')] string $no
                ) {}
            `,
			expected: `
                <?php
                /**
                 * Hello World!
                 *
                 * @param string $test The test description.
                 */
                function test(
                    string $test
                ) {}
            `,
		},
		{
			name:    "should have right amount of commas",
			version: "7.0",
			input: `
                <?php
                function test(
                    #[PhpStormStubsElementAvailable(from: '5.3', to: '7.0')] $query,
                    #[PhpStormStubsElementAvailable(from: '7.1', to: '7.4')] $query
                ) {}
            `,
			expected: `
                <?php
                function test(
                    #[PhpStormStubsElementAvailable(from: '5.3', to: '7.0')] $query
                ) {}
            `,
		},
		{
			name:    "variadic arguments",
			version: "7.0",
			input: `
                <?php
                /**
                 * @param mixed ...$args
                 */
                function test(
                    #[PhpStormStubsElementAvailable(from: '8.0')] ...$args
                ) {}
            `,
			expected: `
                <?php
                
                function test(
                ) {}
            `,
		},
	}

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

			trans := transformer.NewElementAvailableAttribute(phpv)
			trans.Transform(ast)

			out := bytes.NewBufferString("")
			p := printer.NewPrinter(out)
			ast.Accept(p)

			cExpected := strings.TrimSpace(scenario.expected)
			cOut := strings.TrimSpace(out.String())
			if cExpected != cOut {
				edits := myers.ComputeEdits(span.URIFromPath("expected.txt"), cExpected, cOut)
				t.Errorf(
					"Expected output and actual output don't match:\n%v",
					gotextdiff.ToUnified("expected.txt", "out.txt", cExpected, edits),
				)
			}
		})
	}
}
