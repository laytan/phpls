package index

import (
	"fmt"
	"log"
	"path"
	"reflect"
	"strings"
	"sync"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/parsing"
	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/pkg/symboltrie"
	"github.com/laytan/elephp/pkg/traversers"
	"github.com/laytan/elephp/pkg/typer"
	"golang.org/x/exp/slices"
)

const (
	ErrParseFmt    = "Error indexing %s, unable to parse: %w"
	ErrNotFoundFmt = "Could not find %s of kinds(%v) in the index"
	ErrPoolFmt     = "Could not create symbol traverser: %w"
)

var stubsPath = path.Join(pathutils.Root(), "phpstorm-stubs")

type Index interface {
	Index(path string, content string) error

	Refresh(path string, content string) error

	Delete(path string) error

	// Finds a symbol with the given FQN matching the given node kinds.
	// The given namespace must be fully qualified.
	//
	// Giving this no kinds will return any kind.
	Find(FQN string, kind ...ir.NodeKind) (*traversers.TrieNode, error)

	// Finds a prefix/completes a string.
	// Do not call this with a namespaced symbol, only the class or function name.
	//
	// Giving this no kinds will return any kind.
	// A max of 0 will return everything.
	FindPrefix(prefix string, max uint, kind ...ir.NodeKind) []*traversers.TrieNode
}

type index struct {
	normalParser     parsing.Parser
	stubParser       parsing.Parser
	symbolTrie       *symboltrie.Trie[*traversers.TrieNode]
	symbolTraversers *sync.Pool
}

func New(phpv *phpversion.PHPVersion) Index {
	trie := symboltrie.New[*traversers.TrieNode]()

	p := &sync.Pool{
		New: func() any {
			return traversers.NewSymbol(trie)
		},
	}

	// TODO: 2 parsers are not ideal
	normalParser := parsing.New(phpv)
	stubsParser := parsing.New(phpversion.EightOne())

	return &index{
		normalParser:     normalParser,
		stubParser:       stubsParser,
		symbolTrie:       trie,
		symbolTraversers: p,
	}
}

func (i *index) Index(path string, content string) error {
	root, err := i.parser(path).Parse(content)
	if err != nil {
		return fmt.Errorf(ErrParseFmt, path, err)
	}

	traverser := i.symbolTraversers.Get()
	t := traverser.(*traversers.Symbol)

	t.SetPath(path)

	func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println(
					fmt.Errorf("[WARNING] Could not index %s, parse/syntax error: %v", path, err),
				)
			}
		}()

		root.Walk(t)
	}()

	i.symbolTraversers.Put(traverser)

	return nil
}

func (i *index) Find(FQN string, kind ...ir.NodeKind) (*traversers.TrieNode, error) {
	FQNObj := typer.NewFQN(FQN)

	results := i.symbolTrie.SearchExact(FQNObj.Name())
	for _, result := range results {
		if result.Namespace != FQNObj.Namespace() {
			continue
		}

		if len(kind) == 0 || slices.Contains(kind, result.Symbol.NodeKind()) {
			return result, nil
		}
	}

	return nil, fmt.Errorf(ErrNotFoundFmt, FQN, kind)
}

func (i *index) FindPrefix(prefix string, max uint, kind ...ir.NodeKind) []*traversers.TrieNode {
	results := i.symbolTrie.SearchPrefix(prefix, max)

	values := make([]*traversers.TrieNode, len(results))
	for i, result := range results {
		values[i] = result.Value
	}

	return values
}

func (i *index) Refresh(path string, content string) error {
	if err := i.Delete(path); err != nil {
		return err
	}

	return i.Index(path, content)
}

// PERF: this is bad.
func (i *index) Delete(path string) error {
	parser := i.parser(path)
	content, err := parser.Read(path)
	if err != nil {
		return fmt.Errorf(ErrParseFmt, path, err)
	}

	prevRoot, err := parser.Parse(content)
	if err != nil {
		return fmt.Errorf(ErrParseFmt, path, err)
	}

	removeTrie := symboltrie.New[*traversers.TrieNode]()
	removeTraverser := traversers.NewSymbol(removeTrie)
	removeTraverser.SetPath(path)

	prevRoot.Walk(removeTraverser)
	toRemove := removeTrie.SearchPrefix("", 0)

	for _, node := range toRemove {
		i.symbolTrie.Delete(node.Key, func(tn *traversers.TrieNode) bool {
			return reflect.DeepEqual(node.Value, tn)
		})
	}

	log.Printf("Removed %d symbols from %s out of the symboltrie", len(toRemove), path)

	return nil
}

func (i *index) parser(path string) parsing.Parser {
	if strings.HasPrefix(path, stubsPath) {
		return i.stubParser
	}

	return i.normalParser
}
