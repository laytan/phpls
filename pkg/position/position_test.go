package position

import (
	"fmt"
	"io/ioutil"
	"path"
	"testing"

	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/matryer/is"
)

func TestPosition(t *testing.T) {
	is := is.New(t)

	expectations := map[uint][]uint{
		74: {7, 5},
		35: {4, 1},
		46: {5, 1},
		61: {5, 16},
		66: {5, 21},
	}

	content, err := ioutil.ReadFile(
		path.Join(pathutils.Root(), "test", "testdata", "definitions", "parameter.php"),
	)
	is.NoErr(err)

	for pos, loc := range expectations {
		t.Run(fmt.Sprint(pos), func(t *testing.T) {
			is := is.New(t)

			row, col := PosToLoc(string(content), pos)
			is.Equal(row, loc[0])
			is.Equal(col, loc[1])
		})
	}

	for want, loc := range expectations {
		t.Run(fmt.Sprint(want), func(t *testing.T) {
			is := is.New(t)

			pos := LocToPos(string(content), loc[0], loc[1])
			is.Equal(pos, want)
		})
	}
}
