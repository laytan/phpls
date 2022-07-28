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
		{
			file:     "class.php",
			position: &Position{Row: 12, Col: 19},
			outPosition: &Position{
				Row:  165,
				Col:  1,
				Path: path.Join(stubsFolder, "date", "date_c.php"),
			},
		},
		{
			file:        "class.php",
			position:    &Position{Row: 15, Col: 22},
			outPosition: &Position{Row: 8, Col: 1},
		},
		{
			file:     "class.php",
			position: &Position{Row: 17, Col: 20},
			outPosition: &Position{
				Row:  7,
				Col:  1,
				Path: path.Join(stubsFolder, "swoole", "Swoole", "WebSocket", "Server.php"),
			},
		},
		{
			file:     "class.php",
			position: &Position{Row: 19, Col: 16},
			outPosition: &Position{
				Row:  7,
				Col:  1,
				Path: path.Join(stubsFolder, "swoole", "Swoole", "Process.php"),
			},
		},
		{
			file:     path.Join("trait", "trait_user.php"),
			position: &Position{Row: 7, Col: 9},
			outPosition: &Position{
				Row:  3,
				Col:  1,
				Path: path.Join(definitionsFolder, "trait", "trait.php"),
			},
		},
		{
			file:     path.Join("trait", "trait_user.php"),
			position: &Position{Row: 8, Col: 9},
			outPosition: &Position{
				Row:  5,
				Col:  1,
				Path: path.Join(definitionsFolder, "trait", "trait_in_namespace.php"),
			},
		},
		{
			file:     path.Join("interface", "interface_user.php"),
			position: &Position{Row: 5, Col: 43},
			outPosition: &Position{
				Row:  3,
				Col:  1,
				Path: path.Join(definitionsFolder, "interface", "interface.php"),
			},
		},
		{
			file:     path.Join("interface", "interface_user.php"),
			position: &Position{Row: 9, Col: 46},
			outPosition: &Position{
				Row:  5,
				Col:  1,
				Path: path.Join(definitionsFolder, "interface", "interface_in_namespace.php"),
			},
		},
		{
			file:     path.Join("interface", "interface_user.php"),
			position: &Position{Row: 13, Col: 43},
			outPosition: &Position{
				Row:  13,
				Col:  1,
				Path: path.Join(stubsFolder, "date", "date_c.php"),
			},
		},
		{
			file:        path.Join("interface", "interface_user.php"),
			position:    &Position{Row: 21, Col: 43},
			outPosition: &Position{Row: 17, Col: 1},
		},
		{
			file:        path.Join("interface", "interface_user.php"),
			position:    &Position{Row: 25, Col: 45},
			outPosition: &Position{Row: 17, Col: 1},
		},
		{
			file:     path.Join("interface", "interface_user.php"),
			position: &Position{Row: 25, Col: 75},
			outPosition: &Position{
				Row:  13,
				Col:  1,
				Path: path.Join(stubsFolder, "date", "date_c.php"),
			},
		},
		{
			file:     "extends.php",
			position: &Position{Row: 5, Col: 32},
			outPosition: &Position{
				Row:  165,
				Col:  1,
				Path: path.Join(stubsFolder, "date", "date_c.php"),
			},
		},
		{
			file:     "extends.php",
			position: &Position{Row: 9, Col: 42},
			outPosition: &Position{
				Row:  7,
				Col:  1,
				Path: path.Join(stubsFolder, "swoole", "Swoole", "Client.php"),
			},
		},
		{
			file: "multiple_namespaces_in_one_file.php",
			position: &Position{
				Row: 7,
				Col: 47,
			},
			outPosition: &Position{
				Row:  713,
				Col:  1,
				Path: path.Join(stubsFolder, "http", "http3.php"),
			},
		},
		{
			file:     "use.php",
			position: &Position{Row: 3, Col: 5},
			outPosition: &Position{
				Row:  7,
				Col:  1,
				Path: path.Join(stubsFolder, "swoole", "Swoole", "Client.php"),
			},
		},
		{
			file:     "use.php",
			position: &Position{Row: 4, Col: 6},
			outPosition: &Position{
				Row:  165,
				Col:  1,
				Path: path.Join(stubsFolder, "date", "date_c.php"),
			},
		},
	}

	project := NewProject(definitionsFolder)
	err := project.Parse()
	is.NoErr(err)

	for _, test := range expectations {
		t.Run(
			fmt.Sprintf("%s:%d,%d", test.file, test.position.Row, test.position.Col),
			func(t *testing.T) {
				is := is.New(t)

				testPath := path.Join(definitionsFolder, test.file)

				pos, err := project.Definition(
					testPath,
					test.position,
				)

				// Error is expected when no out position is given.
				if test.outPosition == nil {
					is.True(errors.Is(err, ErrNoDefinitionFound))
				} else {
					is.NoErr(err)
				}

				// If test out position has no path, assume the same path as the input.
				if test.outPosition != nil && test.outPosition.Path == "" && pos.Path != "" {
					is.Equal(testPath, pos.Path)
					is.Equal(test.outPosition.Col, pos.Col)
					is.Equal(test.outPosition.Row, pos.Row)
				} else {
					is.Equal(pos, test.outPosition)
				}
			},
		)
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

func BenchmarkParsing(b *testing.B) {
	is := is.New(b)

	for i := 0; i < b.N; i++ {
		project := NewProject(definitionsFolder)
		err := project.Parse()
		is.NoErr(err)
	}
}
