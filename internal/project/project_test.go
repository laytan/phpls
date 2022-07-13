package project

import (
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
	err := project.parse()
	if err != nil {
		panic(err)
	}

	project.definition(
		"/Users/laytan/projects/elephp/fixtures/definitions/variable.php",
		5,
		8,
	)
}
