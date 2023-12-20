package wrkspc

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/lexer"
	"github.com/laytan/phpls/internal/config"
	"github.com/laytan/phpls/pkg/cache"
	"github.com/laytan/phpls/pkg/parsing"
	"github.com/laytan/phpls/pkg/pathutils"
	"github.com/laytan/phpls/pkg/phpversion"
	"golang.org/x/sync/errgroup"
)

const astCacheCapacity = 25

var (
	ErrFileNotIndexed    = errors.New("file is not indexed in the workspace")
	ErrRead              = errors.New("unable to read file")
	indexGoRoutinesLimit = 64
	Current              Wrkspc
)

type Wrkspc interface {
	Root() string

	PutOverlay(path, content string)
	DeleteOverlay(path string)

	Walk(files chan<- *ParsedFile, total *atomic.Uint64, opts WalkOptions) error

	Content(path string) (string, error)
	ContentF(path string) string

	Ast(path string) (*ast.Root, error)
	AstF(path string) *ast.Root

	Parse(path string, content []byte) (*ast.Root, error)

	All(path string) (content string, root *ast.Root, err error)
	AllF(path string) (content string, root *ast.Root)

	Lexer(path string) (lexer.Lexer, error)

	IsPhp(path string) bool
}

type ParsedFile struct {
	Path    string
	Content string
}

// fileExtensions should all start with a period.
func New(phpv *phpversion.PHPVersion, root string, stubs string) Wrkspc {
	return &wrkspc{
		normalParser:    parsing.New(phpv),
		stubParser:      parsing.New(phpversion.Latest()),
		roots:           []string{root, stubs},
		fileExtensions:  config.Current.Extensions,
		ignoredDirNames: config.Current.IgnoredDirectories,
		overlays:        map[string]string{},
		asts:            cache.New[string, *ast.Root](astCacheCapacity),
		totals:          map[WalkOptions]uint64{},
	}
}

type wrkspc struct {
	normalParser    *parsing.Parser
	stubParser      *parsing.Parser
	roots           []string
	fileExtensions  []string
	ignoredDirNames []string
	asts            *cache.Cache[string, *ast.Root]

	// Overlays are open files/buffers, the contents might not be saved yet,
	// So we can't rely on the file contents.
	overlays   map[string]string
	overlaysMu sync.RWMutex

	// Cache the totals, so that after the first walk, the total will be pretty accurate from the start.
	totals map[WalkOptions]uint64
}

func (w *wrkspc) Root() string {
	return w.roots[0]
}

func (w *wrkspc) PutOverlay(path string, content string) {
	w.overlaysMu.Lock()
	w.overlays[path] = content
	w.overlaysMu.Unlock()

	// Refresh/put updates in the cache.
	root, err := w.Parse(path, []byte(content))
	if err != nil {
		log.Printf("[WARN]: could not parse new overlay AST")
		w.asts.Delete(path)
		return
	}

	w.asts.Put(path, root)
}

func (w *wrkspc) DeleteOverlay(path string) {
	w.overlaysMu.Lock()
	delete(w.overlays, path)
	w.overlaysMu.Unlock()
}

type WalkOptions struct {
	DoStubs  bool
	DoVendor bool
}

var WalkAll = WalkOptions{
	DoStubs:  true,
	DoVendor: true,
}

// TODO: error handling.
func (w *wrkspc) Walk(files chan<- *ParsedFile, total *atomic.Uint64, opts WalkOptions) error {
	w.overlaysMu.RLock()
	defer w.overlaysMu.RUnlock()

	defer close(files)

	// Set total to the cached total so that it is "accurate" from the start.
	var actTotal *atomic.Uint64
	if cachedTotal, ok := w.totals[opts]; ok {
		total.Store(cachedTotal)
		actTotal = &atomic.Uint64{}
	} else {
		actTotal = total
	}

	for path, content := range w.overlays {
		actTotal.Add(1)
		files <- &ParsedFile{Path: path, Content: content}
	}

	g := errgroup.Group{}
	g.SetLimit(indexGoRoutinesLimit)

	if err := w.walk(opts, func(path string, d fs.DirEntry) error {
		if _, ok := w.overlays[path]; ok {
			return nil
		}

		actTotal.Add(1)

		g.Go(func() error {
			content, err := w.parser(path).Read(path)
			if err != nil {
				return fmt.Errorf("%q: %w: %w", path, ErrRead, err)
			}

			files <- &ParsedFile{Path: path, Content: content}
			return nil
		})

		return nil
	}); err != nil {
		panic(err)
	}

	// Update the cache and set total to the actual total.
	act := actTotal.Load()
	total.Store(act)
	w.totals[opts] = act

	if err := g.Wait(); err != nil {
		panic(err)
	}

	return nil
}

