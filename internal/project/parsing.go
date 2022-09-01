package project

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irconv"
	"github.com/VKCOM/php-parser/pkg/parser"
	"github.com/laytan/elephp/pkg/symboltrie"
	"github.com/laytan/elephp/pkg/traversers"
	"golang.org/x/sync/errgroup"
)

func (p *Project) Parse(numFiles *atomic.Uint32) error {
	// Parsing creates alot of garbage, after parsing, run a gc cycle manually
	// because we know there is a lot to clean up.
	defer func() {
		go runtime.GC()
	}()

	// TODO: do this concurrently.
	for _, root := range p.roots {
		if err := p.ParseRoot(root, numFiles); err != nil {
			return err
		}
	}

	return nil
}

func (p *Project) ParseRoot(root *root, numFiles *atomic.Uint32) error {
	log.Printf("Parsing root %s with PHP version %s\n", root.Path, root.Version.String())

	g := new(errgroup.Group)
	g.SetLimit(runtime.NumCPU())

	// NOTE: This does not walk symbolic links, is that a problem?
	err := filepath.WalkDir(root.Path, func(path string, info fs.DirEntry, err error) error {
		if !p.ShouldParseFile(info) {
			return nil
		}

		if err != nil {
			log.Println(fmt.Errorf("Error parsing %s: %w", path, err))
			return nil
		}

		g.Go(func() error {
			defer func() { numFiles.Add(1) }()

			finfo, err := info.Info()
			if err != nil {
				log.Println(fmt.Errorf("Error reading file info of %s: %w", path, err))
			}

			if err := p.ParseFile(path, finfo.ModTime()); err != nil {
				log.Println(fmt.Errorf("Error parsing file %s: %w", path, err))
			}

			return nil
		})

		return nil
	})
	if err != nil {
		return fmt.Errorf("Error parsing project: %w", err)
	}
	if err := g.Wait(); err != nil {
		return fmt.Errorf("Error parsing project: %w", err)
	}

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
	rootNode, err := parser.Parse(content, p.ParserConfigWrapWithPath(path))
	if err != nil {
		return fmt.Errorf("Error parsing file %s into AST: %w", path, err)
	}

	irNode := irconv.ConvertNode(rootNode)
	irRootNode, ok := irNode.(*ir.Root)
	if !ok {
		return errors.New("AST root node could not be converted to IR root node")
	}

	symbolTraverser := traversers.NewSymbol(p.symbolTrie)
	symbolTraverser.SetPath(path)
	irRootNode.Walk(symbolTraverser)

	// Writing/Reading from a map needs to be done by one go routine at a time.
	p.mu.Lock()
	defer p.mu.Unlock()

	p.files[path] = &File{
		content:  string(content),
		modified: modTime,
		path:     path,
	}

	return nil
}

func (p *Project) ParseFileUpdate(path string, content string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// If we already have it, and the content hasn't changed, skip.
	if f, ok := p.files[path]; ok {
		if removeWhitespace(content) == removeWhitespace(f.content) {
			log.Println("File parsing with only whitespace changes, skipping this change")
			return nil
		}
	}

	rootNode, err := parser.Parse([]byte(content), p.ParserConfigWrapWithPath(path))
	if err != nil {
		return fmt.Errorf("Error parsing file %s into AST: %w", path, err)
	}

	irNode := irconv.ConvertNode(rootNode)
	irRootNode, ok := irNode.(*ir.Root)
	if !ok {
		return errors.New("AST root node could not be converted to IR root node")
	}

	// Remove all symbols that are in this file before adding the updated ones.
	p.removeSymbolsOf(path)

	symbolTraverser := traversers.NewSymbol(p.symbolTrie)
	symbolTraverser.SetPath(path)
	irRootNode.Walk(symbolTraverser)

	// Content changed, so delete from cache, this does reset its score if
	// it is entered again, might not be ideal.
	p.cache.Delete(path)

	p.files[path] = &File{
		content:  string(content),
		modified: time.Now(),
		path:     path,
	}

	return nil
}

func (p *Project) ParseFileCached(file *File) *ir.Root {
	if file == nil {
		return nil
	}

	return p.cache.Cached(file.path, func() *ir.Root {
		root, err := file.parse(p.ParserConfigWrapWithPath(file.path))
		if err != nil {
			log.Println(err)
		}

		return root
	})
}

func (p *Project) ShouldParseFile(info fs.DirEntry) bool {
	if info.IsDir() {
		return false
	}

	name := info.Name()
	for _, extension := range p.fileExtensions {
		if strings.HasSuffix(name, extension) {
			return true
		}
	}

	return false
}

func (p *Project) removeSymbolsOf(path string) {
	prevFile := p.GetFile(path)
	prevRootNode := p.ParseFileCached(prevFile)
	removeTrie := symboltrie.New[*traversers.TrieNode]()
	removeTraverser := traversers.NewSymbol(removeTrie)
	removeTraverser.SetPath(path)
	prevRootNode.Walk(removeTraverser)
	toRemove := removeTrie.SearchPrefix("", 0)

	for _, node := range toRemove {
		p.symbolTrie.Delete(node.Key, func(tn *traversers.TrieNode) bool {
			return reflect.DeepEqual(node.Value, tn)
		})
	}

	log.Printf("Removed %d symbols from %s out of the symboltrie", len(toRemove), path)
}

// Fast way of removing all whitespace from a string, credit: https://stackoverflow.com/a/32081891.
func removeWhitespace(text string) string {
	var b strings.Builder
	b.Grow(len(text))
	for _, ch := range text {
		if !unicode.IsSpace(ch) {
			b.WriteRune(ch)
		}
	}

	return b.String()
}
