package project

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irconv"
	"github.com/VKCOM/php-parser/pkg/conf"
	perrors "github.com/VKCOM/php-parser/pkg/errors"
	"github.com/VKCOM/php-parser/pkg/parser"
	"github.com/VKCOM/php-parser/pkg/version"
	"github.com/laytan/elephp/internal/traversers"
	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/elephp/pkg/symboltrie"
	"github.com/shivamMg/trie"
	log "github.com/sirupsen/logrus"
)

var ErrNoDefinitionFound = errors.New("No definition found for symbol at given position")

func NewProject(root string, phpv *phpversion.PHPVersion) *Project {
	// OPTIM: parse phpstorm stubs once and store on filesystem or even in the binary because they won't change.
	//
	// OPTIM: This is also parsing tests for the specific stubs,
	// OPTIM: and has specific files for different php versions that should be excluded/handled appropriately.
	stubs := path.Join(pathutils.Root(), "phpstorm-stubs")

	roots := []string{root, stubs}
	return &Project{
		files:      make(map[string]file),
		namespaces: make(map[string][]string),
		roots:      roots,
		parserConfig: conf.Config{
			ErrorHandlerFunc: func(e *perrors.Error) {
				// TODO: when we get a parse/syntax error, publish the error
				// TODO: via diagnostics (lsp).
				// OPTIM: when we get a parse error, maybe don't use the faulty ast but use the latest
				// OPTIM: valid version. Currently it tries to parse as much as it can but stops on an error.
				log.Info(e)
			},
			Version: &version.Version{
				Major: uint64(phpv.Major),
				Minor: uint64(phpv.Minor),
			},
		},
		symbolTrie: symboltrie.New[*traversers.TrieNode](),
	}
}

type Project struct {
	mu sync.Mutex

	// Path to file map.
	files map[string]file
	// Namespace to file map.
	namespaces map[string][]string

	roots        []string
	parserConfig conf.Config

	// Symbol trie for global variables, classes, interfaces etc.
	// End goal being: never needing to traverse the whole project to search
	// for something.
	symbolTrie *symboltrie.Trie[*traversers.TrieNode]
}

type file struct {
	symbolTrie *trie.Trie
	content    string
	namespaces []string
	uses       []*ir.UseStmt
	modified   time.Time
}

func (f *file) parse(config conf.Config) (*ir.Root, error) {
	rootNode, err := parser.Parse([]byte(f.content), config)
	if err != nil {
		return nil, fmt.Errorf("Error parsing file into AST: %w", err)
	}

	irNode := irconv.ConvertNode(rootNode)
	irRootNode, ok := irNode.(*ir.Root)
	if !ok {
		return nil, errors.New("AST root node could not be converted to IR root node")
	}

	return irRootNode, nil
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
			"Parsed %d files of %d root folders in %s\n",
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
	// Waitgroup so this funtion can wait for everything to be parsed
	// before returning.
	wg := sync.WaitGroup{}

	// Semaphore to limit the number of go routines working at the same time.
	sem := make(chan struct{}, runtime.NumCPU())

	// NOTE: This does not walk symbolic links, is that a problem?
	err := filepath.WalkDir(root, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			log.Error(fmt.Errorf("Error parsing %s: %w", path, err))
			return nil
		}

		// TODO: make configurable what a php file is.
		// NOTE: the TextDocumentItem struct has the languageid on it, maybe use that?
		if !strings.HasSuffix(path, ".php") || info.IsDir() {
			return nil
		}

		// Start a new go routine to do the actual parsing work.
		// Make sure we wait for it to finish by adding to the wait group.
		wg.Add(1)
		go func(path string, info fs.DirEntry) {
			// Takes from the semaphore, this blocks until there is an available spot.
			sem <- struct{}{}
			// Make sure we release the semaphore when we are done.
			defer func() { <-sem }()
			// Signal the wait group this is done too.
			defer wg.Done()

			finfo, err := info.Info()
			if err != nil {
				log.Error(fmt.Errorf("Error reading file info of %s: %w", path, err))
			}

			// If we currently have this parsed and the file hasn't changed, don't parse it again.
			// if existing, ok := p.files[path]; ok {
			// 	if !existing.modified.Before(finfo.ModTime()) {
			// 		return
			// 	}
			// }

			if err := p.ParseFile(path, finfo.ModTime()); err != nil {
				log.Error(fmt.Errorf("Error parsing file %s: %w", path, err))
			}
		}(path, info)

		return nil
	})
	if err != nil {
		return fmt.Errorf("Error parsing project: %w", err)
	}

	wg.Wait()
	close(sem)

	return nil
}

