package annotated

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/laytan/elephp/pkg/position"
	"github.com/stretchr/testify/require"
)

// out is nil when isNoDef is true.
type AnnotedScenario struct {
	IsNoDef    bool
	IsDump     bool
	ShouldSkip bool
	In         position.Position
	Out        []*position.Position
}

var annotationRgx = regexp.MustCompile(`@t_(\w+)\(([\w\s]+), (\d+)\)`)

func Aggregate(t *testing.T, root string) map[string]map[string]*AnnotedScenario {
	t.Helper()

	scenarios := make(map[string]map[string]*AnnotedScenario)
	var scenarioLen uint
	aggrStart := time.Now()

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		content, rErr := os.ReadFile(path)
		require.NoError(t, rErr)

		strcontent := string(content)

		indexes := annotationRgx.FindAllStringIndex(strcontent, -1)
		if len(indexes) == 0 {
			return nil
		}

		matches := annotationRgx.FindAllStringSubmatch(strcontent, -1)
		require.Equal(t, len(indexes), len(matches), "indexes and matches must have equal length")

		for i, match := range matches {
			require.Greater(t, len(match), 3)

			row, _ := position.PosToLoc(strcontent, uint(indexes[i][0]))
			function := match[1]
			name := match[2]
			col := match[3]

			colint, err := strconv.Atoi(col)
			require.NoError(t, err)

			group, name, ok := strings.Cut(name, "_")
			if !ok {
				name = group
				group = ""
			}

			g, ok := scenarios[group]
			if !ok {
				g = make(map[string]*AnnotedScenario)
				scenarios[group] = g
			}

			s, ok := g[name]
			if !ok {
				s = &AnnotedScenario{
					IsNoDef: false,
					In:      position.Position{},
					Out:     []*position.Position{},
				}
				g[name] = s
				scenarioLen++
			}

			if strings.HasPrefix(function, "skip_") {
				s.ShouldSkip = true
				function = strings.TrimPrefix(function, "skip_")
			}

			pos := position.Position{
				Row:  row,
				Col:  uint(colint),
				Path: path,
			}

			switch function {
			case "in":
				// Already had an int for this, so it's a naming collision.
				if s.In.Path != "" {
					t.Fatalf("naming collision, t_in is already set for test with name '%s'", name)
				}

				s.In = pos

			case "out":
				s.Out = append(s.Out, &pos)

			case "nodef":
				if ok {
					t.Fatalf("naming collision, there is already a test with the name: '%s'", name)
				}

				s.IsNoDef = true
				s.In = pos

			case "dump":
				if ok {
					t.Fatalf("naming collision, there is already a test with the name: '%s'", name)
				}

				s.IsDump = true
				s.In = pos

			default:
				t.Fatalf("unsupported @t_ function: %s_%s", group, name)
			}
		}

		return nil
	})
	require.NoError(t, err)

	t.Logf(
		"aggregated %d test scenarios from annotations in %s, running now",
		scenarioLen,
		time.Since(aggrStart),
	)

	return scenarios
}
