package project

import (
	"errors"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irconv"
	"github.com/VKCOM/php-parser/pkg/conf"
	perrors "github.com/VKCOM/php-parser/pkg/errors"
	"github.com/VKCOM/php-parser/pkg/parser"
	"github.com/VKCOM/php-parser/pkg/version"
	"github.com/laytan/elephp/internal/traversers"
	"github.com/laytan/elephp/pkg/position"
	log "github.com/sirupsen/logrus"
)

func NewProject(root string) *Project {
	// OPTIM: parse phpstorm stubs once and store on filesystem or even in the binary because they won't change.
	//
	// OPTIM: This is also parsing tests for the specific stubs,
	// OPTIM: and has specific files for different php versions that should be excluded/handled appropriately.
	stubs := "/Users/laytan/projects/elephp/phpstorm-stubs"

	roots := []string{root, stubs}
	return &Project{
		files: make(map[string]file),
		roots: roots,
	}
}

type Project struct {
	files map[string]file
	roots []string
}

type file struct {
	ast      *ir.Root
	content  string
	modified time.Time
}

type Position struct {
	Row  uint
	Col  uint
	Path string
}

func (p *Project) Parse() error {
	start := time.Now()
	defer func() {
		log.Infof(
			"Parsed %d files of %d root folders in %s\n.",
			len(p.files),
			len(p.roots),
			time.Since(start),
		)
	}()

	config := conf.Config{
		ErrorHandlerFunc: func(e *perrors.Error) {
			panic(e)
		},
		// TODO: Get php version from 'php --version', or is there a builtin lsp way of getting language version?
		Version: &version.Version{Major: 8, Minor: 0},
	}

	for _, root := range p.roots {
		p.ParseRoot(root, config)
	}

	return nil
}

func (p *Project) ParseRoot(root string, config conf.Config) error {
	filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
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

		irNode := irconv.ConvertNode(rootNode)
		irRootNode, ok := irNode.(*ir.Root)
		if !ok {
			panic("Not ok")
		}

		p.files[path] = file{
			ast:      irRootNode,
			content:  string(content),
			modified: info.ModTime(),
		}
		return nil
	})

	return nil
}

func (p *Project) Definition(path string, pos *Position) (*Position, error) {
	start := time.Now()
	defer func() {
		log.Infof("Retrieving definition took %s\n", time.Since(start))
	}()

	file, ok := p.files[path]
	if !ok {
		// TODO: better
		panic("File not found " + path)
	}

	// goDumper := dumper.NewDumper(os.Stdout).WithTokens().WithPositions()
	// file.ast.Accept(goDumper)

	apos := position.FromLocation(file.content, pos.Row, pos.Col)
	nap := traversers.NewNodeAtPos(apos)
	file.ast.Walk(nap)

	// Root.
	var scope ir.Node
	for _, node := range nap.Nodes {
		// fmt.Printf("%T\n", node)
		switch typedNode := node.(type) {
		case *ir.Root:
			scope = typedNode
			break
		case *ir.FunctionStmt:
			scope = typedNode
			break
		case *ir.SimpleVar:
			assignment := p.assignment(scope, typedNode)
			// fmt.Printf("%#v\n", assignment)
			if assignment == nil {
				return nil, errors.New("No assignment found matching the variable at given position")
			}

			pos := ir.GetPosition(assignment)
			_, col := position.ToLocation(file.content, uint(pos.StartPos))

			return &Position{
				Row: uint(pos.StartLine),
				Col: col,
			}, nil
		case *ir.FunctionCallExpr:
			function, path := p.function(scope, typedNode)

			if function == nil {
				return nil, errors.New("No assignment found matching the variable at given position")
			}

			pos := ir.GetPosition(function)
			_, col := position.ToLocation(file.content, uint(pos.StartPos))

			return &Position{
				Row:  uint(pos.StartLine),
				Col:  col,
				Path: path,
			}, nil
		}
	}

	return nil, errors.New("No definition found for given arguments")
}

func (p *Project) assignment(scope ir.Node, variable *ir.SimpleVar) ir.Node {
	// OPTIM: will in the future need to span multiple files, but lets be basic about this.

	traverser := traversers.NewAssignment(variable)
	scope.Walk(traverser)

	return traverser.Assignment
}

func (p *Project) function(scope ir.Node, call *ir.FunctionCallExpr) (*ir.FunctionStmt, string) {
	traverser, err := traversers.NewFunction(call)
	if err != nil {
		return nil, ""
	}

	scope.Walk(traverser)
	if traverser.Function != nil {
		return traverser.Function, ""
	}

	for path, file := range p.files {
		file.ast.Walk(traverser)

		if traverser.Function != nil {
			return traverser.Function, path
		}
	}

	return nil, ""
}
