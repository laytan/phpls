package project

import (
	"fmt"
	"log"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/php-parser/pkg/conf"
	perrors "github.com/VKCOM/php-parser/pkg/errors"
	"github.com/VKCOM/php-parser/pkg/version"
	"github.com/laytan/elephp/pkg/arccache"
	"github.com/laytan/elephp/pkg/datasize"
	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/elephp/pkg/symboltrie"
	"github.com/laytan/elephp/pkg/traversers"
	"github.com/laytan/elephp/pkg/typer"
)

const cacheSize = datasize.MegaByte * 100

type root struct {
	Path    string
	Version *phpversion.PHPVersion
}

type Project struct {
	mu sync.Mutex

	// Path to file map.
	files map[string]*File

	roots []*root

	// Symbol trie for global variables, classes, interfaces etc.
	// End goal being: never needing to traverse the whole project to search
	// for something.
	symbolTrie *symboltrie.Trie[*traversers.TrieNode]

	cache *arccache.Cache[string, *ir.Root]

	typer typer.Typer

	// Extensions to treat as PHP source files.
	fileExtensions []string
}

func NewProject(r string, phpv *phpversion.PHPVersion, fileExtensions []string) *Project {
	roots := []*root{
		{Path: r, Version: phpv},
		// Parse stubs using the latest supported PHP version, bcs they use the latest features.
		{Path: path.Join(pathutils.Root(), "phpstorm-stubs"), Version: phpversion.EightOne()},
	}

	return &Project{
		files:          make(map[string]*File),
		roots:          roots,
		symbolTrie:     symboltrie.New[*traversers.TrieNode](),
		cache:          arccache.New[string, *ir.Root](cacheSize),
		typer:          typer.New(),
		fileExtensions: fileExtensions,
	}
}

func (p *Project) GetFile(path string) *File {
	file, ok := p.files[path]
	if !ok {
		if err := p.ParseFile(path, time.Now()); err != nil {
			return nil
		}

		file = p.files[path]
	}

	return file
}

func (p *Project) ParserConfig(phpv *phpversion.PHPVersion) conf.Config {
	return p.ParserConfigWith(phpv, func(e *perrors.Error) {
		// TODO: when we get a parse/syntax error, publish the error
		// TODO: via diagnostics (lsp).
		// OPTIM: when we get a parse error, maybe don't use the faulty ast but use the latest
		// OPTIM: valid version. Currently it tries to parse as much as it can but stops on an error.
		log.Println(e)
	})
}

func (p *Project) ParserConfigWith(
	phpv *phpversion.PHPVersion,
	errHandler func(*perrors.Error),
) conf.Config {
	return conf.Config{
		ErrorHandlerFunc: errHandler,
		Version: &version.Version{
			Major: uint64(phpv.Major),
			Minor: uint64(phpv.Minor),
		},
	}
}

func (p *Project) ParserConfigWrapWithPath(path string) conf.Config {
	phpv, err := p.phpversionForPath(path)
	if err != nil {
		panic(err)
	}

	return p.ParserConfigWith(phpv, func(err *perrors.Error) {
		log.Printf(`Parse error for path "%s": %+v`, path, err)
	})
}

// Checks in what root the path belongs and returns its version.
func (p *Project) phpversionForPath(path string) (*phpversion.PHPVersion, error) {
	var phpv *phpversion.PHPVersion
	for _, root := range p.roots {
		if strings.HasPrefix(path, root.Path) {
			phpv = root.Version
		}
	}

	if phpv == nil {
		return nil, fmt.Errorf("Path %s is not in any of the tracked root/workspaces", path)
	}

	return phpv, nil
}

// Returns the position for the namespace statement that matches the given position.
func (p *Project) Namespace(pos *position.Position) *position.Position {
	file := p.GetFile(pos.Path)
	root := p.ParseFileCached(file)

	traverser := traversers.NewNamespace(pos.Row)
	root.Walk(traverser)

	if traverser.Result == nil {
		log.Println("Did not find namespace")
		return nil
	}

	row, col := position.PosToLoc(file.content, uint(traverser.Result.Position.StartPos))

	return &position.Position{
		Row:  row,
		Col:  col,
		Path: pos.Path,
	}
}

// Returns whether the file at given pos needs a use statement for the given fqn.
func (p *Project) NeedsUseStmtFor(pos *position.Position, fqn string) bool {
	file := p.GetFile(pos.Path)
	root := p.ParseFileCached(file)

	parts := strings.Split(fqn, `\`)
	className := parts[len(parts)-1]

	// Get how it would be resolved in the current file state.
	actFqn := p.FQN(root, &ir.Name{Value: className})

	// If the resolvement in current state equals the wanted fqn, no use stmt is needed.
	return actFqn.String() != fqn
}
