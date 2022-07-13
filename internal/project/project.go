package project

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/conf"
	"github.com/VKCOM/php-parser/pkg/errors"
	"github.com/VKCOM/php-parser/pkg/parser"
	"github.com/VKCOM/php-parser/pkg/version"
	"github.com/VKCOM/php-parser/pkg/visitor/dumper"
	"github.com/VKCOM/php-parser/pkg/visitor/traverser"
)

func NewProject(root string) project {
	return project{
		files: make(map[string]file),
		root:  root,
	}
}

type project struct {
	files map[string]file
	root  string
}

type file struct {
	ast      ast.Vertex
	modified time.Time
}

// TODO: get phpstorm stubs, parse them once, store them serialized, retrieve them when needed (already an ast) (no need to parse again).

func (p *project) parse() error {
	start := time.Now()

	config := conf.Config{
		ErrorHandlerFunc: func(e *errors.Error) {
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

func (p *project) definition(path string, row int, col int) error {
	file, ok := p.files[path]
	if !ok {
		// TODO: better
		panic("File not found")
	}

	goDumper := dumper.NewDumper(os.Stdout).WithTokens().WithPositions()
	file.ast.Accept(goDumper)

	definitionTraverser := NewDefinitionTraverser(row, col)
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

	if node, ok := definitionTraverser.getNode(); ok {
		fmt.Printf("%T\n", node)
	}

	return nil
}
