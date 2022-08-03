package project

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irconv"
	"github.com/VKCOM/php-parser/pkg/parser"
	"github.com/VKCOM/php-parser/pkg/visitor/dumper"
	"github.com/laytan/elephp/pkg/traversers"
	log "github.com/sirupsen/logrus"
)

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
	rootNode, err := parser.Parse(content, p.ParserConfig)
	if err != nil {
		return fmt.Errorf("Error parsing file %s into AST: %w", path, err)
	}

	if strings.HasSuffix(path, "methods.php") {
		goDumper := dumper.NewDumper(os.Stdout).WithPositions()
		rootNode.Accept(goDumper)
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
	}

	return nil
}
