package typer

import (
	"strconv"
	"testing"

	"github.com/matryer/is"
)

func TestFQN(t *testing.T) {
	cases := []struct {
		FQN       string
		Name      string
		Namespace string
	}{
		{
			FQN:       "\\Testing\\One\\Two\\Three\\Four",
			Name:      "Four",
			Namespace: "Testing\\One\\Two\\Three",
		},
		{
			FQN:       "\\",
			Name:      "",
			Namespace: "",
		},
		{
			FQN:       "\\Test",
			Name:      "Test",
			Namespace: "",
		},
		{
			FQN:       "\\Test\\Test",
			Name:      "Test",
			Namespace: "Test",
		},
	}

	for i, test := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			is := is.New(t)

			f := NewFQN(test.FQN)
			is.Equal(f.String(), test.FQN)
			is.Equal(f.Name(), test.Name)
			is.Equal(f.Namespace(), test.Namespace)
		})
	}
}
