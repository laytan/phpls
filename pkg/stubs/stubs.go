package stubs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/laytan/phpls/pkg/phpversion"
	"github.com/laytan/phpls/pkg/stubs/stubtransform"
	thirdparty "github.com/laytan/phpls/third_party"
)

// Generate `total.go`.
//go:generate go run total_gen.go

var ErrNotExists = os.ErrNotExist

// Path returns the path to the root of the stubs directory for the given version.
// Returns ErrNotExists if the directory does not exists.
// Call Generate to generate the stubs.
func Path(root string, version *phpversion.PHPVersion) (string, error) {
	stubsPath := filepath.Join(root, version.String())
	_, err := os.Lstat(stubsPath)
	if err != nil {
		return "", fmt.Errorf("checking if stub dir at %s exists: %w", stubsPath, err)
	}

	return stubsPath, nil
}

// Generate generates the stubs for the given version in the given root.
// It returns the root directory of the stubs.
// Optionally pass a atomic uint to keep track of the number of files that are done being generated.
// You can compare that with stubs.TotalStubs to keep track of progress.
func Generate(
	root string,
	version *phpversion.PHPVersion,
	progress *atomic.Uint32,
) (string, error) {
	stubsPath := filepath.Join(root, version.String())
	w := stubtransform.NewWalker(
		log.Default(),
		".",
		root,
		version,
		stubtransform.All(version, nil),
	)

	w.Progress = progress
	w.StubsFS = thirdparty.Stubs

	err := w.Walk()
	if err != nil {
		return "", fmt.Errorf("walking and generating stubs: %w", err)
	}

	return stubsPath, nil
}
