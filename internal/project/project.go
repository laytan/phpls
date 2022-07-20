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
	"github.com/VKCOM/php-parser/pkg/conf"
	perrors "github.com/VKCOM/php-parser/pkg/errors"
	"github.com/VKCOM/php-parser/pkg/parser"
	"github.com/VKCOM/php-parser/pkg/version"
	"github.com/laytan/elephp/internal/traversers"
	"github.com/laytan/elephp/pkg/position"
	log "github.com/sirupsen/logrus"
)

var ErrNoDefinitionFound = errors.New("No definition found for symbol at given position")

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
		parserConfig: conf.Config{
			ErrorHandlerFunc: func(e *perrors.Error) {
				// TODO: when we get a parse/syntax error, publish the error
				// TODO: via diagnostics (lsp).
				// OPTIM: when we get a parse error, maybe don't use the faulty ast but use the latest
				// OPTIM: valid version. Currently it tries to parse as much as it can but stops on an error.
				log.Info(e)
			},
			// TODO: Get php version from 'php --version', or is there a builtin lsp way of getting language version?
			Version: &version.Version{Major: 8, Minor: 0},
		},
	}
}

type Project struct {
	files        map[string]file
	roots        []string
	parserConfig conf.Config
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

	for _, root := range p.roots {
		if err := p.ParseRoot(root); err != nil {
			return err
		}
	}

	return nil
}

func (p *Project) ParseRoot(root string) error {
	// NOTE: This does not walk symbolic links, is that a problem?
	return filepath.WalkDir(root, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			log.Error(fmt.Errorf("Error parsing root %s: %w", root, err))
			return nil
		}

		// OPTIM: https://github.com/rjeczalik/notify to keep an eye on file changes and adds.

		// TODO: make configurable what a php file is.
		// NOTE: the TextDocumentItem struct has the languageid on it, maybe use that?
		if !strings.HasSuffix(path, ".php") || info.IsDir() {
			return nil
		}

		finfo, err := info.Info()
		if err != nil {
			log.Error(fmt.Errorf("Error reading file info of %s: %w", path, err))
		}

		// If we currently have this parsed and the file hasn't changed, don't parse it again.
		if existing, ok := p.files[path]; ok {
			if !existing.modified.Before(finfo.ModTime()) {
				return nil
			}
		}

		if err := p.ParseFile(path, finfo.ModTime()); err != nil {
			log.Error(fmt.Errorf("Error parsing file %s: %w", path, err))
		}

		return nil
	})
}

func (p *Project) ParseFile(path string, modTime time.Time) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Error reading file %s: %w", path, err)
	}

	return p.ParseFileContent(path, content, modTime)
}

func (p *Project) ParseFileContent(path string, content []byte, modTime time.Time) error {
	rootNode, err := parser.Parse(content, p.parserConfig)
	if err != nil {
		return fmt.Errorf("Error parsing file %s into AST: %w", path, err)
	}

	irNode := irconv.ConvertNode(rootNode)
	irRootNode, ok := irNode.(*ir.Root)
	if !ok {
		return errors.New("AST root node could not be converted to IR root node")
	}

	p.files[path] = file{
		ast:      irRootNode,
		content:  string(content),
		modified: modTime,
	}

	return nil
}

func (p *Project) Definition(path string, pos *Position) (*Position, error) {
	start := time.Now()
	defer func() {
		log.Infof("Retrieving definition took %s\n", time.Since(start))
	}()

	file, ok := p.files[path]
	if !ok {
		if err := p.ParseFile(path, time.Now()); err != nil {
			return nil, fmt.Errorf("Definition error: %w", err)
		}

		file = p.files[path]
	}

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
				return nil, ErrNoDefinitionFound
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
				return nil, ErrNoDefinitionFound
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

	return nil, ErrNoDefinitionFound
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
