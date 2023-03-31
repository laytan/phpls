package stubtransform_test

import (
	"testing"

	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/pkg/stubs/stubtransform"
	"github.com/laytan/php-parser/pkg/ast"
)

func TestElementAvailableAttribute(t *testing.T) {
	t.Parallel()

	scenarios := []scenario{
		{
			name:    "function that should be removed",
			version: "7.0",
			input: `
                <?php
                use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;
                #[PhpStormStubsElementAvailable(from: '8.0')]
                function test(string $string, &$result): bool {}
            `,
			expected: `
                <?php
                use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;
            `,
		},
		{
			name:    "method that should be removed",
			version: "7.0",
			input: `
                <?php
                use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;
                class Test {
                    #[PhpStormStubsElementAvailable(to: '6.0')]
                    public function test() {}
                }
            `,
			expected: `
                <?php
                use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;
                class Test {
                }
            `,
		},
		{
			name:    "multiple parameters should remove the trailing comma",
			version: "7.0",
			input: `
                <?php
                use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;
                function test(
                    $test,
                    #[PhpStormStubsElementAvailable(from: '8.0')] $query
                ) {}
            `,
			expected: `
                <?php
                use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;
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
                use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;
                function test(
                    $test,
                    #[PhpStormStubsElementAvailable(from: '8.0')] $test2,
                    #[PhpStormStubsElementAvailable(from: '8.0')] $test3,
                    $test4
                ) {}
            `,
			expected: `
                <?php
                use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;
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
                use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;
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
                use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;
                /**
                 * @param string $query A string containing the queries to be executed. Multiple queries must be separated by a semicolon.
                 */
                function mysqli_multi_query( string $query
                ) {}
            `,
		},
		{
			name:    "multiple parameters with the same name that are all removed should still remove the PHPDoc",
			version: "7.0",
			input: `
                <?php
                use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;
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
                use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;
                function mysqli_multi_query(
                ) {}
            `,
		},
		{
			name:    "basic method remove parameter",
			version: "7.0",
			input: `
                <?php
                use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;
                class Test {
                    /**
                     * @param int $test
                     */
                    public function test(#[PhpStormStubsElementAvailable(to: '6.0')] $test) {}
                }
            `,
			expected: `
                <?php
                use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;
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
                use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;
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
                use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;
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
                use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;
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
                use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;
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
                use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;
                function test(
                    #[PhpStormStubsElementAvailable(from: '5.3', to: '7.0')] $query,
                    #[PhpStormStubsElementAvailable(from: '7.1', to: '7.4')] $query
                ) {}
            `,
			expected: `
                <?php
                use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;
                function test( $query
                ) {}
            `,
		},
		{
			name:    "variadic arguments",
			version: "7.0",
			input: `
                <?php
                use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;
                /**
                 * @param mixed ...$args
                 */
                function test(
                    #[PhpStormStubsElementAvailable(from: '8.0')] ...$args
                ) {}
            `,
			expected: `
                <?php
                use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;
                function test(
                ) {}
            `,
		},
		{
			name:    "variadic and normal arguments mixed",
			version: "7.0",
			input: `
                <?php
                use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;
                /**
                 * Find lowest value
                 * @link https://php.net/manual/en/function.min.php
                 * @param array|mixed $value Array to look through or first value to compare
                 * @param mixed ...$values any comparable value
                 * @return mixed min returns the numerically lowest of the
                 * parameter values.
                 */
                #[Pure]
                function min(
                    #[PhpStormStubsElementAvailable(from: '8.0')] mixed $value,
                    #[PhpStormStubsElementAvailable(from: '5.3', to: '7.4')] mixed $values,
                    mixed ...$values
                ): mixed {}
            `,
			expected: `
                <?php
                use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;
                /**
                 * Find lowest value
                 *
                 * @link https://php.net/manual/en/function.min.php
                 * @param mixed ...$values any comparable value
                 * @return mixed min returns the numerically lowest of the
                 * parameter values.
                 */
                #[Pure]
                function min( mixed $values,
                    mixed ...$values
                ): mixed {}
            `,
		},
		{
			name:    "variable references and variadic",
			version: "7.0",
			input: `
                <?php
                use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;
                /**
                 * Sort multiple or multi-dimensional arrays
                 * @link https://php.net/manual/en/function.array-multisort.php
                 * @param array &$array <p>
                 * An array being sorted.
                 * </p>
                 * @param  &...$rest [optional] <p>
                 * More arrays, optionally followed by sort order and flags.
                 * Only elements corresponding to equivalent elements in previous arrays are compared.
                 * In other words, the sort is lexicographical.
                 * </p>
                 * @return bool true on success or false on failure.
                 */
                function array_multisort(
                    #[PhpStormStubsElementAvailable(from: '8.0')] &$array,
                    #[PhpStormStubsElementAvailable(from: '5.3', to: '7.4')] $sort_order = SORT_ASC,
                    #[PhpStormStubsElementAvailable(from: '5.3', to: '7.4')] $sort_flags = SORT_REGULAR,
                    &...$rest
                ): bool {}
            `,
			expected: `
                <?php
                use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;
                /**
                 * Sort multiple or multi-dimensional arrays
                 *
                 * @link https://php.net/manual/en/function.array-multisort.php
                 * @param &...$rest [optional] <p>
                 * More arrays, optionally followed by sort order and flags.
                 * Only elements corresponding to equivalent elements in previous arrays are compared.
                 * In other words, the sort is lexicographical.
                 * </p>
                 * @return bool true on success or false on failure.
                 */
                function array_multisort( $sort_order = SORT_ASC, $sort_flags = SORT_REGULAR,
                    &...$rest
                ): bool {}
            `,
		},
	}

	runScenarios(t, scenarios, func(v *phpversion.PHPVersion) ast.Visitor {
		return stubtransform.NewElementAvailableAttribute(v, nil)
	})
}
