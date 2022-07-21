package pathutils

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

var root string

// Tries to find the root of the project independently of the running conditions.
func Root() string {
	if root != "" {
		return root
	}

	wd, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("Error getting working directory: %w", err))
	}

	// If we are testing the project, the wd is that of the current test file
	// and the os.Executable is a temporary file.
	// In this case we go back 2 directories.
	// NOTE: If we have nested packages in the future this needs adjustment.
	if isInTests() {
		// NOTE: Intentionally not setting root here so it gets evaluated every
		// NOTE: call, as tests will have different wd's.
		r := path.Join(wd, "..", "..")
		log.Infof("Detected project root: %s\n", r)
		return r
	}

	// When running through `go run cmd/main.go` the os.Executable call below
	// will point to a temporary file of the binary.
	// but os.Getwd would return the root.
	if _, err := os.Stat(path.Join(wd, "go.mod")); !os.IsNotExist(err) {
		root = wd
		log.Infof("Detected project root: %s\n", root)
		return root
	}

	// We get here if we are running through anything other than `go run cmd/main.go`
	// In this case we want to get the path to the executable's directory.
	ex, err := os.Executable()
	if err != nil {
		panic(fmt.Errorf("Error getting executable path: %w", err))
	}

	ex, err = filepath.EvalSymlinks(ex)
	if err != nil {
		panic(fmt.Errorf("Error evaluating executable symlinks: %w", err))
	}

	root = path.Dir(ex)
	log.Infof("Detected project root: %s\n", root)
	return root
}

func isInTests() bool {
	return strings.HasSuffix(os.Args[0], ".test")
}
