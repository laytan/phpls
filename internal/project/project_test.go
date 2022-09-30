package project_test

import (
	"errors"
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
	"github.com/laytan/elephp/pkg/typer"
	"github.com/matryer/is"
	"github.com/samber/do"
)

var (
	stubsRoot     = filepath.Join(pathutils.Root(), "third_party", "phpstorm-stubs")
	stdlibRoot    = filepath.Join(pathutils.Root(), "test", "testdata", "definitions", "stdlib")
	annotatedRoot = filepath.Join(pathutils.Root(), "test", "testdata", "definitions", "annotated")
	syntaxErrRoot = filepath.Join(pathutils.Root(), "test", "testdata", "syntaxerrors")
)

type stdlibScenario struct {
	in  *position.Position
	out *position.Position
}

//nolint:paralleltest,tparallel // Causes data race (indexing while testing?)
func TestStdlibDefinitions(t *testing.T) {
	is := is.New(t)

	stdlibPath := filepath.Join(stdlibRoot, "stdlib.php")

	scenarios := map[string]*stdlibScenario{
		// TODO: fix & uncomment.
		// "use": {
		// 	in: &position.Position{
		// 		Row:  3,
		// 		Col:  5,
		// 		Path: stdlibPath,
		// 	},
		// 	out: &position.Position{
		// 		Row:  7,
		// 		Col:  1,
		// 		Path: path.Join(stubsRoot, "swoole", "Swoole", "WebSocket", "Server.php"),
		// 	},
		// },
		// "use_alias": {
		// 	in: &position.Position{
		// 		Row:  4,
		// 		Col:  23,
		// 		Path: stdlibPath,
		// 	},
		// 	out: &position.Position{
		// 		Row:  7,
		// 		Col:  1,
		// 		Path: path.Join(stubsRoot, "swoole", "Swoole", "Process.php"),
		// 	},
		// },
		"in_array": {
			in: &position.Position{
				Row:  10,
				Col:  6,
				Path: stdlibPath,
			},
			out: &position.Position{
				Row:  785,
				Col:  1,
				Path: filepath.Join(stubsRoot, "standard", "standard_8.php"),
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
				Path: filepath.Join(stubsRoot, "date", "date_c.php"),
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
				Path: filepath.Join(stubsRoot, "swoole", "Swoole", "WebSocket", "Server.php"),
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
				Path: filepath.Join(stubsRoot, "swoole", "Swoole", "Process.php"),
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
				Path: filepath.Join(stubsRoot, "date", "date_c.php"),
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
				Path: filepath.Join(stubsRoot, "http", "http3.php"),
			},
		},
		"param_function_call": {
			in: &position.Position{
				Row:  25,
				Col:  11,
				Path: stdlibPath,
			},
			out: &position.Position{
				Row:  896,
				Col:  1,
				Path: filepath.Join(stubsRoot, "standard", "standard_1.php"),
			},
		},
	}

	project := setup(stdlibRoot, phpversion.EightOne())
	err := project.ParseWithoutProgress()
	is.NoErr(err)

	for name, scenario := range scenarios {
		name, scenario := name, scenario
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			out, err := project.Definition(scenario.in)
			is.NoErr(err)

			if !reflect.DeepEqual(out, scenario.out) {
				what.Is(out)
				what.Is(scenario.out)
				t.Errorf("definitions don't match, run with `-tags what` to debug")
			}
		})
	}
}

//nolint:paralleltest,tparallel // Causes data race (indexing while testing?)
func TestParserPanicIsRecovered(t *testing.T) {
	is := is.New(t)

	project := setup(
		syntaxErrRoot,
		&phpversion.PHPVersion{
			Major: 7,
			Minor: 4,
		},
	)

	err := project.ParseWithoutProgress()
	is.NoErr(err)
}

//nolint:paralleltest,tparallel // Causes data race (indexing while testing?)
func TestAnnotatedDefinitions(t *testing.T) {
	is := is.New(t)

	proj := setup(annotatedRoot, phpversion.EightOne())
	err := proj.ParseWithoutProgress()
	is.NoErr(err)

	scenarios := annotated.Aggregate(t, annotatedRoot)
	for group, gscenarios := range scenarios {
		group, gscenarios := group, gscenarios
		t.Run(group, func(t *testing.T) {
			t.Parallel()
			for name, scenario := range gscenarios {
				name, scenario := name, scenario
				t.Run(name, func(t *testing.T) {
					t.Parallel()
					is := is.New(t)

					if scenario.ShouldSkip {
						t.SkipNow()
					}

					if scenario.In.Path == "" {
						t.Fatalf("invalid test scenario, no in called for '%s'", name)
					}

					if scenario.IsDump {
						root, err := wrkspc.FromContainer().IROf(scenario.In.Path)
						is.NoErr(err)
						what.Is(root)
						return
					}

					if !scenario.IsNoDef && scenario.Out == nil {
						t.Fatalf("invalid test scenario, no out called for '%s'", name)
					}

					out, err := proj.Definition(&scenario.In)

					if scenario.IsNoDef {
						is.True(errors.Is(err, project.ErrNoDefinitionFound))
						return
					}

					is.NoErr(err)
					is.True(reflect.DeepEqual(out, scenario.Out))
				})
			}
		})
	}
}

func setup(root string, phpv *phpversion.PHPVersion) *project.Project {
	do.OverrideValue(nil, config.Default())
	do.OverrideValue(nil, index.New(phpv))
	do.OverrideValue(nil, wrkspc.New(phpv, root))
	do.OverrideValue(nil, typer.New())

	return project.New()
}
