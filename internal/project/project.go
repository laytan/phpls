package project

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/conf"
	perrors "github.com/VKCOM/php-parser/pkg/errors"
	"github.com/VKCOM/php-parser/pkg/parser"
	"github.com/VKCOM/php-parser/pkg/version"
	"github.com/VKCOM/php-parser/pkg/visitor/dumper"
	"github.com/VKCOM/php-parser/pkg/visitor/traverser"
	"github.com/laytan/elephp/internal/traversers"
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
	modified time.Time
}

// TODO: Change to uint32's
type Position struct {
	Row int
	Col int
}

// TODO: get phpstorm stubs, parse them once, store them serialized, retrieve them when needed (already an ast) (no need to parse again).

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
		// OPTIM: we could use the modification time to invalidate a cache, resulting
		// OPTIM: in not having to parse it multiple times.

		// TODO: make configurable what a php file is.
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
			modified: info.ModTime(),
		}
		return nil
	})

	fmt.Printf("Parsed %d files in %s.\n", len(p.files), time.Since(start))
	return nil
}

// Definition
//
// 1. Parse doc for node that is at the given position.
//
// 2. Based on what it is, run another parser trying to find its definition. (in same file for now).
//
// x at cursor would search for y:
// x: 'ExprVariable' -> y: 'ExprAssign'

func (p *Project) Definition(path string, pos *Position) (*Position, error) {
	file, ok := p.files[path]
	if !ok {
		// TODO: better
		panic("File not found " + path)
	}

	goDumper := dumper.NewDumper(os.Stdout).WithTokens().WithPositions()
	file.ast.Accept(goDumper)

	definitionTraverser := NewDefinitionTraverser(pos.Row, pos.Col)
	traverser.NewTraverser(&definitionTraverser).Traverse(file.ast)

	for row, line := range definitionTraverser.Lines {
		fmt.Printf(
			"Line %d starts at %d and ends at %d with %d nodes.\n",
			row,
			line.startPos,
			line.endPos,
			len(line.nodes),
		)
	}

	if nodes, ok := definitionTraverser.getNode(); ok {
		for _, node := range nodes {
			fmt.Printf("%T\n", node)
		}

		// Walk up the nodes the given pos is in.
		// If we find something definable, try to define it.
		for _, node := range nodes {
			switch typedNode := node.(type) {
			case *ast.ExprVariable:
				assignment := p.assignment(path, typedNode)
				if assignment == nil {
					return nil, errors.New("No assignment found matching the variable at given position.")
				}

				return &Position{
					Row: assignment.Position.StartLine,
					Col: definitionTraverser.getColumn(assignment),
				}, nil
			}
		}
	}

	return nil, errors.New("No definition found for given arguments.")
}

func (p *Project) assignment(path string, variable *ast.ExprVariable) *ast.ExprAssign {
	// OPTIM: will in the future need to span multiple files, but lets be basic about this.

	file, ok := p.files[path]
	if !ok {
		panic("Not ok")
	}

	assignmentTraverser := traversers.NewAssignment(variable)
	traverser.NewTraverser(assignmentTraverser).Traverse(file.ast)

	return assignmentTraverser.Assignment
}
