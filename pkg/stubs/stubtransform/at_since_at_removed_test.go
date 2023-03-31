package stubtransform_test

import (
	"testing"

	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/pkg/stubs/stubtransform"
	"github.com/laytan/php-parser/pkg/ast"
)

func TestAtSinceAtRemoved(t *testing.T) {
	t.Parallel()

	scenarios := []scenario{
		{
			name:    "@since removed on previous versions",
			version: "5.6",
			input: `
                <?php
                /**
                 * @since 7.4
                 */
                 function test() {}
            `,
			expected: `
                <?php
            `,
		},
		{
			name:    "@since keeps on current versions",
			version: "7.4",
			input: `
                <?php
                /**
                 * @since 7.4
                 */
                class Test {}
            `,
			expected: `
                <?php
                /**
                 * @since 7.4
                 */
                class Test {}
            `,
		},
		{
			name:    "@since keeps on later versions",
			version: "8.1",
			input: `
                <?php
                /**
                 * @since 7.4
                 */
                interface Test {}
            `,
			expected: `
                <?php
                /**
                 * @since 7.4
                 */
                interface Test {}
            `,
		},
		{
			name:    "@removed removed on later versions",
			version: "8.1",
			input: `
                <?php
                interface Test {

                    /**
                     * @removed 7.4
                     */
                    public function test() {}
                }
            `,
			expected: `
                <?php
                interface Test {
                }
            `,
		},
		{
			name:    "@removed removed on current versions",
			version: "7.4",
			input: `
                <?php
                class Test {
                    /**
                     * @removed 7.4
                     */
                    public string $test;
                }
            `,
			expected: `
                <?php
                class Test {
                }
            `,
		},
		{
			name:    "@removed keeps on previous versions",
			version: "5.6",
			input: `
                <?php
                class Test {
                    /**
                     * @removed 7.4
                     */
                    public string $test;
                }
            `,
			expected: `
                <?php
                class Test {
                    /**
                     * @removed 7.4
                     */
                    public string $test;
                }
            `,
		},
	}

	runScenarios(t, scenarios, func(p *phpversion.PHPVersion) ast.Visitor {
		return stubtransform.NewAtSinceAtRemoved(p, nil)
	})
}
