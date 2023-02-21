package stubtransform

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/conf"
	"github.com/VKCOM/php-parser/pkg/errors"
	"github.com/VKCOM/php-parser/pkg/parser"
	"github.com/VKCOM/php-parser/pkg/version"
	"github.com/VKCOM/php-parser/pkg/visitor/printer"
	"github.com/laytan/elephp/pkg/phpversion"
	"golang.org/x/sync/errgroup"
)

var (
	// Fastest based on benchmarking.
	MaxConcurrency = 14

	NonStubs = map[string]struct{}{
		filepath.Join(string(os.PathSeparator), ".github"):                          {},
		filepath.Join(string(os.PathSeparator), ".idea"):                            {},
		filepath.Join(string(os.PathSeparator), "FFI", ".phpstorm.meta.php"):        {},
		filepath.Join(string(os.PathSeparator), "meta"):                             {},
		filepath.Join(string(os.PathSeparator), "Reflection", ".phpstorm.meta.php"): {},
		filepath.Join(string(os.PathSeparator), "tests"):                            {},
		filepath.Join(string(os.PathSeparator), ".php-cs-fixer.php"):                {},
		filepath.Join(string(os.PathSeparator), "PhpStormStubsMap.php"):             {},
		filepath.Join(string(os.PathSeparator), ".phpstorm.meta.php"):               {},

		// Special case, this file contains an enum which the parser should support,
		// but it doesn't, TODO: see what's up.
		filepath.Join(string(os.PathSeparator), "relay", "KeyType.php"): {},
	}

	// The version to parse the stubs with.
	parserVersion = &version.Version{
		Major: 8,
		Minor: 1,
	}
)

type Logger interface {
	Printf(format string, args ...any)
}

// logger is allowed to be nil.
func All(version *phpversion.PHPVersion, logger Logger) func() []ast.Visitor {
	return func() []ast.Visitor {
		return []ast.Visitor{
			NewAtSinceAtRemoved(version, logger),
			NewElementAvailableAttribute(version, logger),
			NewLanguageLevelTypeAware(version, logger),
		}
	}
}

type Walker struct {
	Transformers *sync.Pool
	Logger       Logger
	StubsDir     string
	OutDir       string
	Version      *phpversion.PHPVersion
	Progress     *atomic.Uint32
}

func NewWalker(
	logger Logger,
	stubsDir string,
	outDir string,
	version *phpversion.PHPVersion,
	transformers func() []ast.Visitor,
) *Walker {
	return &Walker{
		Transformers: &sync.Pool{
			New: func() any {
				return transformers()
			},
		},
		Logger:   logger,
		StubsDir: stubsDir,
		OutDir:   outDir,
		Version:  version,
	}
}

func (w *Walker) Walk() error {
	g := errgroup.Group{}
	g.SetLimit(MaxConcurrency)

	if err := filepath.WalkDir(w.StubsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath := strings.TrimPrefix(path, w.StubsDir)
		if _, ok := NonStubs[relPath]; ok {
			if d.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		if !d.IsDir() && !strings.HasSuffix(path, ".php") {
			return nil
		}

		finalPath := outPath(w.StubsDir, w.OutDir, path, w.Version.String())

		// Directories need to be created before transformed files are written,
		// So we can't do this in the g.Go call because of race conditions.
		if d.IsDir() {
			if err := os.MkdirAll(finalPath, 0o755); err != nil {
				return fmt.Errorf("creating directories towards %s: %w", finalPath, err)
			}

			return nil
		}

		g.Go(func() error {
			if w.Progress != nil {
				defer w.Progress.Add(1)
			}

			transformers := w.Transformers.Get().([]ast.Visitor)
			if err := TransformFile(w.Logger, transformers, path, finalPath); err != nil {
				return fmt.Errorf("transforming %s: %w", path, err)
			}

			return nil
		})

		return nil
	}); err != nil {
		return fmt.Errorf("walking stubs %s: %w", w.StubsDir, err)
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("waiting for transformations to complete: %w", err)
	}

	return nil
}

func TransformFile(logger Logger, transformers []ast.Visitor, path string, finalPath string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	ast, err := parser.Parse(content, conf.Config{
		Version: parserVersion,
		ErrorHandlerFunc: func(e *errors.Error) {
			logger.Printf(
				"Error parsing into AST, path: %s, message: %s, line: %d",
				path,
				e.Msg,
				e.Pos.StartLine,
			)
		},
	})
	if err != nil {
		return fmt.Errorf("parsing %s into ast: %w", path, err)
	}

	for _, transformer := range transformers {
		ast.Accept(transformer)
	}

	file, err := os.Create(finalPath)
	if err != nil {
		return fmt.Errorf("creating out path %s: %w", finalPath, err)
	}

	defer file.Close()

	// TODO: can we get this to work in php 8?
	// f := formatter.NewFormatter()
	// ast.Accept(f)

	writer := bufio.NewWriter(file)
	p := printer.NewPrinter(writer)
	ast.Accept(p)
	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("writing remaining buffer: %w", err)
	}

	return nil
}

func outPath(stubsDir string, outDir string, path string, version string) string {
	relPath := strings.TrimPrefix(path, stubsDir)
	return filepath.Join(outDir, version, relPath)
}
