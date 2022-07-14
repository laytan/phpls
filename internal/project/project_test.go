package project

import (
	"fmt"
	"testing"
)

// func TestParsing(t *testing.T) {
// 	project := NewProject("/Users/laytan/projects/elephp/fixtures/test-project")
// 	err := project.parse()
// 	if err != nil {
// 		panic(err)
// 	}
// }

func TestDefinitions(t *testing.T) {
	project := NewProject("/Users/laytan/projects/elephp/fixtures/definitions")
	err := project.Parse()
	if err != nil {
		panic(err)
	}

	pos, err := project.Definition(
		"/Users/laytan/projects/elephp/fixtures/definitions/variable.php",
		&Position{Row: 7, Col: 8},
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Definition is at (%d, %d)\n", pos.Row, pos.Col)
}
