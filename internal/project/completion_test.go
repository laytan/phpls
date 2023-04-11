package project_test

import (
	"sync/atomic"
	"testing"

	"github.com/laytan/elephp/internal/project"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/conf"
	"github.com/laytan/php-parser/pkg/errors"
	"github.com/laytan/php-parser/pkg/lexer"
	"github.com/laytan/php-parser/pkg/version"
)

func TestCompletionQuery(t *testing.T) {
	t.Parallel()

	input := `
<?php

new namespace\
	`

	wrkspc.Current = &mockWrkspc{
		input: []byte(input),
		t:     t,
	}

	project.GetCompletionQuery2(&position.Position{
		Row:  4,
		Col:  14, // is this 0 or 1 based?
		Path: "test",
	})
}

type mockWrkspc struct {
	input []byte
	t     *testing.T
}

// AllOf implements wrkspc.Wrkspc
func (*mockWrkspc) AllOf(path string) (string, *ast.Root, error) {
	panic("unimplemented")
}

// ContentOf implements wrkspc.Wrkspc
func (*mockWrkspc) ContentOf(path string) (string, error) {
	panic("unimplemented")
}

// FAllOf implements wrkspc.Wrkspc
func (*mockWrkspc) FAllOf(path string) (string, *ast.Root) {
	panic("unimplemented")
}

// FContentOf implements wrkspc.Wrkspc
func (*mockWrkspc) FContentOf(path string) string {
	panic("unimplemented")
}

// FIROf implements wrkspc.Wrkspc
func (*mockWrkspc) FIROf(path string) *ast.Root {
	panic("unimplemented")
}

// FLexerOf implements wrkspc.Wrkspc
func (w *mockWrkspc) FLexerOf(path string) lexer.Lexer {
	l, err := lexer.New(w.input, conf.Config{
		Version: &version.Version{
			Major: 8,
			Minor: 2,
		},
		ErrorHandlerFunc: func(e *errors.Error) {
			w.t.Fatal(e)
		},
	})
	if err != nil {
		w.t.Fatal(err)
	}

	return l
}

// IROf implements wrkspc.Wrkspc
func (*mockWrkspc) IROf(path string) (*ast.Root, error) {
	panic("unimplemented")
}

// Index implements wrkspc.Wrkspc
func (*mockWrkspc) Index(
	files chan<- *wrkspc.ParsedFile,
	total *atomic.Uint64,
	totalDone chan<- bool,
) error {
	panic("unimplemented")
}

// Refresh implements wrkspc.Wrkspc
func (*mockWrkspc) Refresh(path string) error {
	panic("unimplemented")
}

// RefreshFrom implements wrkspc.Wrkspc
func (*mockWrkspc) RefreshFrom(path string, content string) error {
	panic("unimplemented")
}

// Root implements wrkspc.Wrkspc
func (*mockWrkspc) Root() string {
	panic("unimplemented")
}

func (*mockWrkspc) IsPhpFile(string) bool {
	return true
}

var _ wrkspc.Wrkspc = &mockWrkspc{}
