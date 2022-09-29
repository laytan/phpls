package index

import (
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/parsing"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/pkg/symboltrie"
	"github.com/laytan/elephp/pkg/traversers"
	"github.com/samber/do"
	"golang.org/x/exp/slices"
)

const (
	errParseFmt    = "Error indexing %s, unable to parse: %w"
	errNotFoundFmt = "Could not find %s of kinds(%#v) in the index"
)

var (
	ErrNotFound = fmt.Errorf(errNotFoundFmt, "", "")
	ErrParse    = fmt.Errorf(errParseFmt, "", nil)
)

var stubsPath = filepath.Join(pathutils.Root(), "phpstorm-stubs")

type Index interface {
	Index(path string, content string) error

	Refresh(path string, content string) error

	Delete(path string) error

	// Finds a symbol with the given FQN matching the given node kinds.
	// The given namespace must be fully qualified.
	//
	// Giving this no kinds or ir.KindRoot will return any kind.
	Find(fqn string, kind ...ir.NodeKind) (*traversers.TrieNode, error)

	// Finds a prefix/completes a string.
	// Do not call this with a namespaced symbol, only the class or function name.
	//
	// Giving this no kinds will return any kind.
	// A max of 0 or passing ir.KindRoot will return everything.
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

func FromContainer() Index {
	return do.MustInvoke[Index](nil)
}

func (i *index) Index(path string, content string) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Could not index %s, parse/syntax error: %v", path, r)
		}
	}()

	root, err := i.parser(path).Parse(content)
	if err != nil {
		return fmt.Errorf(errParseFmt, path, err)
	}

	traverser := i.symbolTraversers.Get()
	t := traverser.(*traversers.Symbol)
	t.SetPath(path)
	root.Walk(t)

	i.symbolTraversers.Put(traverser)

	return nil
}

func (i *index) Find(fqnStr string, kind ...ir.NodeKind) (*traversers.TrieNode, error) {
	FQNObj := fqn.NewFQN(fqnStr)

	retAll := len(kind) == 0 || slices.Contains(kind, ir.KindRoot)

	results := i.symbolTrie.SearchExact(FQNObj.Name())
	for _, result := range results {
		if result.Namespace != FQNObj.Namespace() {
			continue
		}

		if retAll || slices.Contains(kind, result.Symbol.NodeKind()) {
			return result, nil
		}
	}

	return nil, fmt.Errorf(errNotFoundFmt, fqnStr, kind)
}

func (i *index) FindPrefix(prefix string, max uint, kind ...ir.NodeKind) []*traversers.TrieNode {
	results := i.symbolTrie.SearchPrefix(prefix, max)

	retAll := len(kind) == 0 || slices.Contains(kind, ir.KindRoot)

	values := make([]*traversers.TrieNode, 0, len(results))
	for _, result := range results {
		if retAll || slices.Contains(kind, result.Value.Symbol.NodeKind()) {
			values = append(values, result.Value)
		}
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
		return fmt.Errorf(errParseFmt, path, err)
	}

	prevRoot, err := parser.Parse(content)
	if err != nil {
		return fmt.Errorf(errParseFmt, path, err)
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
