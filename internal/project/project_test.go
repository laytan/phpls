package project

import (
	"errors"
	"io/fs"
	"io/ioutil"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"appliedgo.net/what"
	"github.com/laytan/elephp/internal/config"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/elephp/pkg/typer"
	"github.com/matryer/is"
	"github.com/samber/do"
)

var (
	stubsRoot     = path.Join(pathutils.Root(), "phpstorm-stubs")
	stdlibRoot    = path.Join(pathutils.Root(), "test", "testdata", "definitions", "stdlib")
	annotatedRoot = path.Join(pathutils.Root(), "test", "testdata", "definitions", "annotated")
	syntaxErrRoot = path.Join(pathutils.Root(), "test", "testdata", "syntaxerrors")
)

type stdlibScenario struct {
	in  *position.Position
	out *position.Position
}

func TestStdlibDefinitions(t *testing.T) {
	is := is.New(t)

	stdlibPath := path.Join(stdlibRoot, "stdlib.php")

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
				Path: path.Join(stubsRoot, "standard", "standard_8.php"),
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
				Path: path.Join(stubsRoot, "date", "date_c.php"),
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
				Path: path.Join(stubsRoot, "swoole", "Swoole", "WebSocket", "Server.php"),
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
				Path: path.Join(stubsRoot, "swoole", "Swoole", "Process.php"),
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
				Path: path.Join(stubsRoot, "date", "date_c.php"),
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
				Path: path.Join(stubsRoot, "http", "http3.php"),
			},
		},
		"param_function_call": {
			in: &position.Position{
				Row:  25,
				Col:  11,
				Path: stdlibPath,
			},
			out: &position.Position{
				Row:  891,
				Col:  1,
				Path: path.Join(stubsRoot, "standard", "standard_1.php"),
			},
		},
	}

	project := setup(stdlibRoot, phpversion.EightOne())
	err := project.ParseWithoutProgress()
	is.NoErr(err)

	for name, scenario := range scenarios {
		t.Run(name, func(t *testing.T) {
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

func TestAnnotatedDefinitions(t *testing.T) {
	is := is.New(t)

	project := setup(annotatedRoot, phpversion.EightOne())
	err := project.ParseWithoutProgress()
	is.NoErr(err)

	scenarios := aggregateAnnotations(t, annotatedRoot)
	for group, gscenarios := range scenarios {
		t.Run(group, func(t *testing.T) {
			for name, scenario := range gscenarios {
				t.Run(name, func(t *testing.T) {
					is := is.New(t)

					if scenario.in.Path == "" {
						t.Fatalf("invalid test scenario, no in called for '%s'", name)
					}

					if !scenario.isNoDef && scenario.out == nil {
						t.Fatalf("invalid test scenario, no out called for '%s'", name)
					}

					out, err := project.Definition(&scenario.in)

					if scenario.isNoDef {
						is.True(errors.Is(err, ErrNoDefinitionFound))
						return
					}

					is.NoErr(err)
					is.True(reflect.DeepEqual(out, scenario.out))
				})
			}
		})
	}
}

func setup(root string, phpv *phpversion.PHPVersion) *Project {
	do.OverrideValue(nil, config.Default())
	do.OverrideValue(nil, index.New(phpv))
	do.OverrideValue(nil, wrkspc.New(phpv, root))
	do.OverrideValue(nil, typer.New())

	return New()
}

// out is nil when isNoDef is true.
type annotedScenario struct {
	isNoDef bool
	in      position.Position
	out     *position.Position
}

var annotationRgx = regexp.MustCompile(`@t_(\w+)\(([\w\s]+), (\d+)\)`)

func aggregateAnnotations(t *testing.T, root string) map[string]map[string]*annotedScenario {
	is := is.New(t)

	scenarios := make(map[string]map[string]*annotedScenario)
	var scenarioLen uint
	aggrStart := time.Now()

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		content, err := ioutil.ReadFile(path)
		is.NoErr(err)

		strcontent := string(content)

		indexes := annotationRgx.FindAllStringIndex(strcontent, -1)
		matches := annotationRgx.FindAllStringSubmatch(strcontent, -1)
		is.Equal(len(indexes), len(matches))

		for i, match := range matches {
			is.True(len(match) > 3)

			row, _ := position.PosToLoc(strcontent, uint(indexes[i][0]))
			function := match[1]
			name := match[2]
			col := match[3]

			colint, err := strconv.Atoi(col)
			is.NoErr(err)

			group, name, ok := strings.Cut(name, "_")
			if !ok {
				name = group
				group = ""
			}

			g, ok := scenarios[group]
			if !ok {
				g = make(map[string]*annotedScenario)
				scenarios[group] = g
			}

			s, ok := g[name]
			if !ok {
				s = &annotedScenario{
					isNoDef: false,
					in:      position.Position{},
					out:     nil,
				}
				g[name] = s
				scenarioLen++
			}

			pos := position.Position{
				Row:  row,
				Col:  uint(colint),
				Path: path,
			}

			switch function {
			case "in":
				// Already had an int for this, so it's a naming collision.
				if s.in.Path != "" {
					t.Fatalf("naming collision, t_in is already set for test with name '%s'", name)
				}

				s.in = pos

			case "out":
				// Already had an out for this, so it's a naming collision.
				if s.out != nil {
					t.Fatalf("naming collision, t_out is already set for test with name '%s'", name)
				}

				s.out = &pos

			case "nodef":
				if ok {
					t.Fatalf("naming collision, there is already a test with the name: '%s'", name)
				}

				s.isNoDef = true
				s.in = pos

			default:
				// Invalid option so we ignore it.
				t.Logf("skipping %s, invalid function %s called", name, function)
				delete(g, name)
				scenarioLen--
			}
		}

		return nil
	})
	is.NoErr(err)

	t.Logf(
		"aggregated %d test scenarios from annotations in %s, running now",
		scenarioLen,
		time.Since(aggrStart),
	)

	return scenarios
}
