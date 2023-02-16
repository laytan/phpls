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

func TestLanguageLevelTypeAware(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		name     string
		version  string
		input    string
		expected string
	}{
		{
			name:    "above function as return type",
			version: "8.1",
			input: `
                <?php
                #[LanguageLevelTypeAware(["8.0" => "InflateContext|false"], default: "resource|false")]
                function inflate_init() {}
            `,
			expected: `
                <?php
                /**
                 * @return InflateContext|false
                 */
                function inflate_init() {}
            `,
		},
		{
			name:    "above function as return type default",
			version: "7.4",
			input: `
                <?php
                #[LanguageLevelTypeAware(["8.0" => "InflateContext|false"], default: "resource|false")]
                function inflate_init() {}
            `,
			expected: `
                <?php
                /**
                 * @return resource|false
                 */
                function inflate_init() {}
            `,
		},
		{
			name:    "above function as return type with other attribute",
			version: "7.4",
			input: `
                <?php
                #[LanguageLevelTypeAware(["8.0" => "InflateContext|false"], default: "resource|false")]
                #[Pure]
                function inflate_init() {}
            `,
			expected: `
                <?php
                /**
                 * @return resource|false
                 */
                 #[Pure]
                function inflate_init() {}
            `,
		},
		{
			name:    "above function as return type with other attribute above",
			version: "7.4",
			input: `
                <?php
                #[Pure]
                #[LanguageLevelTypeAware(["8.0" => "InflateContext|false"], default: "resource|false")]
                function inflate_init() {}
            `,
			expected: `
                <?php
                /**
                 * @return resource|false
                 */
                 #[Pure]
                function inflate_init() {}
            `,
		},
		{
			name:    "above function as return type should remove return type hint",
			version: "7.0",
			input: `
                <?php
                /**
                 * @return int|false
                 */
                #[LanguageLevelTypeAware(['7.1' => 'int|false'], default: 'bool')]
                #[TentativeType]
                function gc($max_lifetime): int|false {}
            `,
			expected: `
                <?php
                /**
                 * @return bool
                 */
                #[TentativeType]
                function gc($max_lifetime) {}
            `,
		},
		{
			name:    "property var empty default == mixed",
			version: "7.0",
			input: `
                <?php
                class Test {
                    /**
                     * @var string Fully qualified class name where this method was defined
                     */
                    #[Immutable]
                    #[LanguageLevelTypeAware(['8.1' => 'string'], default: '')]
                    public $class;
                }
            `,
			expected: `
                <?php
                class Test {
                    /**
                     * @var mixed Fully qualified class name where this method was defined
                     */
                    #[Immutable]
                    public $class;
                }
            `,
		},
		{
			name:    "property var nothing else",
			version: "8.1",
			input: `
                <?php
                class Test {
                    /**
                     * @var string Fully qualified class name where this method was defined
                     */
                    #[LanguageLevelTypeAware(['8.1' => 'bool'], default: 'string')]
                    private string $class;
                }
            `,
			expected: `
                <?php
                class Test {
                    /**
                     * @var bool Fully qualified class name where this method was defined
                     */
                    private $class;
                }
            `,
		},
		{
			name:    "on methods (interface)",
			version: "7.0",
			input: `
                <?php
                interface Test {
                    /**
                     * @return int|false
                     */
                    #[LanguageLevelTypeAware(['7.1' => 'int|false'], default: 'bool')]
                    #[TentativeType]
                    function gc($max_lifetime): int|false;
                }
            `,
			expected: `
                <?php
                interface Test {
                    /**
                     * @return bool
                     */
                    #[TentativeType]
                    function gc($max_lifetime);
                }
            `,
		},
		{
			name:    "in namespace block",
			version: "8.1",
			input: `
                <?php
                namespace Test {
                    class Test {
                        /**
                         * @return int|false
                         */
                        #[LanguageLevelTypeAware(['7.1' => 'int|false'], default: 'bool')]
                        public function gc($max_lifetime): int|false;
                    }
                }
            `,
			expected: `
                <?php
                namespace Test {
                    class Test {
                        /**
                         * @return int|false
                         */
                        public function gc($max_lifetime);
                    }
                }
            `,
		},
		{
			name:    "in parameter",
			version: "8.0.0",
			input: `
                <?php
                function session_cache_expire(
                    #[LanguageLevelTypeAware(['8.0' => 'null|int'], default: 'int')] $value
                ) {}
            `,
			expected: `
                <?php
                /**
                 * @param null|int $value
                 */
                function session_cache_expire($value) {}
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

			trans := transformer.NewLanguageLevelTypeAware(phpv)
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
