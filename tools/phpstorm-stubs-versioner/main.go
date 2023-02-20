package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/tools/phpstorm-stubs-versioner/pkg/transformer"
)

const (
	latestParserMajor = 8
	latestParserMinor = 1
)

func main() {
	var versionStr string
	var in string
	var out string

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

	_, _ = fmt.Printf(
		"\nUsing PHP version \"%s\"\nUsing input path \"%s\"\nUsing output path \"%s\"\n\n",
		phpv.String(),
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

	l := &logger{}

	w := transformer.NewWalker(l, in, out, phpv, transformer.All(phpv, l))
	err = w.Go()
	if err != nil {
		panic(err)
	}
}

type logger struct{}

func (l *logger) Printf(format string, args ...any) {
	_, _ = fmt.Printf(format, args...)
}