func (p *Project) ParseFile(path string, modTime time.Time) error {
	content, err := os.ReadFile(path)
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

	// if strings.HasSuffix(path, "global-var.php") {
	// 	// TODO: comment out.
	// 	goDumper := dumper.NewDumper(os.Stdout).WithPositions()
	// 	rootNode.Accept(goDumper)
	// }

	irNode := irconv.ConvertNode(rootNode)
	irRootNode, ok := irNode.(*ir.Root)
	if !ok {
		return errors.New("AST root node could not be converted to IR root node")
	}

	symbolTraverser := traversers.NewSymbol(p.symbolTrie)
	symbolTraverser.SetPath(path)
	irRootNode.Walk(symbolTraverser)

	traverser := traversers.NewNamespaces()
	irRootNode.Walk(traverser)

	// Writing/Reading from a map needs to be done by one go routine at a time.
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, namespace := range traverser.Namespaces {
		namespaces, ok := p.namespaces[namespace]
		if ok {
			p.namespaces[namespace] = append(namespaces, path)
			continue
		}

		p.namespaces[namespace] = []string{path}
	}

	p.files[path] = file{
		content:    string(content),
		namespaces: traverser.Namespaces,
		uses:       traverser.Uses,
		modified:   modTime,
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

	ast, err := file.parse(p.parserConfig)
	if err != nil {
		return nil, fmt.Errorf("Error parsing %s: %w", path, err)
	}

	apos := position.FromLocation(file.content, pos.Row, pos.Col)
	nap := traversers.NewNodeAtPos(apos)
	ast.Walk(nap)

	// Root.
	var scope ir.Node
	for i, node := range nap.Nodes {
		// fmt.Printf("%T\n", node)
		switch typedNode := node.(type) {
		case *ir.Root:
			scope = typedNode

		case *ir.FunctionStmt:
			scope = typedNode

		case *ir.GlobalStmt:
			rootNode, ok := nap.Nodes[0].(*ir.Root)
			if !ok {
				log.Errorln("First node was not the root node, this should not happen")
				return nil, ErrNoDefinitionFound
			}

			globalVar, ok := nap.Nodes[i+1].(*ir.SimpleVar)
			if !ok {
				log.Errorln("Node after the global stmt was not a variable, which we expected")
				return nil, ErrNoDefinitionFound
			}

			assignment := p.globalAssignment(rootNode, globalVar)
			if assignment == nil {
				return nil, ErrNoDefinitionFound
			}

			pos := ir.GetPosition(assignment)
			_, col := position.ToLocation(file.content, uint(pos.StartPos))

			return &Position{
				Row: uint(pos.StartLine),
				Col: col,
			}, nil

		case *ir.SimpleVar:
			assignment := p.assignment(scope, typedNode)
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
			function, destPath := p.function(scope, typedNode)

			if function == nil {
				return nil, ErrNoDefinitionFound
			}

			if destPath == "" {
				destPath = path
			}

			destFile, ok := p.files[destPath]
			if !ok {
				log.Errorf("Destination at %s is not in parsed files cache\n", path)
				return nil, ErrNoDefinitionFound
			}

			pos := function.Position
			_, col := position.ToLocation(destFile.content, uint(pos.StartPos))

			return &Position{
				Row:  uint(pos.StartLine),
				Col:  col,
				Path: destPath,
			}, nil

		case *ir.Name:
			rootNode, ok := nap.Nodes[0].(*ir.Root)
			if !ok {
				log.Errorln("First node was not the root node, this should not happen")
				return nil, ErrNoDefinitionFound
			}

			// TODO: this should already be a pointer (file).
			classLike, destPath := p.classLike(&file, rootNode, typedNode)
			if classLike == nil {
				return nil, ErrNoDefinitionFound
			}

			if destPath == "" {
				destPath = path
			}

			destFile, ok := p.files[destPath]
			if !ok {
				log.Errorf("Destination at %s is not in parsed files cache\n", destPath)
				return nil, ErrNoDefinitionFound
			}

			pos := ir.GetPosition(classLike)
			_, col := position.ToLocation(destFile.content, uint(pos.StartPos))

			return &Position{
				Row:  uint(pos.StartLine),
				Col:  col,
				Path: destPath,
			}, nil
		}
	}

	return nil, ErrNoDefinitionFound
}

func (p *Project) assignment(scope ir.Node, variable *ir.SimpleVar) *ir.SimpleVar {
	// TODO: will in the future need to span multiple files, but lets be basic about this.

	traverser := traversers.NewAssignment(variable)
	scope.Walk(traverser)

	return traverser.Assignment
}

func (p *Project) globalAssignment(root *ir.Root, globalVar *ir.SimpleVar) *ir.SimpleVar {
	// First search the current file for the assignment.
	traverser := traversers.NewGlobalAssignment(globalVar)
	root.Walk(traverser)

	// TODO: search the whole project if the global is not assigned here.

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

	// No definition found locally, searching globally.
	name, ok := call.Function.(*ir.Name)
	if !ok {
		return nil, ""
	}

	results := p.symbolTrie.SearchExact(name.Value)

	if len(results) == 0 {
		return nil, ""
	}

	for _, res := range results {
		if function, ok := res.Node.(*ir.FunctionStmt); ok {
			return function, res.Path
		}
	}

	return nil, ""
}

// NOTE: the NameTkn stays the same, only the Value is changed to the FQN.
func (p *Project) nameToFQN(root *ir.Root, name *ir.Name) *ir.Name {
	if name.IsFullyQualified() {
		return name
	}

	traverser, err := traversers.NewFQN(name)
	if err != nil {
		return nil
	}

	root.Walk(traverser)

	return &ir.Name{
		Position: name.Position,
		NameTkn:  name.NameTkn,
		Value:    traverser.Result(),
	}
}

// Returns either *ir.ClassStmt, *ir.InterfaceStmt or *ir.TraitStmt.
func (p *Project) classLike(sourceFile *file, root *ir.Root, name *ir.Name) (ir.Node, string) {
	// OPTIM: first check the current file and included files (use statements), in most cases, that will match.
	// OPTIM: and only if there is no match, search globally.
	// OPTIM: there does need to be an index of path, to namespaces for this to work.
	// OPTIM: should we store the namespaces of a file in the file struct, would that be to much initial parsing?
	//
	// OPTIM: could even have an index of global (stdlib) classlikes for easy access.

	// TODO: abstract this common functionality.
	fqn := p.nameToFQN(root, name)
	namespace := strings.Trim(strings.TrimSuffix(fqn.Value, fqn.LastPart()), "\\")
	classLikeName := fqn.LastPart()

	traverser, err := traversers.NewClassLike(fqn)
	if err != nil {
		log.Error(err)
		return nil, ""
	}

	root.Walk(traverser)
	if traverser.Result != nil {
		return traverser.Result, ""
	}

	// TODO: aliasses?
	for _, usage := range sourceFile.uses {
		usageNamespace := strings.Trim(
			strings.TrimSuffix(usage.Use.Value, usage.Use.LastPart()),
			"\\",
		)

		paths, ok := p.namespaces[usageNamespace]
		if !ok {
			log.Error(fmt.Errorf("Namespace %s is not indexed", usageNamespace))
			continue
		}

		for _, path := range paths {
			file, ok := p.files[path]
			if !ok {
				log.Error(fmt.Errorf("No file struct for path %s", path))
				continue
			}

			ast, err := file.parse(p.parserConfig)
			if err != nil {
				log.Error(err)
				continue
			}

			ast.Walk(traverser)
			if traverser.Result != nil {
				return traverser.Result, path
			}
		}
	}

	results := p.symbolTrie.SearchExact(classLikeName)

	if len(results) == 0 {
		return nil, ""
	}

	for _, res := range results {
		if res.Namespace != namespace {
			continue
		}

		// Check is class like.
		nodeKind := ir.GetNodeKind(res.Node)
		if nodeKind != ir.KindClassStmt &&
			nodeKind != ir.KindTraitStmt &&
			nodeKind != ir.KindInterfaceStmt {
			continue
		}

		return res.Node, res.Path
	}

	return nil, ""
}
