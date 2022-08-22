package project

import (
	"path"
	"strings"
	"sync"
	"time"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/php-parser/pkg/conf"
	perrors "github.com/VKCOM/php-parser/pkg/errors"
	"github.com/VKCOM/php-parser/pkg/version"
	"github.com/laytan/elephp/pkg/datasize"
	"github.com/laytan/elephp/pkg/lfudacache"
	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/elephp/pkg/symboltrie"
	"github.com/laytan/elephp/pkg/traversers"
	log "github.com/sirupsen/logrus"
)

const cacheSize = datasize.MegaByte * 100

type Project struct {
	mu sync.Mutex

	// Path to file map.
	files map[string]*File

	roots []string

	version *phpversion.PHPVersion

	// Symbol trie for global variables, classes, interfaces etc.
	// End goal being: never needing to traverse the whole project to search
	// for something.
	symbolTrie *symboltrie.Trie[*traversers.TrieNode]

	cache *lfudacache.Cache[string, *ir.Root]
}

func NewProject(root string, phpv *phpversion.PHPVersion) *Project {
	// OPTIM: parse phpstorm stubs once and store somehow, they won't change.
	stubs := path.Join(pathutils.Root(), "phpstorm-stubs")

	roots := []string{root, stubs}
	return &Project{
		files:      make(map[string]*File),
		roots:      roots,
		version:    phpv,
		symbolTrie: symboltrie.New[*traversers.TrieNode](),
		cache:      lfudacache.New[string, *ir.Root](cacheSize),
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

func (p *Project) ParserConfig() conf.Config {
	return p.ParserConfigWith(func(e *perrors.Error) {
		// TODO: when we get a parse/syntax error, publish the error
		// TODO: via diagnostics (lsp).
		// OPTIM: when we get a parse error, maybe don't use the faulty ast but use the latest
		// OPTIM: valid version. Currently it tries to parse as much as it can but stops on an error.
		log.Info(e)
	})
}

func (p *Project) ParserConfigWith(errHandler func(*perrors.Error)) conf.Config {
	return conf.Config{
		ErrorHandlerFunc: errHandler,
		// TODO: this should use the php version of the system,
		// TODO: but for phpstorm-stubs we need to pass a later version as it uses attributes for example.
		// TODO: so need to make a distinction between parsing project and stubs, (or just always use the latest php?).
		// TODO: if we are going to serialize the ast for stubs anyway, we could just parse with latest version and store that.
		Version: &version.Version{
			Major: 8,
			Minor: 1,
		},
	}
}

func (p *Project) ParserConfigWrapWithPath(path string) conf.Config {
	return p.ParserConfigWith(func(err *perrors.Error) {
		log.Infof(`Parse error for path "%s": %+v`, path, err)
	})
}

// Returns the position for the namespace statement that matches the given position.
func (p *Project) Namespace(pos *position.Position) *position.Position {
	file := p.GetFile(pos.Path)
	root := p.ParseFileCached(file)

	traverser := traversers.NewNamespace(pos.Row)
	root.Walk(traverser)

	if traverser.Result == nil {
		log.Infoln("Did not find namespace")
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

	log.Infof("Class: %s, fqn: %s\n", className, fqn)

	// Get how it would be resolved in the current file state.
	actFqn := p.FQN(root, &ir.Name{Value: className})

	log.Infof("Reolved: %s\n", actFqn.String())

	// If the resolvement in current state equals the wanted fqn, no use stmt is needed.
	return actFqn.String() != fqn
}
