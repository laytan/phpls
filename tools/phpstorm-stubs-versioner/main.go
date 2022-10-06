package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/conf"
	"github.com/VKCOM/php-parser/pkg/errors"
	"github.com/VKCOM/php-parser/pkg/parser"
	"github.com/VKCOM/php-parser/pkg/version"
	"github.com/VKCOM/php-parser/pkg/visitor/printer"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/tools/phpstorm-stubs-versioner/pkg/transformer"
	"golang.org/x/sync/errgroup"
)

var (
	in           = "/Users/laytan/projects/elephp/third_party/phpstorm-stubs"
	out          = "/Users/laytan/projects/elephp/versioned-phpstorm-stubs"
	limit        = runtime.NumCPU()
	genVersion   = &phpversion.PHPVersion{Major: 7, Minor: 0}
	transformers = []Transformer{
		transformer.NewAtSinceAtRemoved(genVersion),
		transformer.NewElementAvailableAttribute(genVersion),
	}

	// Parsing is done with the latest version of the parser, because this parses the stubs.
	parserConfig = conf.Config{
		Version: &version.Version{Major: 8, Minor: 1},
		ErrorHandlerFunc: func(e *errors.Error) {
			panic(e)
		},
	}
)

type Transformer interface {
	Transform(ast ast.Vertex)
}

func main() {
	g := errgroup.Group{}
	g.SetLimit(limit)

	if err := os.RemoveAll(filepath.Join(out, genVersion.String())); err != nil {
		panic(err)
	}

	if err := filepath.WalkDir(in, func(path string, d fs.DirEntry, err error) error {
		// Directories need to be created before transformed files are written,
		// So we can't do this in the g.Go call because of race conditions.
		if d.IsDir() {
			if err := os.MkdirAll(outPath(path, genVersion.String()), 0o755); err != nil {
				return fmt.Errorf("os.MkDirAll(%s, nil, %d): %w", path, fs.ModeDir, err)
			}
		}

		g.Go(func() error {
			if err != nil {
				return err
			}

			if !strings.HasSuffix(path, ".php") {
				return nil
			}

			if err := transform(path); err != nil {
				return fmt.Errorf("transform(%s): %w", path, err)
			}

			return nil
		})

		return nil
	}); err != nil {
		panic(err)
	}

	if err := g.Wait(); err != nil {
		panic(err)
	}
}

func transform(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("os.ReadFile(%s): %w", path, err)
	}

	ast, err := parser.Parse(content, parserConfig)
	if err != nil {
		return fmt.Errorf("parser.Parse(...): %w", err)
	}

	for _, transformer := range transformers {
		transformer.Transform(ast)
	}

	file, err := os.Create(outPath(path, genVersion.String()))
	if err != nil {
		return fmt.Errorf("os.Create(%s): %w", outPath(path, genVersion.String()), err)
	}

	defer file.Close()

	p := printer.NewPrinter(file)
	ast.Accept(p)

	return nil
}

func outPath(path string, version string) string {
	relPath := strings.TrimPrefix(path, in)
	return filepath.Join(out, version, relPath)
}
