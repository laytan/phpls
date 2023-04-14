package server_test

import (
	"strings"
	"sync/atomic"
	"testing"

	"github.com/laytan/phpls/internal/server"
	"github.com/laytan/phpls/internal/wrkspc"
	"github.com/laytan/phpls/pkg/fqn"
	"github.com/laytan/phpls/pkg/parsing"
	"github.com/laytan/phpls/pkg/phpversion"
	"github.com/laytan/phpls/pkg/position"
	"github.com/laytan/phpls/pkg/strutil"
	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/lexer"
	"github.com/stretchr/testify/require"
)

func TestUseInsertion(t *testing.T) { // nolint:tparallel // doesn't work for some reason.
	t.Parallel()

	scenarios := []struct {
		name     string
		input    string
		expect   string
		insert   *fqn.FQN
		position *position.Position
	}{
		{
			name: "add after opening brace if no namespace",
			input: `
<?php

function test() {

}
		   			`,
			expect: `
<?php

use Test\Test;

function test() {

}
		   			`,
			insert: fqn.New("\\Test\\Test"),
			position: &position.Position{
				Row:  6,
				Col:  0,
				Path: "test",
			},
		},
		{
			name: "add after namespace",
			input: `
<?php

namespace Test;

function test() {
}
			`,
			expect: `
<?php

namespace Test;

use Test2\Test;

function test() {
}
			`,
			insert: fqn.New("\\Test2\\Test"),
			position: &position.Position{
				Row:  6,
				Col:  0,
				Path: "test",
			},
		},
		{
			name: "add after use statement without namespace",
			input: `
<?php

use Test\Test;

function test() {
}
			`,
			expect: `
<?php

use Test\Test;
use Test1\Test;

function test() {
}
			`,
			insert: fqn.New("\\Test1\\Test"),
			position: &position.Position{
				Row:  7,
				Col:  0,
				Path: "test",
			},
		},
		{
			name: "empty",
			input: `
<?php

			`,
			expect: `
<?php

use Test\Test;

			`,
			insert: fqn.New("\\Test\\Test"),
			position: &position.Position{
				Row:  3,
				Col:  0,
				Path: "test",
			},
		},
		{
			name: "multiple namespaces",
			input: `
<?php

namespace Test;

namespace Testing;

function test() {
}
			`,
			expect: `
<?php

namespace Test;

namespace Testing;

use Test1\Test;

function test() {
}
			`,
			insert: fqn.New("\\Test1\\Test"),
			position: &position.Position{
				Row:  9,
				Col:  0,
				Path: "test",
			},
		},
		{
			name: "multiple namespaces insert at middle",
			input: `
<?php

namespace Test;

function test() {
}

namespace Testing;

			`,
			expect: `
<?php

namespace Test;

use Test1\Test;

function test() {
}

namespace Testing;

			`,
			insert: fqn.New("\\Test1\\Test"),
			position: &position.Position{
				Row:  7,
				Col:  0,
				Path: "test",
			},
		},
		{
			name: "namespace block",
			input: `
<?php

namespace Testing;

namespace {

	function test() {
	}
}
			`,
			expect: `
<?php

namespace Testing;

namespace {

use Test\Test;

	function test() {
	}
}
			`,
			insert: fqn.New("\\Test\\Test"),
			position: &position.Position{
				Row:  8,
				Col:  0,
				Path: "test",
			},
		},
	}

	for _, scenario := range scenarios { // nolint:paralleltest // doesn't work for some reason.
		t.Run(scenario.name, func(t *testing.T) {
			parser := parsing.New(phpversion.EightOne())
			root, err := parser.Parse([]byte(scenario.input))
			require.NoError(t, err)

			w := mockWrkspc{
				path:    scenario.position.Path,
				content: scenario.input,
				root:    root,
				t:       t,
			}
			wrkspc.Current = &w

			edits := server.InsertUseStmt(scenario.insert, scenario.position)
			output := applyEdits(scenario.input, edits)
			require.Equal(t, scenario.expect, output)
		})
	}
}

