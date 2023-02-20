package transformer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

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
	MaxConcurrency = 4

	NonStubs = []string{
		"/.github",
		"/.idea",
		"/FFI/.phpstorm.meta.php",
		"/meta",
		"/Reflection/.phpstorm.meta.php",
		"/tests",
		"/.php-cs-fixer.php",
		"/PhpStormStubsMap.php",
	}

	// The version to parse the stubs with.
	parserVersion = &version.Version{
		Major: 8,
		Minor: 1,
	}
)

type Transformer interface {
	Transform(ast ast.Vertex)
}

type Logger interface {
	Printf(format string, args ...any)
}

// logger is allowed to be nil.
func All(version *phpversion.PHPVersion, logger Logger) []Transformer {
	return []Transformer{
		NewAtSinceAtRemoved(version, logger),
		NewElementAvailableAttribute(version, logger),
		NewLanguageLevelTypeAware(version, logger),
	}
}

func Transform(
	logger Logger,
	version *phpversion.PHPVersion,
	stubsDir string,
	outDir string,
	transformers ...Transformer,
) error {
	g := errgroup.Group{}
	g.SetLimit(MaxConcurrency)

	if err := filepath.WalkDir(stubsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath := strings.TrimPrefix(path, stubsDir)
		for _, ns := range NonStubs {
			if strings.HasPrefix(relPath, ns) {
				return nil
			}
		}

		if !d.IsDir() && !strings.HasSuffix(path, ".php") {
			return nil
		}

		finalPath := outPath(stubsDir, outDir, path, version.String())

		// Directories need to be created before transformed files are written,
		// So we can't do this in the g.Go call because of race conditions.
		if d.IsDir() {
			if err := os.MkdirAll(finalPath, 0o755); err != nil {
				return fmt.Errorf("creating directories towards %s: %w", finalPath, err)
			}

			return nil
		}

		g.Go(func() error {
			if err := TransformFile(logger, transformers, path, finalPath); err != nil {
				return fmt.Errorf("transforming %s: %w", path, err)
			}

			return nil
		})

		return nil
	}); err != nil {
		return fmt.Errorf("walking stubs %s: %w", stubsDir, err)
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("waiting for transformations to complete: %w", err)
	}

	return nil
}

func TransformFile(logger Logger, transformers []Transformer, path string, finalPath string) error {
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
		transformer.Transform(ast)
	}

	file, err := os.Create(finalPath)
	if err != nil {
		return fmt.Errorf("creating out path %s: %w", finalPath, err)
	}

	defer file.Close()

	// TODO: can we get this to work in php 8?
	// f := formatter.NewFormatter()
	// ast.Accept(f)

	p := printer.NewPrinter(file)
	ast.Accept(p)

	return nil
}

func outPath(stubsDir string, outDir string, path string, version string) string {
	relPath := strings.TrimPrefix(path, stubsDir)
	return filepath.Join(outDir, version, relPath)
}
