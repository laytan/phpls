package position_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/laytan/elephp/pkg/position"
	"github.com/matryer/is"
)

func TestPosition(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	expectations := map[uint][]uint{
		74: {7, 5},
		35: {4, 1},
		46: {5, 1},
		61: {5, 16},
		66: {5, 21},
	}

	// Adjusted positions because windows uses \r\n,
	// meaning each line will be 1 rune longer.
	if runtime.GOOS == "windows" {
		expectations = map[uint][]uint{
			80: {7, 5},
			38: {4, 1},
			50: {5, 1},
			65: {5, 16},
			70: {5, 21},
		}
	}

	content, err := os.ReadFile(
		filepath.Join(
			pathutils.Root(),
			"test",
			"testdata",
			"position",
			"position.php",
		),
	)
	is.NoErr(err)

	for pos, loc := range expectations {
		pos, loc := pos, loc
		t.Run(fmt.Sprint(pos), func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			row, col := position.PosToLoc(string(content), pos)
			is.Equal(row, loc[0])
			is.Equal(col, loc[1])
		})
	}

	for want, loc := range expectations {
		want, loc := want, loc
		t.Run(fmt.Sprint(want), func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			pos := position.LocToPos(string(content), loc[0], loc[1])
			is.Equal(pos, want)
		})
	}
}
