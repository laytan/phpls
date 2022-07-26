package project

import (
	"errors"
	"fmt"
	"path"
	"testing"

	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/matryer/is"
)

var (
	definitionsFolder = path.Join(pathutils.Root(), "fixtures", "definitions")
	stubsFolder       = path.Join(pathutils.Root(), "phpstorm-stubs")
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
			file:        "variable.php",
			position:    &Position{Row: 7, Col: 8},
			outPosition: &Position{Row: 5, Col: 1},
		},
		{
			file:        "parameter.php",
			position:    &Position{Row: 7, Col: 13},
			outPosition: &Position{Row: 5, Col: 17},
		},
		{
			file:        "function.php",
			position:    &Position{Row: 7, Col: 1},
			outPosition: &Position{Row: 3, Col: 1},
		},
		{
			file:     "stdlib.php",
			position: &Position{Row: 3, Col: 6},
			outPosition: &Position{
				Row:  779,
				Col:  1,
				Path: path.Join(stubsFolder, "standard", "standard_8.php"),
			},
		},
		{
			file:     "global_var.php",
			position: &Position{Row: 1, Col: 1},
		},
		{
			file:        "global_var.php",
			position:    &Position{Row: 12, Col: 10},
			outPosition: &Position{Row: 11, Col: 12},
		},
		{
			file:        "global_var.php",
			position:    &Position{Row: 11, Col: 12},
			outPosition: &Position{Row: 2, Col: 1},
		},
		{
			file:        "function.php",
			position:    &Position{Row: 15, Col: 5},
			outPosition: &Position{Row: 11, Col: 5},
		},
		{
			file:     "function.php",
			position: &Position{Row: 18, Col: 1},
		},
	}

	project := NewProject(definitionsFolder)
	err := project.Parse()
	is.NoErr(err)

	for i, test := range expectations {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			is := is.New(t)

			pos, err := project.Definition(
				path.Join(definitionsFolder, test.file),
				test.position,
			)

			// Error is expected when no out position is given.
			if test.outPosition == nil {
				is.True(errors.Is(err, ErrNoDefinitionFound))
			} else {
				is.NoErr(err)
			}

			is.Equal(pos, test.outPosition)
		})
	}
}

func BenchmarkStdlibFunction(b *testing.B) {
	is := is.New(b)
	project := NewProject(definitionsFolder)
	err := project.Parse()
	is.NoErr(err)

	for i := 0; i < b.N; i++ {
		_, err := project.Definition(
			path.Join(definitionsFolder, "stdlib.php"),
			&Position{Row: 3, Col: 6},
		)
		is.NoErr(err)
	}
}
