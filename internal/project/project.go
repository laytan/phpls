package project

import (
	"path"
	"sync"
	"time"

	"github.com/VKCOM/php-parser/pkg/conf"
	perrors "github.com/VKCOM/php-parser/pkg/errors"
	"github.com/VKCOM/php-parser/pkg/version"
	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/pkg/symboltrie"
	"github.com/laytan/elephp/pkg/traversers"
	log "github.com/sirupsen/logrus"
)

type Project struct {
	mu sync.Mutex

	// Path to file map.
	files map[string]*File

	roots        []string
	ParserConfig conf.Config

	// Symbol trie for global variables, classes, interfaces etc.
	// End goal being: never needing to traverse the whole project to search
	// for something.
	symbolTrie *symboltrie.Trie[*traversers.TrieNode]
}

func NewProject(root string, phpv *phpversion.PHPVersion) *Project {
	// OPTIM: parse phpstorm stubs once and store somehow, they won't change.
	stubs := path.Join(pathutils.Root(), "phpstorm-stubs")

	roots := []string{root, stubs}
	return &Project{
		files: make(map[string]*File),
		roots: roots,
		ParserConfig: conf.Config{
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

// OPTIM: make use of a LRU cache here, that would make multiple calls to the same more performant.
// OPTIM: in general you are editting a file for some time and GetFile would be called a lot with the same files.
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