// Content returns the content of the file at the given path.
func (w *wrkspc) Content(path string) (string, error) {
	w.overlaysMu.RLock()
	defer w.overlaysMu.RUnlock()
	if overlay, ok := w.overlays[path]; ok {
		return overlay, nil
	}

	content, err := w.parser(path).Read(path)
	if err != nil {
		return "", fmt.Errorf("%q: %w: %w", path, ErrRead, err)
	}

	return content, nil
}

// ContentF returns the content of the file at the given path.
// If an error occurs, it logs it and returns an empty string.
// Use Content for access to the error.
func (w *wrkspc) ContentF(path string) string {
	content, err := w.Content(path)
	if err != nil {
		log.Println(err)
	}

	return content
}

// Ast returns the parsed root node of the given path.
// If an error occurs, it returns an empty root node, this NEVER returns nil.
func (w *wrkspc) Ast(path string) (*ast.Root, error) {
	root := w.asts.Cached(path, func() *ast.Root {
		content, err := w.Content(path)
		if err != nil {
			log.Println(err)
			return nil
		}

		root, err := w.Parse(path, []byte(content))
		if err != nil {
			log.Println(err)
			return nil
		}

		return root
	})

	if root == nil {
		return &ast.Root{}, ErrFileNotIndexed
	}

	return root, nil
}

// AstF returns the root of the given path, if an error occurs, it logs it
// and returns an empty root node. This never returns nil.
// Use AstF for access to the error.
func (w *wrkspc) AstF(path string) *ast.Root {
	root, err := w.Ast(path)
	if err != nil {
		log.Println(err)
	}

	return root
}

func (w *wrkspc) Parse(path string, content []byte) (*ast.Root, error) {
	root, err := w.parser(path).Parse(content)
	if err != nil {
		return nil, fmt.Errorf("parsing path %q with content %q: %w", path, string(content), err)
	}

	return root, nil
}

// All returns the content & parsed root node of the given path.
// If an error occurs, it returns an empty root node, this NEVER returns nil.
func (w *wrkspc) All(path string) (string, *ast.Root, error) {
	content, cErr := w.Content(path)
	root, rErr := w.Ast(path)

	if cErr != nil {
		return content, root, cErr
	}

	if rErr != nil {
		return content, root, rErr
	}

	return content, root, nil
}

// AllF returns both the content and root node of the given path.
// If an error occurs, it returns an empty root node and string after logging the error.
func (w *wrkspc) AllF(path string) (string, *ast.Root) {
	content, root, err := w.All(path)
	if err != nil {
		log.Println(err)
	}

	return content, root
}

func (w *wrkspc) Lexer(path string) (lexer.Lexer, error) {
	lexer, err := w.parser(path).Lexer([]byte(w.ContentF(path)))
	if err != nil {
		return nil, fmt.Errorf("creating lexer: %w", err)
	}

	return lexer, nil
}

func (w *wrkspc) IsPhp(path string) bool {
	for _, extension := range w.fileExtensions {
		if strings.HasSuffix(path, extension) {
			return true
		}
	}
	return false
}

func (w *wrkspc) walk(opts WalkOptions, walker func(path string, d fs.DirEntry) error) error {
	wg := sync.WaitGroup{}

	var finalErr error

	for _, root := range w.roots {
		if !opts.DoStubs && isStubs(root) {
			continue
		}

		wg.Add(1)
		go func(root string) {
			defer wg.Done()

			if err := w.parser(root).Walk(root, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				doParse, err := w.shouldParse(opts.DoVendor, d)
				if err != nil {
					return err
				}

				if doParse {
					return walker(path, d)
				}

				return nil
			}); err != nil {
				finalErr = err
			}
		}(root)
	}

	wg.Wait()
	return finalErr
}

func (w *wrkspc) shouldParse(doVendor bool, d fs.DirEntry) (bool, error) {
	if d.IsDir() {
		// Skip ignored directories.
		n := d.Name()
		if !doVendor && n == "vendor" {
			return false, filepath.SkipDir
		}

		for _, ignored := range w.ignoredDirNames {
			if n == ignored {
				return false, filepath.SkipDir
			}
		}

		return false, nil
	}

	if w.IsPhp(d.Name()) {
		return true, nil
	}

	return false, nil
}

func (w *wrkspc) parser(path string) *parsing.Parser {
	if isStubs(path) {
		return w.stubParser
	}

	return w.normalParser
}

var stubsDir = filepath.Join(pathutils.Root(), "third_party", "phpstorm-stubs")

func isStubs(path string) bool {
	return strings.HasPrefix(path, config.Current.StubsPath) || strings.HasPrefix(path, stubsDir)
}
