package throws_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"appliedgo.net/what"
	"github.com/laytan/elephp/internal/config"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/fqner"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/project"
	"github.com/laytan/elephp/internal/throws"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/annotated"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/functional"
	"github.com/laytan/elephp/pkg/nodescopes"
	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/samber/do"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(
		m,
		// The cache size logger.
		goleak.IgnoreTopFunction("github.com/laytan/elephp/internal/wrkspc.New.func1"),
	)
}

func TestAnnotateThrows(t *testing.T) {
	t.SkipNow()
	t.Parallel()

	root := filepath.Join(pathutils.Root(), "internal", "throws", "testdata")

	err := setup(root, phpversion.EightOne())
	require.NoError(t, err)

	scenarios := annotated.Aggregate(t, root)
	for categoryName, category := range scenarios {
		categoryName, category := categoryName, category
		t.Run(categoryName, func(t *testing.T) {
			t.Parallel()

			for name, scenario := range category {
				name, scenario := name, scenario
				t.Run(name, func(t *testing.T) {
					t.Parallel()

					if scenario.ShouldSkip {
						t.SkipNow()
					}

					if scenario.IsDump {
						root, err := wrkspc.FromContainer().IROf(scenario.In.Path)
						require.NoError(t, err)
						what.Is(root)
						return
					}

					if !scenario.IsNoDef && len(scenario.Out) == 0 {
						t.Error("invalid test scenario, no out called")
					}

					ctx, err := context.New(&scenario.In)
					require.NoError(t, err)

					// Get the first method or function in the context.
					var scope ast.Vertex
					var scopeKind ast.Type
					for advanced := true; advanced; advanced = ctx.Advance() {
						scope = ctx.Current()
						scopeKind = scope.GetType()
						if scopeKind == ast.TypeStmtClassMethod ||
							scopeKind == ast.TypeStmtFunction {
							break
						}
					}

					if scopeKind != ast.TypeStmtClassMethod && scopeKind != ast.TypeStmtFunction {
						t.Errorf("in node is not a method or function: %v", ctx)
					}

					rooter := wrkspc.NewRooter(scenario.In.Path)
					resolver := throws.NewResolver(rooter, scope)
					thrown := resolver.Throws()

					expectedThrown := functional.Map(
						scenario.Out,
						func(pos *position.Position) *fqn.FQN {
							ctx, err := context.New(pos)
							require.NoError(t, err)

							cls := ctx.Current()
							if !nodescopes.IsClassLike(cls.GetType()) {
								t.Errorf("out node is not a class like node %v", cls)
							}

							root, err := wrkspc.FromContainer().IROf(pos.Path)
							require.NoError(t, err)

							fqn := fqner.FullyQualifyName(
								root,
								cls.(*ast.StmtClass).Name.(*ast.Name),
							)

							return fqn
						},
					)

					if scenario.IsNoDef {
						if len(thrown) > 0 {
							t.Errorf("expected no throws, got %v", thrown)
						}

						return
					}

					if len(thrown) != len(expectedThrown) {
						t.Errorf("throws %v but expected to throw %v", thrown, expectedThrown)
					}

				Thrown:
					for _, th := range thrown {
						for _, tht := range expectedThrown {
							if th.String() == tht.String() {
								break Thrown
							}
						}

						t.Errorf("throws %v, but the thrown FQN %v was not part of the expected throws %v", thrown, th, expectedThrown)
					}
				})
			}
		})
	}
}

func setup(root string, phpv *phpversion.PHPVersion) error {
	i := index.New(phpv)
	do.OverrideValue(nil, config.Default())
	do.OverrideValue(nil, i)
	do.OverrideValue(
		nil,
		wrkspc.New(phpv, root, filepath.Join(pathutils.Root(), "third_party", "phpstorm-stubs")),
	)

	p := project.New()
	if err := p.ParseWithoutProgress(); err != nil {
		return fmt.Errorf("[throws_test.setup]: %w", err)
	}

	return nil
}
