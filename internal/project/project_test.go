package project

import (
	"errors"
	"fmt"
	"path"
	"sync/atomic"
	"testing"

	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/pkg/position"
	"github.com/matryer/is"
)

var (
	definitionsFolder = path.Join(pathutils.Root(), "fixtures", "definitions")
	stubsFolder       = path.Join(pathutils.Root(), "phpstorm-stubs")
)

type testDefinitionsInput struct {
	file        string
	position    *position.Position
	outPosition *position.Position
}

func TestDefinitions(t *testing.T) {
	is := is.New(t)

	expectations := []testDefinitionsInput{
		{
			file:        "variable.php",
			position:    &position.Position{Row: 7, Col: 8},
			outPosition: &position.Position{Row: 5, Col: 1},
		},
		{
			file:        "parameter.php",
			position:    &position.Position{Row: 7, Col: 13},
			outPosition: &position.Position{Row: 5, Col: 17},
		},
		{
			file:        "function.php",
			position:    &position.Position{Row: 7, Col: 1},
			outPosition: &position.Position{Row: 3, Col: 1},
		},
		{
			file:     "stdlib.php",
			position: &position.Position{Row: 3, Col: 6},
			outPosition: &position.Position{
				Row:  785,
				Col:  1,
				Path: path.Join(stubsFolder, "standard", "standard_8.php"),
			},
		},
		{
			file:     "global_var.php",
			position: &position.Position{Row: 1, Col: 1},
		},
		{
			file:        "global_var.php",
			position:    &position.Position{Row: 12, Col: 10},
			outPosition: &position.Position{Row: 11, Col: 12},
		},
		{
			file:        "global_var.php",
			position:    &position.Position{Row: 11, Col: 12},
			outPosition: &position.Position{Row: 2, Col: 1},
		},
		{
			file:        "function.php",
			position:    &position.Position{Row: 15, Col: 5},
			outPosition: &position.Position{Row: 11, Col: 5},
		},
		{
			file:     "function.php",
			position: &position.Position{Row: 18, Col: 1},
		},
		{
			file:     "class.php",
			position: &position.Position{Row: 12, Col: 19},
			outPosition: &position.Position{
				Row:  170,
				Col:  1,
				Path: path.Join(stubsFolder, "date", "date_c.php"),
			},
		},
		{
			file:        "class.php",
			position:    &position.Position{Row: 15, Col: 22},
			outPosition: &position.Position{Row: 8, Col: 1},
		},
		{
			file:     "class.php",
			position: &position.Position{Row: 17, Col: 20},
			outPosition: &position.Position{
				Row:  7,
				Col:  1,
				Path: path.Join(stubsFolder, "swoole", "Swoole", "WebSocket", "Server.php"),
			},
		},
		{
			file:     "class.php",
			position: &position.Position{Row: 19, Col: 16},
			outPosition: &position.Position{
				Row:  7,
				Col:  1,
				Path: path.Join(stubsFolder, "swoole", "Swoole", "Process.php"),
			},
		},
		{
			file:     path.Join("trait", "trait_user.php"),
			position: &position.Position{Row: 7, Col: 9},
			outPosition: &position.Position{
				Row:  3,
				Col:  1,
				Path: path.Join(definitionsFolder, "trait", "trait.php"),
			},
		},
		{
			file:     path.Join("trait", "trait_user.php"),
			position: &position.Position{Row: 8, Col: 9},
			outPosition: &position.Position{
				Row:  5,
				Col:  1,
				Path: path.Join(definitionsFolder, "trait", "trait_in_namespace.php"),
			},
		},
		{
			file:     path.Join("interface", "interface_user.php"),
			position: &position.Position{Row: 5, Col: 43},
			outPosition: &position.Position{
				Row:  3,
				Col:  1,
				Path: path.Join(definitionsFolder, "interface", "interface.php"),
			},
		},
		{
			file:     path.Join("interface", "interface_user.php"),
			position: &position.Position{Row: 9, Col: 46},
			outPosition: &position.Position{
				Row:  5,
				Col:  1,
				Path: path.Join(definitionsFolder, "interface", "interface_in_namespace.php"),
			},
		},
		{
			file:     path.Join("interface", "interface_user.php"),
			position: &position.Position{Row: 13, Col: 43},
			outPosition: &position.Position{
				Row:  13,
				Col:  1,
				Path: path.Join(stubsFolder, "date", "date_c.php"),
			},
		},
		{
			file:        path.Join("interface", "interface_user.php"),
			position:    &position.Position{Row: 21, Col: 43},
			outPosition: &position.Position{Row: 17, Col: 1},
		},
		{
			file:        path.Join("interface", "interface_user.php"),
			position:    &position.Position{Row: 25, Col: 45},
			outPosition: &position.Position{Row: 17, Col: 1},
		},
		{
			file:     path.Join("interface", "interface_user.php"),
			position: &position.Position{Row: 25, Col: 75},
			outPosition: &position.Position{
				Row:  13,
				Col:  1,
				Path: path.Join(stubsFolder, "date", "date_c.php"),
			},
		},
		{
			file:     "extends.php",
			position: &position.Position{Row: 5, Col: 32},
			outPosition: &position.Position{
				Row:  170,
				Col:  1,
				Path: path.Join(stubsFolder, "date", "date_c.php"),
			},
		},
		{
			file:     "extends.php",
			position: &position.Position{Row: 9, Col: 42},
			outPosition: &position.Position{
				Row:  7,
				Col:  1,
				Path: path.Join(stubsFolder, "swoole", "Swoole", "Client.php"),
			},
		},
		{
			file: "multiple_namespaces_in_one_file.php",
			position: &position.Position{
				Row: 7,
				Col: 47,
			},
			outPosition: &position.Position{
				Row:  713,
				Col:  1,
				Path: path.Join(stubsFolder, "http", "http3.php"),
			},
		},
		{
			file:     "use.php",
			position: &position.Position{Row: 3, Col: 5},
			outPosition: &position.Position{
				Row:  7,
				Col:  1,
				Path: path.Join(stubsFolder, "swoole", "Swoole", "Client.php"),
			},
		},
		{
			file:     "use.php",
			position: &position.Position{Row: 4, Col: 6},
			outPosition: &position.Position{
				Row:  170,
				Col:  1,
				Path: path.Join(stubsFolder, "date", "date_c.php"),
			},
		},
		{
			file:        path.Join("methods", "methods.php"),
			position:    &position.Position{Row: 28, Col: 10},
			outPosition: &position.Position{Row: 12, Col: 1},
		},
		{
			file:        path.Join("methods", "methods.php"),
			position:    &position.Position{Row: 28, Col: 16},
			outPosition: &position.Position{Row: 14, Col: 5},
		},
		{
			file:        path.Join("methods", "methods.php"),
			position:    &position.Position{Row: 29, Col: 16},
			outPosition: &position.Position{Row: 18, Col: 5},
		},
		{
			file:        path.Join("methods", "methods.php"),
			position:    &position.Position{Row: 30, Col: 16},
			outPosition: &position.Position{Row: 22, Col: 5},
		},
		{
			file:     path.Join("methods", "methods_child.php"),
			position: &position.Position{Row: 10, Col: 16},
			outPosition: &position.Position{
				Row:  18,
				Col:  5,
				Path: path.Join(definitionsFolder, "methods", "methods.php"),
			},
		},
		{
			file:        path.Join("methods", "methods_child.php"),
			position:    &position.Position{Row: 18, Col: 16},
			outPosition: nil,
		},
		{
			file:        path.Join("methods", "methods_child.php"),
			position:    &position.Position{Row: 12, Col: 16},
			outPosition: &position.Position{Row: 15, Col: 5},
		},
		{
			file:        path.Join("methods", "methods_trait.php"),
			position:    &position.Position{Row: 13, Col: 16},
			outPosition: &position.Position{Row: 7, Col: 5},
		},
		{
			file:     path.Join("methods", "methods.php"),
			position: &position.Position{Row: 41, Col: 16},
			outPosition: &position.Position{
				Row:  7,
				Col:  5,
				Path: path.Join(definitionsFolder, "methods", "methods_trait.php"),
			},
		},
		{
			file:     path.Join("methods", "methods.php"),
			position: &position.Position{Row: 42, Col: 16},
			outPosition: &position.Position{
				Row:  11,
				Col:  5,
				Path: path.Join(definitionsFolder, "methods", "methods_trait.php"),
			},
		},
		{
			file:     path.Join("methods", "methods.php"),
			position: &position.Position{Row: 43, Col: 16},
			outPosition: &position.Position{
				Row:  16,
				Col:  5,
				Path: path.Join(definitionsFolder, "methods", "methods_trait.php"),
			},
		},
		{
			file:     path.Join("methods", "methods_child.php"),
			position: &position.Position{Row: 9, Col: 16},
			outPosition: &position.Position{
				Row:  14,
				Col:  5,
				Path: path.Join(definitionsFolder, "methods", "methods.php"),
			},
		},
		{
			file:     path.Join("methods", "methods_child.php"),
			position: &position.Position{Row: 21, Col: 16},
			outPosition: &position.Position{
				Row:  16,
				Col:  5,
				Path: path.Join(definitionsFolder, "methods", "methods_trait.php"),
			},
		},
		{
			file:        path.Join("properties", "properties.php"),
			position:    &position.Position{Row: 13, Col: 17},
			outPosition: &position.Position{Row: 5, Col: 1},
		},
		{
			file:        path.Join("properties", "properties.php"),
			position:    &position.Position{Row: 13, Col: 23},
			outPosition: &position.Position{Row: 7, Col: 5},
		},
		{
			file:        path.Join("properties", "properties.php"),
			position:    &position.Position{Row: 18, Col: 50},
			outPosition: &position.Position{Row: 7, Col: 5},
		},
		{
			file:     path.Join("properties", "properties.php"),
			position: &position.Position{Row: 19, Col: 24},
		},
		{
			file:     path.Join("properties", "properties.php"),
			position: &position.Position{Row: 20, Col: 24},
		},
		{
			file:        path.Join("properties", "properties.php"),
			position:    &position.Position{Row: 26, Col: 18},
			outPosition: &position.Position{Row: 7, Col: 5},
		},
		{
			file:        path.Join("properties", "properties.php"),
			position:    &position.Position{Row: 32, Col: 21},
			outPosition: &position.Position{Row: 7, Col: 5},
		},
		{
			file:        path.Join("methods", "methods.php"),
			position:    &position.Position{Row: 56, Col: 21},
			outPosition: &position.Position{Row: 14, Col: 5},
		},
		{
			file:     path.Join("methods", "methods.php"),
			position: &position.Position{Row: 57, Col: 21},
		},
		{
			file:     path.Join("methods", "methods.php"),
			position: &position.Position{Row: 58, Col: 21},
		},
		{
			file:     path.Join("methods", "methods.php"),
			position: &position.Position{Row: 60, Col: 21},
			outPosition: &position.Position{
				Row:  16,
				Col:  5,
				Path: path.Join(definitionsFolder, "methods", "methods_trait.php"),
			},
		},
		{
			file:        path.Join("methods", "methods.php"),
			position:    &position.Position{Row: 66, Col: 18},
			outPosition: &position.Position{Row: 14, Col: 5},
		},
		{
			file:        path.Join("methods", "methods.php"),
			position:    &position.Position{Row: 72, Col: 21},
			outPosition: &position.Position{Row: 14, Col: 5},
		},
		{
			file:        "parameter.php",
			position:    &position.Position{Row: 13, Col: 10},
			outPosition: &position.Position{Row: 12, Col: 26},
		},
		{
			file:        "parameter.php",
			position:    &position.Position{Row: 17, Col: 10},
			outPosition: &position.Position{Row: 16, Col: 27},
		},
		{
			file:        "parameter.php",
			position:    &position.Position{Row: 20, Col: 6},
			outPosition: &position.Position{Row: 10, Col: 19},
		},
		{
			file:     "parameter.php",
			position: &position.Position{Row: 23, Col: 10},
			outPosition: &position.Position{
				Row:  891,
				Col:  1,
				Path: path.Join(stubsFolder, "standard", "standard_1.php"),
			},
		},
		{
			file:        path.Join("properties", "properties.php"),
			position:    &position.Position{Row: 57, Col: 31},
			outPosition: &position.Position{Row: 7, Col: 5},
		},
		{
			file:        path.Join("properties", "properties.php"),
			position:    &position.Position{Row: 58, Col: 30},
			outPosition: &position.Position{Row: 7, Col: 5},
		},
		{
			file:        path.Join("properties", "properties.php"),
			position:    &position.Position{Row: 59, Col: 37},
			outPosition: &position.Position{Row: 7, Col: 5},
		},
		{
			file:     path.Join("properties", "properties.php"),
			position: &position.Position{Row: 60, Col: 31},
		},
		{
			file:     path.Join("properties", "properties.php"),
			position: &position.Position{Row: 61, Col: 31},
		},
		{
			file:     path.Join("properties", "properties.php"),
			position: &position.Position{Row: 62, Col: 34},
		},
		{
			file:        path.Join("properties", "properties.php"),
			position:    &position.Position{Row: 69, Col: 26},
			outPosition: &position.Position{Row: 41, Col: 5},
		},
		{
			file:        path.Join("properties", "properties.php"),
			position:    &position.Position{Row: 69, Col: 36},
			outPosition: &position.Position{Row: 7, Col: 5},
		},
		{
			file:     path.Join("properties", "properties.php"),
			position: &position.Position{Row: 70, Col: 34},
		},
	}

	project := NewProject(definitionsFolder, phpversion.EightOne())
	err := project.Parse(&atomic.Uint32{})
	is.NoErr(err)

	for _, test := range expectations {
		t.Run(
			fmt.Sprintf("%s:%d,%d", test.file, test.position.Row, test.position.Col),
			func(t *testing.T) {
				is := is.New(t)

				testPath := path.Join(definitionsFolder, test.file)
				test.position.Path = testPath

				pos, err := project.Definition(test.position)

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
	project := NewProject(definitionsFolder, phpversion.EightOne())
	err := project.Parse(&atomic.Uint32{})
	is.NoErr(err)

	for i := 0; i < b.N; i++ {
		_, err := project.Definition(&position.Position{
			Row:  3,
			Col:  6,
			Path: path.Join(definitionsFolder, "stdlib.php"),
		})
		is.NoErr(err)
	}
}

func BenchmarkParsing(b *testing.B) {
	is := is.New(b)

	for i := 0; i < b.N; i++ {
		project := NewProject(definitionsFolder, phpversion.EightOne())
		err := project.Parse(&atomic.Uint32{})
		is.NoErr(err)
	}
}
