package project

import (
	"fmt"
	"testing"

	"github.com/matryer/is"
)

type testDefinitionsInput struct {
	file        string
	position    *Position
	outPosition *Position
}

func TestDefinitions(t *testing.T) {
	is := is.New(t)

	expectations := []testDefinitionsInput{
		{
			file:        "/variable.php",
			position:    &Position{Row: 7, Col: 8},
			outPosition: &Position{Row: 5, Col: 1},
		},
		// {
		// 	file:        "/parameter.php",
		// 	position:    &Position{Row: 7, Col: 13},
		// 	outPosition: &Position{Row: 5, Col: 17},
		// },
	}

	project := NewProject("/Users/laytan/projects/elephp/fixtures/definitions")
	err := project.Parse()
	is.NoErr(err)

	for i, test := range expectations {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			is := is.New(t)

			pos, err := project.Definition(
				"/Users/laytan/projects/elephp/fixtures/definitions"+test.file,
				test.position,
			)
			is.NoErr(err)

			is.Equal(pos, test.outPosition)
		})
	}
}
