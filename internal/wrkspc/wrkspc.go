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
	"time"

	"appliedgo.net/what"
	"github.com/laytan/elephp/internal/config"
	"github.com/laytan/elephp/pkg/cache"
	"github.com/laytan/elephp/pkg/parsing"
	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/php-parser/pkg/ast"
	"golang.org/x/sync/errgroup"
)

const (
	ErrParseFmt     = "Error parsing file %s syntax nodes: %w"
	ErrReadFmt      = "Error reading file %s: %w"
	ErrParseWalkFmt = "Error walking the workspace files: %w"

	irCacheCapacity      = 25
	contentCacheCapacity = 50
)

var (
	ErrFileNotIndexed    = errors.New("File is not indexed in the workspace")
	indexGoRoutinesLimit = 64
	Current              Wrkspc
)

type file struct {
	content string
}

func newFile(content string) *file {
	return &file{
		content: content,
	}
}

type ParsedFile struct {
	Path    string
	Content string
}

type Wrkspc interface {
	Root() string

	Index(files chan<- *ParsedFile, total *atomic.Uint64, totalDone chan<- bool) error

	// ContentOf returns the content of the file at the given path.
	ContentOf(path string) (string, error)
	// FContentOf returns the content of the file at the given path.
	// If an error occurs, it logs it and returns an empty string.
	// Use ContentOf for access to the error.
	FContentOf(path string) string

	// IROf returns the parsed root node of the given path.
	// If an error occurs, it returns an empty root node, this NEVER returns nil.
	IROf(path string) (*ast.Root, error)
	// FIROf returns the root of the given path, if an error occurs, it logs it
	// and returns an empty root node. This never returns nil.
	// Use IROf for access to the error.
	FIROf(path string) *ast.Root

	// AllOf returns the content & parsed root node of the given path.
	// If an error occurs, it returns an empty root node, this NEVER returns nil.
	AllOf(path string) (string, *ast.Root, error)
	// FAllOf returns both the content and root node of the given path.
	// If an error occurs, it returns an empty root node and string after logging the error.
	FAllOf(path string) (string, *ast.Root)

	Refresh(path string) error

	RefreshFrom(path string, content string) error
}

// fileExtensions should all start with a period.
func New(phpv *phpversion.PHPVersion, root string, stubs string) Wrkspc {
	normalParser := parsing.New(phpv)
	// TODO: not ideal, temporary
	stubParser := parsing.New(phpversion.EightOne())

	files := cache.New[string, *file](contentCacheCapacity)
	irs := cache.New[string, *ast.Root](irCacheCapacity)

	t := time.NewTicker(time.Second * 60)
	go func() {
		for {
			<-t.C
			log.Printf("File cache: %s", files.String())
			log.Printf("IR cache: %s", irs.String())
		}
	}()

	return &wrkspc{
		normalParser:    normalParser,
		stubParser:      stubParser,
		roots:           []string{root, stubs},
		fileExtensions:  config.Current.FileExtensions(),
		ignoredDirNames: config.Current.IgnoredDirNames(),
		files:           files,
		irs:             irs,
	}
}

type wrkspc struct {
	normalParser    parsing.Parser
	stubParser      parsing.Parser
	roots           []string
	fileExtensions  []string
	ignoredDirNames []string
	files           *cache.Cache[string, *file]
	irs             *cache.Cache[string, *ast.Root]
}

func (w *wrkspc) Root() string {
	return w.roots[0]
}

// TODO: error handling.
func (w *wrkspc) Index(
	files chan<- *ParsedFile,
	total *atomic.Uint64,
	totalDone chan<- bool,
) error {
	defer close(files)

	go func() {
		if err := w.walk(func(_ string, d fs.DirEntry) error {
			total.Add(1)
			return nil
		}); err != nil {
			panic(err)
		}

		totalDone <- true
		close(totalDone)
	}()

	g := errgroup.Group{}
	g.SetLimit(indexGoRoutinesLimit)

	if err := w.walk(func(path string, d fs.DirEntry) error {
		g.Go(func() error {
			content, err := w.parser(path).Read(path)
			if err != nil {
				return fmt.Errorf(ErrReadFmt, path, err)
			}

			files <- &ParsedFile{Path: path, Content: content}
			return nil
		})

		return nil
	}); err != nil {
		panic(err)
	}

	if err := g.Wait(); err != nil {
		panic(err)
	}

	return nil
}

