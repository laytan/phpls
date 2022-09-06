package wrkspc

import (
	"errors"
	"fmt"
	"io/fs"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/parsing"
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
	startingIndexSize = 2000
)

var ErrFileNotIndexed = errors.New("File is not indexed in the workspace")

var indexGoRoutinesLimit = runtime.NumCPU()

type file struct {
	content string
}

func newFile(content string) *file {
	return &file{
		content: content,
	}
}

type ParsedFile struct {
	path    string
	content string
}

type Wrkspc interface {
	Root() string

	FileExtensions() []string

	Index(files chan<- *ParsedFile, total *atomic.Uint64, totalDone chan<- bool) error

	Has(path string) bool

	ContentOf(path string) (string, error)

	IROf(path string) (*ir.Root, error)

	Refresh(path string) error

	RefreshFrom(path string, content string) error
}

// fileExtensions should all start with a period.
func NewWrkspc(parser parsing.Parser, root string, fileExtensions []string) Wrkspc {
	return &wrkspc{
		parser:         parser,
		root:           root,
		fileExtensions: fileExtensions,
		files:          make(map[string]*file, startingIndexSize),
	}
}

type wrkspc struct {
	parser parsing.Parser

	root string

	fileExtensions []string

	files      map[string]*file
	filesMutex sync.RWMutex
}

func (w *wrkspc) Root() string {
	return w.root
}

func (w *wrkspc) FileExtensions() []string {
	return w.fileExtensions
}

// TODO: error handling
func (w *wrkspc) Index(
	files chan<- *ParsedFile,
	total *atomic.Uint64,
	totalDone chan<- bool,
) error {
	go func() {
		if err := w.walk(func(_ string, d fs.DirEntry) error {
			total.Add(1)
			return nil
		}); err != nil {
			panic(err)
		}

		totalDone <- true
	}()

	g := errgroup.Group{}
	g.SetLimit(indexGoRoutinesLimit)

	if err := w.walk(func(path string, d fs.DirEntry) error {
		g.Go(func() error {
			file, err := w.refresh(path)
			if err != nil {
				return err
			}

			files <- &ParsedFile{path: path, content: file.content}

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

	ir, err := w.parser.Parse(content)
	if err != nil {
		// TODO: is this the proper way to wrap an error with fmt (does errors.Is work)?
		return nil, fmt.Errorf(ErrParseFmt, err)
	}

	return ir, nil
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
	content, err := w.parser.Read(path)
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
	if err := w.parser.Walk(w.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if w.shouldParse(d) {
			return walker(path, d)
		}

		return nil
	}); err != nil {
		return fmt.Errorf(ErrParseWalkFmt, err)
	}

	return nil
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