func TestApplyEdits(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		name   string
		inp    string
		edits  []protocol.TextEdit
		output string
	}{
		{
			name: "adding to empty",
			inp:  "",
			edits: []protocol.TextEdit{{
				Range: protocol.Range{
					Start: protocol.Position{},
					End:   protocol.Position{},
				},
				NewText: "test",
			}},
			output: "test",
		},
		{
			name: "adding lines to empty",
			inp:  "",
			edits: []protocol.TextEdit{{
				Range: protocol.Range{
					Start: protocol.Position{},
					End:   protocol.Position{},
				},
				NewText: "hello\nworld",
			}},
			output: "hello\nworld",
		},
		{
			name: "adding to existing",
			inp:  "hello\nworld",
			edits: []protocol.TextEdit{{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      1,
						Character: 5,
					},
					End: protocol.Position{
						Line:      1,
						Character: 5,
					},
				},
				NewText: "!",
			}},
			output: "hello\nworld!",
		},
		{
			name: "remove everything",
			inp:  "hello, world!",
			edits: []protocol.TextEdit{{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      0,
						Character: 0,
					},
					End: protocol.Position{
						Line:      0,
						Character: 13,
					},
				},
				NewText: "",
			}},
			output: "",
		},
		{
			name: "change a character",
			inp:  "helxo",
			edits: []protocol.TextEdit{{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      0,
						Character: 3,
					},
					End: protocol.Position{
						Line:      0,
						Character: 4,
					},
				},
				NewText: "l",
			}},
			output: "hello",
		},
		{
			name: "delete multiple lines",
			inp:  "hello\nWorld!\nHow\nAre\nYah!",
			edits: []protocol.TextEdit{{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      0,
						Character: 1,
					},
					End: protocol.Position{
						Line:      2,
						Character: 3,
					},
				},
				NewText: "",
			}},
			output: "h\nAre\nYah!",
		},
		{
			name: "delete multiple lines and add some chars",
			inp:  "hello\nWorld!\nHow\nAre\nYah!",
			edits: []protocol.TextEdit{{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      0,
						Character: 1,
					},
					End: protocol.Position{
						Line:      2,
						Character: 3,
					},
				},
				NewText: "ayo",
			}},
			output: "hayo\nAre\nYah!",
		},
		{
			name: "change multiple lines",
			inp:  "hello\nWorld!\nHow\nAre\nYah!",
			edits: []protocol.TextEdit{{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      0,
						Character: 1,
					},
					End: protocol.Position{
						Line:      2,
						Character: 3,
					},
				},
				NewText: "ayo\nnasty",
			}},
			output: "hayo\nnasty\nAre\nYah!",
		},
		{
			name: "change multiple lines with end in between",
			inp:  "hello\nWorld!\nHow\nAre\nYah!",
			edits: []protocol.TextEdit{{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      1,
						Character: 5,
					},
					End: protocol.Position{
						Line:      3,
						Character: 1,
					},
				},
				NewText: "test",
			}},
			output: "hello\nWorldtestre\nYah!",
		},
		{
			name: "adding a use",
			inp: `<?php
function test() {

}`,
			edits: []protocol.TextEdit{{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      0,
						Character: 5,
					},
					End: protocol.Position{
						Line:      0,
						Character: 5,
					},
				},
				NewText: "\nuse Test;",
			}},
			output: `<?php
use Test;
function test() {

}`,
		},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()

			out := applyEdits(scenario.inp, scenario.edits)
			if out != scenario.output {
				t.Errorf("Result not as expected, got %v, want %v", out, scenario.output)
			}
		})
	}
}

// NOTE: we might need this for other lsp interactions too, like incremental file change instead of full if that is more performant.
func applyEdits(inp string, edits []protocol.TextEdit) string {
	if len(edits) == 0 {
		return inp
	}

	for _, edit := range edits {
		sr, sc, er, ec := edit.Range.Start.Line, edit.Range.Start.Character, edit.Range.End.Line, edit.Range.End.Character
		lines := strutil.Lines(inp)

		if sr == er {
			lines[sr] = lines[sr][0:sc] + edit.NewText + lines[sr][ec:]
			inp = strings.Join(lines, "\n")
			continue
		}

		newLines := strutil.Lines(edit.NewText)
		newLinesMiddle := newLines[1:]
		if len(newLines) > 2 {
			newLinesMiddle = newLines[1 : len(newLines)-2]
		}

		endLeft := lines[er][ec:]
		lines = append(lines[:sr+1], lines[er+1:]...)
		lines = append(lines[:sr+1], append(newLinesMiddle, lines[sr+1:]...)...)
		lines[sr] = lines[sr][:sc] + newLines[0] + endLeft

		inp = strings.Join(lines, "\n")
		continue
	}

	return inp
}

type mockWrkspc struct {
	path    string
	content string
	root    *ast.Root
	t       *testing.T
}

var _ wrkspc.Wrkspc = &mockWrkspc{}

func (*mockWrkspc) FLexerOf(path string) lexer.Lexer {
	panic("unimplemented")
}

func (*mockWrkspc) IsPhpFile(path string) bool {
	panic("unimplemented")
}

func (mockWrkspc) AllOf(path string) (string, *ast.Root, error) {
	panic("unimplemented")
}

func (mockWrkspc) ContentOf(path string) (string, error) {
	panic("unimplemented")
}

func (w mockWrkspc) FAllOf(path string) (string, *ast.Root) {
	if w.path != path {
		w.t.Errorf("expected wrkspc.FAllOf to be called with %s, but called with %s", w.path, path)
	}

	return w.content, w.root
}

func (mockWrkspc) FContentOf(path string) string {
	panic("unimplemented")
}

func (w mockWrkspc) FIROf(path string) *ast.Root {
	return w.root
}

func (mockWrkspc) IROf(path string) (*ast.Root, error) {
	panic("unimplemented")
}

func (mockWrkspc) Index(
	files chan<- *wrkspc.ParsedFile,
	total *atomic.Uint64,
	totalDone chan<- bool,
) error {
	panic("unimplemented")
}

func (mockWrkspc) Refresh(path string) error {
	panic("unimplemented")
}

func (mockWrkspc) RefreshFrom(path string, content string) error {
	panic("unimplemented")
}

func (mockWrkspc) Root() string {
	panic("unimplemented")
}

var _ wrkspc.Wrkspc = &mockWrkspc{}
