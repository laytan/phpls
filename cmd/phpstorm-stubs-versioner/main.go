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
	"github.com/laytan/elephp/internal/transformer"
	"golang.org/x/sync/errgroup"
)

var (
	in  = "/Users/laytan/projects/elephp/third_party/phpstorm-stubs"
	out = "/Users/laytan/projects/elephp/versioned-phpstorm-stubs"
	// versions = []string{"5.4", "5.6", "7.0", "7.1", "7.4", "8.0", "8.1"}
	versions     = []string{"7.4.30"}
	limit        = runtime.NumCPU()
	transformers = []Transformer{
		&transformer.AtSinceAtRemoved{}, // Remove nodes based on @since and @removed.
	}
)

type Transformer interface {
	Transform(ast ast.Vertex, version string) ast.Vertex
}

func main() {
	g := errgroup.Group{}
	g.SetLimit(limit)

	pc := conf.Config{
		Version: &version.Version{Major: 8, Minor: 1},
		ErrorHandlerFunc: func(e *errors.Error) {
			panic(e)
		},
	}

	for _, version := range versions {
		if err := os.RemoveAll(filepath.Join(out, version)); err != nil {
			panic(err)
		}
	}

	if err := filepath.WalkDir(in, func(path string, d fs.DirEntry, err error) error {
		// Directories need to be created before transformed files are written,
		// So we can't do this in the g.Go call because of race conditions.
		if d.IsDir() {
			if err := os.MkdirAll(outPath(path, versions[0]), 0755); err != nil {
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

			if err := transform(path, versions[0], pc); err != nil {
				return fmt.Errorf("transform(%s, %s): %w", path, versions[0], err)
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

func transform(path string, version string, parserConfig conf.Config) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("os.ReadFile(%s): %w", path, err)
	}

	ast, err := parser.Parse(content, parserConfig)
	if err != nil {
		return fmt.Errorf("parser.Parse(...): %w", err)
	}

	for _, transformer := range transformers {
		ast = transformer.Transform(ast, version)
	}

	file, err := os.Create(outPath(path, version))
	if err != nil {
		return fmt.Errorf("os.Create(%s): %w", outPath(path, version), err)
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
