package position

import (
	"fmt"
	"io/ioutil"
	"testing"

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
		"/Users/laytan/projects/elephp/fixtures/definitions/parameter.php",
	)
	is.NoErr(err)

	for pos, loc := range expectations {
		t.Run(fmt.Sprint(pos), func(t *testing.T) {
			is := is.New(t)

			row, col := ToLocation(string(content), pos)
			fmt.Printf("Got row: %d, col: %d for pos %d\n", row, col, pos)
			is.Equal(row, loc[0])
			is.Equal(col, loc[1])
		})
	}

	for want, loc := range expectations {
		t.Run(fmt.Sprint(want), func(t *testing.T) {
			is := is.New(t)

			pos := FromLocation(string(content), loc[0], loc[1])
			is.Equal(pos, want)
		})
	}
}