// ContentOf returns the content of the file at the given path.
func (w *wrkspc) ContentOf(path string) (string, error) {
	file := w.files.Cached(path, func() *file {
		what.Happens("Getting fresh content of %s", path)

		content, err := w.parser(path).Read(path)
		if err != nil {
			log.Println(err)
			return nil
		}

		return newFile(content)
	})

	if file == nil {
		return "", ErrFileNotIndexed
	}

	return file.content, nil
}

// FContentOf returns the content of the file at the given path.
// If an error occurs, it logs it and returns an empty string.
// Use ContentOf for access to the error.
func (w *wrkspc) FContentOf(path string) string {
	content, err := w.ContentOf(path)
	if err != nil {
		log.Println(err)
	}

	return content
}

// IROf returns the parsed root node of the given path.
// If an error occurs, it returns an empty root node, this NEVER returns nil.
func (w *wrkspc) IROf(path string) (*ast.Root, error) {
	root := w.irs.Cached(path, func() *ast.Root {
		what.Happens("Getting fresh AST of %s", path)

		content, err := w.ContentOf(path)
		if err != nil {
			log.Println(err)
			return nil
		}

		root, err := w.parser(path).Parse([]byte(content))
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

// FIROf returns the root of the given path, if an error occurs, it logs it
// and returns an empty root node. This never returns nil.
// Use IROf for access to the error.
func (w *wrkspc) FIROf(path string) *ast.Root {
	root, err := w.IROf(path)
	if err != nil {
		log.Println(err)
	}

	return root
}

// AllOf returns the content & parsed root node of the given path.
// If an error occurs, it returns an empty root node, this NEVER returns nil.
func (w *wrkspc) AllOf(path string) (string, *ast.Root, error) {
	content, cErr := w.ContentOf(path)
	root, rErr := w.IROf(path)

	if cErr != nil {
		return content, root, cErr
	}

	if rErr != nil {
		return content, root, rErr
	}

	return content, root, nil
}

// FAllOf returns both the content and root node of the given path.
// If an error occurs, it returns an empty root node and string after logging the error.
func (w *wrkspc) FAllOf(path string) (string, *ast.Root) {
	content, root, err := w.AllOf(path)
	if err != nil {
		log.Println(err)
	}

	return content, root
}

func (w *wrkspc) Refresh(path string) error {
	content, err := w.parser(path).Read(path)
	if err != nil {
		return fmt.Errorf(ErrReadFmt, path, err)
	}

	return w.RefreshFrom(path, content)
}

func (w *wrkspc) RefreshFrom(path string, content string) error {
	root, err := w.parser(path).Parse([]byte(content))
	if err != nil {
		return fmt.Errorf(ErrParseFmt, path, err)
	}

	file := newFile(content)

	w.files.Put(path, file)
	w.irs.Put(path, root)

	return nil
}

func (w *wrkspc) walk(walker func(path string, d fs.DirEntry) error) error {
	wg := sync.WaitGroup{}
	wg.Add(len(w.roots))

	var finalErr error

	for _, root := range w.roots {
		go func(root string) {
			defer wg.Done()

			if err := w.parser(root).Walk(root, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				doParse, err := w.shouldParse(d)
				if err != nil {
					return err
				}

				if doParse {
					return walker(path, d)
				}

				return nil
			}); err != nil {
				log.Println(fmt.Errorf(ErrParseWalkFmt, err))
				finalErr = err
			}
		}(root)
	}

	wg.Wait()
	return finalErr
}

func (w *wrkspc) shouldParse(d fs.DirEntry) (bool, error) {
	if d.IsDir() {
		// Skip ignored directories.
		n := d.Name()
		for _, ignored := range w.ignoredDirNames {
			if n == ignored {
				return false, filepath.SkipDir
			}
		}

		return false, nil
	}

	name := d.Name()
	for _, extension := range w.fileExtensions {
		if strings.HasSuffix(name, extension) {
			return true, nil
		}
	}

	return false, nil
}

var stubsDir = filepath.Join(pathutils.Root(), "third_party", "phpstorm-stubs")

func (w *wrkspc) parser(path string) parsing.Parser {
	if strings.HasPrefix(path, config.Current.StubsDir()) || strings.HasPrefix(path, stubsDir) {
		return w.stubParser
	}

	return w.normalParser
}
