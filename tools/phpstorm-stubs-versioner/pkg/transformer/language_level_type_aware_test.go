package transformer_test

import (
	"testing"

	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/tools/phpstorm-stubs-versioner/pkg/transformer"
)

func TestLanguageLevelTypeAware(t *testing.T) {
	t.Parallel()

	scenarios := []scenario{
		{
			name:    "above function as return type",
			version: "8.1",
			input: `
                <?php
                use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;
                #[LanguageLevelTypeAware(["8.0" => "InflateContext|false"], default: "resource|false")]
                function inflate_init() {}
            `,
			expected: `
                <?php
                use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;
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
                use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;
                #[LanguageLevelTypeAware(["8.0" => "InflateContext|false"], default: "resource|false")]
                function inflate_init() {}
            `,
			expected: `
                <?php
                use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;
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
                use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;
                #[LanguageLevelTypeAware(["8.0" => "InflateContext|false"], default: "resource|false")]
                #[Pure]
                function inflate_init() {}
            `,
			expected: `
                <?php
                use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;
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
                use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;
                #[Pure]
                #[LanguageLevelTypeAware(["8.0" => "InflateContext|false"], default: "resource|false")]
                function inflate_init() {}
            `,
			expected: `
                <?php
                use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;
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
                use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;
                /**
                 * @return int|false
                 */
                #[LanguageLevelTypeAware(['7.1' => 'int|false'], default: 'bool')]
                #[TentativeType]
                function gc($max_lifetime): int|false {}
            `,
			expected: `
                <?php
                use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;
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
                use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;
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
                use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;
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
                use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;
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
                use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;
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
                use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;
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
                use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;
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
                    use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;
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
                    use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;
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
                use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;
                function session_cache_expire(
                    #[LanguageLevelTypeAware(['8.0' => 'null|int'], default: 'int')] $value
                ) {}
            `,
			expected: `
                <?php
                use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;
                /**
                 * @param null|int $value
                 */
                function session_cache_expire($value) {}
            `,
		},
		{
			name:    "multiple constraints",
			version: "8.0.0",
			input: `
                <?php
                use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;
                #[LanguageLevelTypeAware([
                    '5.6' => 'bool',
                    '7.4' => 'null|true',
                ], default: '')]
                function session_cache_expire() {}
            `,
			expected: `
                <?php
                use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;
                /**
                 * @return null|true
                 */
                function session_cache_expire() {}
            `,
		},
		{
			name:    "multiple constraints 2",
			version: "8.0",
			input: `
                <?php
                use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;
                #[LanguageLevelTypeAware([
                    '5.6' => 'string|bool',
                    '7.4' => 'string',
                    '8.1' => 'non-empty-string',
                ], default: '')]
                function test(): bool {}
            `,
			expected: `
                <?php
                use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;
                /**
                 * @return string
                 */
                function test() {}
            `,
		},
		{
			name:    "doesn't do anything without use",
			version: "8.0",
			input: `
            <?php
            #[LanguageLevelTypeAware(['5.6' => 'string'], default: 'bool')]
            function test(): bool {}
            `,
			expected: `
            <?php
            #[LanguageLevelTypeAware(['5.6' => 'string'], default: 'bool')]
            function test(): bool {}
            `,
		},
		{
			name:    "aliasses",
			version: "8.0",
			input: `
            <?php
            use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware as VersionAware;
            #[VersionAware(['5.6' => 'string'], default: 'bool')]
            function test(): bool {}
            `,
			expected: `
            <?php
            use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware as VersionAware;
            /**
             * @return string
             */
            function test() {}
            `,
		},
		{
			name:    "fully qualified",
			version: "8.0",
			input: `
            <?php
            #[\JetBrains\PhpStorm\Internal\LanguageLevelTypeAware(['5.6' => 'string'], default: 'bool')]
            function test(): bool {}
            `,
			expected: `
            <?php
            /**
             * @return string
             */
            function test() {}
            `,
		},
		{
			name:    "namespace block resets usage",
			version: "8.0",
			input: `
            <?php
            use \JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;

            namespace Test {
                #[LanguageLevelTypeAware(['5.6' => 'string'], default: 'bool')]
                function test(): bool {}
            }
            `,
			expected: `
            <?php
            use \JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;

            namespace Test {
                #[LanguageLevelTypeAware(['5.6' => 'string'], default: 'bool')]
                function test(): bool {}
            }
            `,
		},
	}

	runScenarios(t, scenarios, func(v *phpversion.PHPVersion) transformer.Transformer {
		return transformer.NewLanguageLevelTypeAware(v, nil)
	})
}