//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/laytan/elephp/pkg/stubs/stubtransform"
)

var fileTempl = template.Must(template.New("").Parse(`// Code generated by go generate; DO NOT EDIT.

package stubs

// The total amount of stub files.
// Retrieved from phpstorm-stubs version {{ .StubsVersion }} in the directory {{ .StubsDir }}.
const TotalStubs = {{ .TotalStubs }}
`))

var (
	verRgx   = regexp.MustCompile(`\((.*)\)`)
	stubsDir = filepath.Join("..", "..", "third_party", "phpstorm-stubs")
)

type TemplData struct {
	StubsVersion string
	TotalStubs   int
	StubsDir     string
}

// main generates the TotalStubs counter.
func main() {
	count := 0

	// Walk all stubs, incrementing count for each stub file.
	err := filepath.WalkDir(stubsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("WalkDir error: %w", err)
		}

		// Ignore non-stubs as defined in the stubtransform package.
		relPath := strings.TrimPrefix(path, stubsDir)
		if _, ok := stubtransform.NonStubs[relPath]; ok {
			if d.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		// Ignore non-php files.
		if !d.IsDir() && !strings.HasSuffix(path, ".php") {
			return nil
		}

		count++
		return nil
	})
	if err != nil {
		panic(fmt.Errorf("could not count total stubs: %w", err))
	}

	verOut, err := getStubsVersion()
	if err != nil {
		panic(err)
	}

	f, err := os.Create("total.go")
	if err != nil {
		panic(fmt.Errorf("could not create total.go: %w", err))
	}

	defer f.Close()

	fileTempl.Execute(f, TemplData{
		StubsVersion: string(verOut),
		TotalStubs:   count,
		StubsDir:     stubsDir,
	})
}

func getStubsVersion() (string, error) {
	cmd := exec.Command("git", "submodule", "status", stubsDir)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("could not get submodule version: %w", err)
	}

	matches := verRgx.FindSubmatch(out)
	if len(matches) != 2 {
		return "", fmt.Errorf("invalid matches on stub version regex: %v", matches)
	}

	return string(matches[1]), nil
}