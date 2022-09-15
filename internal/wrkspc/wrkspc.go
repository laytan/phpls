package wrkspc

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"path"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"appliedgo.net/what"
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/config"
	"github.com/laytan/elephp/internal/parsing"
	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/samber/do"
	"golang.org/x/sync/errgroup"
)

const (
	ErrParseFmt     = "Error parsing file into syntax nodes: %w"
	ErrReadFmt      = "Error reading file %s: %w"
	ErrParseWalkFmt = "Error walking the workspace files: %w"

	// Starting with a set capacity for the map increases index times by a lot.
	// But there is a balance to be struck, if a project only has 3 files,
	// The index will still be of the high capacity (more memory).
	// But larger projects are more common.
	startingIndexSize = 10000
)

var ErrFileNotIndexed = errors.New("File is not indexed in the workspace")

var indexGoRoutinesLimit = runtime.NumCPU()

var stubsPath = path.Join(pathutils.Root(), "phpstorm-stubs")

var Config = func() config.Config { return do.MustInvoke[config.Config](nil) }

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

	Has(path string) bool

	ContentOf(path string) (string, error)

	IROf(path string) (*ir.Root, error)

	AllOf(path string) (string, *ir.Root, error)

	Refresh(path string) error

	RefreshFrom(path string, content string) error
}

// fileExtensions should all start with a period.
func New(phpv *phpversion.PHPVersion, root string) Wrkspc {
	normalParser := parsing.New(phpv)
	// TODO: not ideal, temporary
	stubParser := parsing.New(phpversion.EightOne())

	return &wrkspc{
		normalParser:   normalParser,
		stubParser:     stubParser,
		roots:          []string{root, path.Join(pathutils.Root(), "phpstorm-stubs")},
		fileExtensions: Config().FileExtensions(),
		files:          make(map[string]*file, startingIndexSize),
	}
}

type wrkspc struct {
	normalParser parsing.Parser
	stubParser   parsing.Parser

	roots []string

	fileExtensions []string

	files      map[string]*file
	filesMutex sync.RWMutex
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
			file, err := w.refresh(path)
			if err != nil {
				return err
			}

			files <- &ParsedFile{Path: path, Content: file.content}

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

func (w *wrkspc) Has(path string) bool {
	w.filesMutex.RLock()
	defer w.filesMutex.RUnlock()

	_, ok := w.files[path]

	return ok
}

func (w *wrkspc) ContentOf(path string) (string, error) {
	w.filesMutex.RLock()
	defer w.filesMutex.RUnlock()

	what.Happens("Getting content of %s", path)

	file, ok := w.files[path]
	if !ok {
		return "", ErrFileNotIndexed
	}

	return file.content, nil
}

// TODO: arccache.
func (w *wrkspc) IROf(path string) (*ir.Root, error) {
	content, err := w.ContentOf(path)
	if err != nil {
		return nil, err
	}

	what.Happens("Getting IR of %s", path)

	ir, err := w.parser(path).Parse(content)
	if err != nil {
		// TODO: is this the proper way to wrap an error with fmt (does errors.Is work)?
		return nil, fmt.Errorf(ErrParseFmt, err)
	}

	return ir, nil
}

func (w *wrkspc) AllOf(path string) (string, *ir.Root, error) {
	content, err := w.ContentOf(path)
	if err != nil {
		return "", nil, err
	}

	ir, err := w.parser(path).Parse(content)
	if err != nil {
		return "", nil, fmt.Errorf(ErrParseFmt, err)
	}

	return content, ir, nil
}

func (w *wrkspc) Refresh(path string) error {
	_, err := w.refresh(path)
	return err
}

func (w *wrkspc) RefreshFrom(path string, content string) error {
	w.refreshFrom(path, content)
	return nil
}

func (w *wrkspc) refresh(path string) (*file, error) {
	content, err := w.parser(path).Read(path)
	if err != nil {
		return nil, fmt.Errorf(ErrReadFmt, path, err)
	}

	return w.refreshFrom(path, content), nil
}

func (w *wrkspc) refreshFrom(path string, content string) *file {
	w.filesMutex.Lock()
	defer w.filesMutex.Unlock()

	file := newFile(content)
	w.files[path] = file

	return file
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

				if w.shouldParse(d) {
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

func (w *wrkspc) shouldParse(d fs.DirEntry) bool {
	if d.IsDir() {
		return false
	}

	name := d.Name()
	for _, extension := range w.fileExtensions {
		if strings.HasSuffix(name, extension) {
			return true
		}
	}

	return false
}

func (w *wrkspc) parser(path string) parsing.Parser {
	if strings.HasPrefix(path, stubsPath) {
		return w.stubParser
	}

	return w.normalParser
}
