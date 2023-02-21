package index

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/config"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/parsing"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/pkg/symbol"
	"github.com/laytan/elephp/pkg/symboltrie"
	"github.com/samber/do"
)

const (
	errParseFmt    = "Error indexing %s, unable to parse: %w"
	errNotFoundFmt = "Could not find %s of kinds(%#v) in the index"
)

var (
	ErrNotFound = fmt.Errorf(errNotFoundFmt, "", "")
	ErrParse    = fmt.Errorf(errParseFmt, "", nil)
	Config      = func() config.Config { return do.MustInvoke[config.Config](nil) }
)

type Index interface {
	Index(path string, content string) error

	Refresh(path string, content string) error

	Delete(path string) error

	// Finds a symbol with the given FQN matching the given node kinds.
	// The given namespace must be fully qualified.
	//
	// Giving this no kinds or ir.KindRoot will return any kind.
	Find(key *fqn.FQN) (*INode, bool)

	// Finds a prefix/completes a string.
	// Do not call this with a namespaced symbol, only the class or function name.
	//
	// Giving this no kinds will return any kind.
	// A max of 0 or passing ir.KindRoot will return everything.
	FindPrefix(prefix string, max int, kind ...ir.NodeKind) []*INode
}

type INode struct {
	FQN    *fqn.FQN
	Path   string
	Symbol symbol.Symbol
}

func NewINode(fqns *fqn.FQN, path string, symbol symbol.Symbol) *INode {
	return &INode{
		FQN:    fqns,
		Path:   path,
		Symbol: symbol,
	}
}

type index struct {
	normalParser parsing.Parser
	stubParser   parsing.Parser

	symbolTrie       *symboltrie.Trie[*INode]
	symbolTraversers *sync.Pool
}

func New(phpv *phpversion.PHPVersion) Index {
	// TODO: 2 parsers are not ideal
	normalParser := parsing.New(phpv)
	stubsParser := parsing.New(phpversion.EightOne())

	ind := &index{
		normalParser: normalParser,
		stubParser:   stubsParser,

		symbolTrie: symboltrie.New[*INode](),
	}

	ind.symbolTraversers = &sync.Pool{
		New: func() any {
			return NewIndexTraverser()
		},
	}

	return ind
}

func FromContainer() Index {
	return do.MustInvoke[Index](nil)
}

func (i *index) Index(path string, content string) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Could not parse %s into an AST: %v", path, r)
		}
	}()

	root, err := i.parser(path).Parse(content)
	if err != nil {
		return fmt.Errorf(errParseFmt, path, err)
	}

	t := i.symbolTraversers.Get().(*INodeTraverser)
	nodes := make(chan *INode, 10)
	t.Reset(path, nodes)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Could not index %s: %v", path, r)
				close(nodes)
			}
		}()

		root.Walk(t)
	}()

	for node := range nodes {
		i.symbolTrie.Put(node.FQN, node)
	}

	i.symbolTraversers.Put(t)

	return nil
}

// Find returns the first result matching the given query.
func (i *index) Find(key *fqn.FQN) (*INode, bool) {
	return i.symbolTrie.SearchExact(key)
}

func (i *index) FindPrefix(prefix string, max int, kind ...ir.NodeKind) []*INode {
	results := i.symbolTrie.SearchNames(prefix, max)

	values := make([]*INode, 0, len(results))
	for _, result := range results {
		if result.Symbol.MatchesKind(kind) {
			values = append(values, result)
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
// TODO: this reads the file content and treats it as "last", but the LSP does
// not care about if something is saved or not.
// We have to see if the LSP comes with the previous content, or explicitly keep
// track of some map from the path to the nodes and delete those.
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

	nodes := make(chan *INode, 25)
	removeTraverser := i.symbolTraversers.Get().(*INodeTraverser)
	removeTraverser.Reset(path, nodes)
	prevRoot.Walk(removeTraverser)
	i.symbolTraversers.Put(removeTraverser)

	j := 0
	for node := range nodes {
		i.symbolTrie.Delete(node.FQN)
		j++
	}

	log.Printf("Removed %d symbols from %s out of the symboltrie", j, path)

	return nil
}

func (i *index) parser(path string) parsing.Parser {
	if strings.HasPrefix(path, Config().StubsDir()) {
		return i.stubParser
	}

	return i.normalParser
}
