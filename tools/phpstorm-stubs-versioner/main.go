package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/VKCOM/php-parser/pkg/conf"
	"github.com/VKCOM/php-parser/pkg/errors"
	"github.com/VKCOM/php-parser/pkg/parser"
	"github.com/VKCOM/php-parser/pkg/version"
	"github.com/VKCOM/php-parser/pkg/visitor/printer"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/tools/phpstorm-stubs-versioner/pkg/transformer"
	"golang.org/x/sync/errgroup"
)

const (
	goroutines = 4

	latestParserMajor = 8
	latestParserMinor = 1
)

var (
	in, out       string
	genVersion    *phpversion.PHPVersion
	transformers  []transformer.Transformer
	parserVersion = &version.Version{
		Major: latestParserMajor,
		Minor: latestParserMinor,
	}
)

func main() {
	var versionStr string

	flag.StringVar(
		&in,
		"in",
		"./third_party/phpstorm-stubs",
		"Path to the original phpstorm-stubs",
	)
	flag.StringVar(
		&out,
		"out",
		"./versioned-phpstorm-stubs",
		"Path to use as output",
	)
	flag.StringVar(
		&versionStr,
		"version",
		fmt.Sprintf("%d.%d.0", latestParserMajor, latestParserMinor),
		"The PHP version to parse the stubs for",
	)

	flag.Parse()

	phpv, ok := phpversion.FromString(versionStr)
	if !ok {
		_, _ = fmt.Printf("Invalid PHP version: %s\n", versionStr)
		os.Exit(1)
	}

	absIn, err := filepath.Abs(in)
	if err != nil {
		_, _ = fmt.Println(err)
		os.Exit(1)
	}

	absOut, err := filepath.Abs(out)
	if err != nil {
		_, _ = fmt.Println(err)
		os.Exit(1)
	}

	in = absIn
	out = absOut

	genVersion = phpv
	transformers = []transformer.Transformer{
		transformer.NewAtSinceAtRemoved(genVersion),
		transformer.NewElementAvailableAttribute(genVersion),
		transformer.NewLanguageLevelTypeAware(genVersion),
	}

	_, _ = fmt.Printf(
		"\nUsing PHP version \"%s\"\nUsing input path \"%s\"\nUsing output path \"%s\"\n\n",
		genVersion.String(),
		in,
		out,
	)

	_, _ = fmt.Print("Continue? (y/n) ")
	var confirmation string
	_, err = fmt.Scanln(&confirmation)
	if err != nil || confirmation != "y" {
		_, _ = fmt.Println("Cancelled")
		os.Exit(0)
	}

	_, _ = fmt.Printf("\n\n")

	start := time.Now()
	defer func() {
		_, _ = fmt.Printf("\nDone in %s\n", time.Since(start))
	}()

	g := errgroup.Group{}
	g.SetLimit(goroutines)

	touched := make(map[string]bool, 1500)
	if err := filepath.WalkDir(in, func(path string, d fs.DirEntry, err error) error {
		touched[strings.TrimPrefix(path, in)] = true

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
		_, _ = fmt.Println(err)
		os.Exit(1) //nolint:gocritic // Not running the defered function is fine.
	}

	if err := g.Wait(); err != nil {
		_, _ = fmt.Println(err)
		os.Exit(1) //nolint:gocritic // Not running the defered function is fine.
	}

	_, _ = fmt.Printf("\n\n")

	// Clean up any files that haven't been touched just now.
	prefix := filepath.Join(out, genVersion.String())
	if err := filepath.WalkDir(prefix, func(path string, d fs.DirEntry, err error) error {
		relPath := strings.TrimPrefix(path, prefix)
		_, ok := touched[relPath]
		if relPath != "" && !ok {
			_, _ = fmt.Printf("Cleaned up %s\n", path)
			if err := os.RemoveAll(path); err != nil {
				return fmt.Errorf("os.RemoveAll(%s): %w", path, err)
			}
		}

		return nil
	}); err != nil {
		_, _ = fmt.Println(err)
		os.Exit(1) //nolint:gocritic // Not running the defered function is fine.
	}
}

func transform(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("os.ReadFile(%s): %w", path, err)
	}

	ast, err := parser.Parse(content, conf.Config{
		Version: parserVersion,
		ErrorHandlerFunc: func(e *errors.Error) {
			log.Printf(
				"Error parsing into AST, path: %s, message: %s, line: %d",
				path,
				e.Msg,
				e.Pos.StartLine,
			)
		},
	})
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

	// TODO: can we get this to work in php 8?
	// f := formatter.NewFormatter()
	// ast.Accept(f)

	p := printer.NewPrinter(file)
	ast.Accept(p)

	return nil
}

func outPath(path string, version string) string {
	relPath := strings.TrimPrefix(path, in)
	return filepath.Join(out, version, relPath)
}
