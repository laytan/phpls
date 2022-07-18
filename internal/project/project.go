package project

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irconv"
	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/conf"
	perrors "github.com/VKCOM/php-parser/pkg/errors"
	"github.com/VKCOM/php-parser/pkg/parser"
	"github.com/VKCOM/php-parser/pkg/version"
	"github.com/VKCOM/php-parser/pkg/visitor/traverser"
	"github.com/laytan/elephp/internal/traversers"
	"github.com/laytan/elephp/pkg/position"
)

func NewProject(root string) *Project {
	return &Project{
		files: make(map[string]file),
		root:  root,
	}
}

type Project struct {
	files map[string]file
	root  string
}

type file struct {
	ast      ast.Vertex
	content  string
	modified time.Time
}

// TODO: Change to uint32's
type Position struct {
	Row uint
	Col uint
}

// TODO: get phpstorm stubs, parse them once, store them serialized, retrieve them when needed (already an ast) (no need to parse again).
// We can add the stub directory to the root at first maybe, need to support multiple roots but that is no problem really.

// Definition needs to look at any accessible symbols
// Where accessible is:
// - used symbols
// - functions in the global scope
// - phpstorm stubs
// - if starts with '$this->' look at symbols in current class, and protected/public symbols in parent classes.
//
// Completion needs to look at everything definition does and also non-used symbols, also returning an edit to 'use'  the symbol.

func (p *Project) Parse() error {
	start := time.Now()

	config := conf.Config{
		ErrorHandlerFunc: func(e *perrors.Error) {
			panic(e)
		},
		// TODO: Get php version from 'php --version', or is there a builtin lsp way of getting language version?
		Version: &version.Version{Major: 8, Minor: 0},
	}

	filepath.Walk(p.root, func(path string, info fs.FileInfo, err error) error {
		// OPTIM: https://github.com/rjeczalik/notify to keep an eye on file changes and adds.

		// TODO: make configurable what a php file is.
		// TODO: does the lsp spec specify a way to get configured file types?
		if !strings.HasSuffix(path, ".php") || info.IsDir() {
			return nil
		}

		// If we currently have this parsed and the file hasn't changed, don't parse it again.
		if existing, ok := p.files[path]; ok {
			if !existing.modified.Before(info.ModTime()) {
				return nil
			}
		}

		content, err := ioutil.ReadFile(path)
		if err != nil {
			// TODO: don't panic willy nilly
			panic(err)
		}

		rootNode, err := parser.Parse(content, config)
		if err != nil {
			// TODO: don't panic, this is always a version not supported error
			panic(err)
		}

		p.files[path] = file{
			ast:      rootNode,
			content:  string(content),
			modified: info.ModTime(),
		}
		return nil
	})

	fmt.Printf("Parsed %d files in %s.\n", len(p.files), time.Since(start))
	return nil
}

func (p *Project) Definition(path string, pos *Position) (*Position, error) {
	file, ok := p.files[path]
	if !ok {
		// TODO: better
		panic("File not found " + path)
	}

	apos := position.FromLocation(file.content, pos.Row, pos.Col)
	nap := traversers.NewNodeAtPos(apos)
	irr := irconv.ConvertNode(file.ast)
	irr.Walk(nap)

	for _, node := range nap.Nodes {
		switch typedNode := node.(type) {
		case *ir.SimpleVar:
			assignment := p.assignment(path, typedNode)
			if assignment == nil {
				return nil, errors.New("No assignment found matching the variable at given position")
			}

			_, col := position.ToLocation(file.content, uint(assignment.Position.StartPos))

			return &Position{
				Row: uint(assignment.Position.StartLine),
				Col: col,
			}, nil
		}
	}

	return nil, errors.New("No definition found for given arguments")
}

func (p *Project) assignment(path string, variable *ir.SimpleVar) *ast.ExprAssign {
	// OPTIM: will in the future need to span multiple files, but lets be basic about this.

	file, ok := p.files[path]
	if !ok {
		panic("Not ok")
	}

	assignmentTraverser := traversers.NewAssignment(variable)
	traverser.NewTraverser(assignmentTraverser).Traverse(file.ast)

	return assignmentTraverser.Assignment
}
