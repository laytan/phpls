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
	"time"

	"appliedgo.net/what"
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/config"
	"github.com/laytan/elephp/internal/parsing"
	"github.com/laytan/elephp/pkg/cache"
	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/samber/do"
	"golang.org/x/sync/errgroup"
)

const (
	ErrParseFmt     = "Error parsing file into syntax nodes: %w"
	ErrReadFmt      = "Error reading file %s: %w"
	ErrParseWalkFmt = "Error walking the workspace files: %w"

	irCacheCapacity      = 25
	contentCacheCapacity = 50
)

var (
	ErrFileNotIndexed    = errors.New("File is not indexed in the workspace")
	indexGoRoutinesLimit = runtime.NumCPU()
	stubsPath            = path.Join(pathutils.Root(), "phpstorm-stubs")
	Config               = func() config.Config { return do.MustInvoke[config.Config](nil) }
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

	files := cache.New[string, *file](contentCacheCapacity)
	irs := cache.New[string, *ir.Root](irCacheCapacity)

	t := time.NewTicker(time.Second * 60)
	go func() {
		for {
			<-t.C
			log.Printf("File cache: %s", files.String())
			log.Printf("IR cache: %s", irs.String())
		}
	}()

	return &wrkspc{
		normalParser:   normalParser,
		stubParser:     stubParser,
		roots:          []string{root, path.Join(pathutils.Root(), "phpstorm-stubs")},
		fileExtensions: Config().FileExtensions(),
		files:          files,
		irs:            irs,
	}
}

type wrkspc struct {
	normalParser   parsing.Parser
	stubParser     parsing.Parser
	roots          []string
	fileExtensions []string
	files          *cache.Cache[string, *file]
	irs            *cache.Cache[string, *ir.Root]
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
				return err
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

func (w *wrkspc) IROf(path string) (*ir.Root, error) {
	root := w.irs.Cached(path, func() *ir.Root {
		what.Happens("Getting fresh IR of %s", path)

		content, err := w.ContentOf(path)
		if err != nil {
			log.Println(err)
			return nil
		}

		root, err := w.parser(path).Parse(content)
		if err != nil {
			log.Println(err)
			return nil
		}

		return root
	})

	if root == nil {
		return nil, ErrFileNotIndexed
	}

	return root, nil
}

func (w *wrkspc) AllOf(path string) (string, *ir.Root, error) {
	content, err := w.ContentOf(path)
	if err != nil {
		return "", nil, err
	}

	root, err := w.IROf(path)
	if err != nil {
		return "", nil, fmt.Errorf(ErrParseFmt, err)
	}

	return content, root, nil
}

func (w *wrkspc) Refresh(path string) error {
	content, err := w.parser(path).Read(path)
	if err != nil {
		return fmt.Errorf(ErrReadFmt, path, err)
	}

	return w.RefreshFrom(path, content)
}

func (w *wrkspc) RefreshFrom(path string, content string) error {
	root, err := w.parser(path).Parse(content)
	if err != nil {
		return err
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
