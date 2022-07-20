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
		{
			file:        "/parameter.php",
			position:    &Position{Row: 7, Col: 13},
			outPosition: &Position{Row: 5, Col: 17},
		},
		{
			file:        "/function.php",
			position:    &Position{Row: 7, Col: 1},
			outPosition: &Position{Row: 3, Col: 1},
		},
		{
			file:     "/stdlib.php",
			position: &Position{Row: 3, Col: 6},
			outPosition: &Position{
				Row:  779,
				Col:  0,
				Path: "/Users/laytan/projects/elephp/phpstorm-stubs/standard/standard_8.php",
			},
		},
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

func BenchmarkStdlibFunction(b *testing.B) {
	is := is.New(b)
	project := NewProject("/Users/laytan/projects/elephp/fixtures/definitions")
	err := project.Parse()
	is.NoErr(err)

	for i := 0; i < b.N; i++ {
		_, err := project.Definition(
			"/Users/laytan/projects/elephp/fixtures/definitions/stdlib.php",
			&Position{Row: 3, Col: 6},
		)
		is.NoErr(err)
	}
}
