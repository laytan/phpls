package project_test

import (
	"path/filepath"
	"reflect"
	"testing"

	"appliedgo.net/what"
	"github.com/laytan/elephp/internal/config"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/project"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/annotated"
	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/pkg/position"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

var (
	stubsDir   = filepath.Join(pathutils.Root(), "third_party", "phpstorm-stubs")
	stdlibRoot = filepath.Join(
		pathutils.Root(),
		"internal",
		"project",
		"testdata",
		"definitions",
		"stdlib",
	)
	annotatedRoot = filepath.Join(
		pathutils.Root(),
		"internal",
		"project",
		"testdata",
		"definitions",
		"annotated",
	)
	syntaxErrRoot = filepath.Join(
		pathutils.Root(),
		"internal",
		"project",
		"testdata",
		"syntaxerrors",
	)
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(
		m,
		// The cache size logger.
		goleak.IgnoreTopFunction("github.com/laytan/elephp/internal/wrkspc.New.func1"),
	)
}

type stdlibScenario struct {
	in  *position.Position
	out *position.Position
}

//nolint:paralleltest,tparallel // Causes data race (indexing while testing?)
func TestStdlibDefinitions(t *testing.T) {
	stdlibPath := filepath.Join(stdlibRoot, "stdlib.php")

	scenarios := map[string]*stdlibScenario{
		"in_array": {
			in: &position.Position{
				Row:  10,
				Col:  6,
				Path: stdlibPath,
			},
			out: &position.Position{
				Row:  797,
				Col:  1,
				Path: filepath.Join(stubsDir, "standard", "standard_8.php"),
			},
		},
		"fqn": {
			in: &position.Position{
				Row:  12,
				Col:  20,
				Path: stdlibPath,
			},
			out: &position.Position{
				Row:  170,
				Col:  1,
				Path: filepath.Join(stubsDir, "date", "date_c.php"),
			},
		},
		"name": {
			in: &position.Position{
				Row:  15,
				Col:  15,
				Path: stdlibPath,
			},
			out: &position.Position{
				Row:  7,
				Col:  1,
				Path: filepath.Join(stubsDir, "swoole", "Swoole", "WebSocket", "Server.php"),
			},
		},
		"name_alias": {
			in: &position.Position{
				Row:  17,
				Col:  16,
				Path: stdlibPath,
			},
			out: &position.Position{
				Row:  7,
				Col:  1,
				Path: filepath.Join(stubsDir, "swoole", "Swoole", "Process.php"),
			},
		},
		"implements_global": {
			in: &position.Position{
				Row:  19,
				Col:  41,
				Path: stdlibPath,
			},
			out: &position.Position{
				Row:  13,
				Col:  1,
				Path: filepath.Join(stubsDir, "date", "date_c.php"),
			},
		},
		"extends_multiple_namespaces_in_one_file": {
			in: &position.Position{
				Row:  21,
				Col:  47,
				Path: stdlibPath,
			},
			out: &position.Position{
				Row:  713,
				Col:  1,
				Path: filepath.Join(stubsDir, "http", "http3.php"),
			},
		},
		"param_function_call": {
			in: &position.Position{
				Row:  25,
				Col:  11,
				Path: stdlibPath,
			},
			out: &position.Position{
				Row:  897,
				Col:  1,
				Path: filepath.Join(stubsDir, "standard", "standard_1.php"),
			},
		},
		"constant_fetch": {
			in: &position.Position{
				Row:  27,
				Col:  6,
				Path: stdlibPath,
			},
			out: &position.Position{
				Row:  209,
				Col:  1,
				Path: filepath.Join(stubsDir, "Core", "Core_d.php"),
			},
		},
	}

	project := setup(stdlibRoot, phpversion.EightOne())
	err := project.ParseWithoutProgress()
	require.NoError(t, err)

	for name, scenario := range scenarios {
		name, scenario := name, scenario
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			out, err := project.Definition(scenario.in)
			require.NoError(t, err)
			require.Len(t, out, 1)

			if !reflect.DeepEqual(out[0], scenario.out) {
				what.Is(out)
				what.Is(scenario.out)
				t.Errorf("definitions don't match, run with `-tags what` to debug")
			}
		})
	}
}

//nolint:paralleltest,tparallel // Causes data race (indexing while testing?)
func TestParserPanicIsRecovered(t *testing.T) {
	project := setup(
		syntaxErrRoot,
		&phpversion.PHPVersion{
			Major: 7,
			Minor: 4,
		},
	)

	err := project.ParseWithoutProgress()
	require.NoError(t, err)
}

//nolint:paralleltest,tparallel // Causes data race (indexing while testing?)
func TestAnnotatedDefinitions(t *testing.T) {
	proj := setup(annotatedRoot, phpversion.EightOne())
	err := proj.ParseWithoutProgress()
	require.NoError(t, err)

	scenarios := annotated.Aggregate(t, annotatedRoot)
	for group, gscenarios := range scenarios {
		group, gscenarios := group, gscenarios
		t.Run(group, func(t *testing.T) {
			t.Parallel()

			for name, scenario := range gscenarios {
				name, scenario := name, scenario
				t.Run(name, func(t *testing.T) {
					t.Parallel()

					if scenario.ShouldSkip {
						t.SkipNow()
					}

					if scenario.In.Path == "" {
						t.Fatalf("invalid test scenario, no in called for '%s'", name)
					}

					if scenario.IsDump {
						root, err := wrkspc.Current.IROf(scenario.In.Path)
						require.NoError(t, err)
						what.Is(root)
						return
					}

					if !scenario.IsNoDef && len(scenario.Out) == 0 {
						t.Fatalf("invalid test scenario, no out called for '%s'", name)
					}

					out, err := proj.Definition(&scenario.In)

					if scenario.IsNoDef {
						require.ErrorIs(t, err, project.ErrNoDefinitionFound)
						return
					}

					require.NoError(t, err)
					require.Equal(t, out, scenario.Out)
				})
			}
		})
	}
}

func setup(root string, phpv *phpversion.PHPVersion) *project.Project {
	config.Current = config.Default()
	index.Current = index.New(phpv)
	wrkspc.Current = wrkspc.New(phpv, root, stubsDir)

	return project.New()
}
