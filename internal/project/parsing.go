package project

import (
	"errors"
	"fmt"
	"log"
	"runtime"
	"sync"
	"sync/atomic"

	"appliedgo.net/what"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/strutil"
)

// This should only be called once at the beginning of the connection with a
// client.
func (p *Project) Parse(done *atomic.Uint64, total *atomic.Uint64, totalDone chan<- bool) error {
	// Parsing creates alot of garbage, after parsing, run a gc cycle manually
	// because we know there is a lot to clean up.
	defer func() {
		go runtime.GC()
	}()

	hasErrors := false

	files := make(chan *wrkspc.ParsedFile)
	wg := sync.WaitGroup{}

	// Need to have a way to know when all waits have been added to the wait
	// group, go reports a race condition when you at to a wait group after
	// already waiting for it.
	wgDone := make(chan bool, 1)

	go func() {
		defer func() { wgDone <- true }()

		for file := range files {
			wg.Add(1)
			go func(file *wrkspc.ParsedFile) {
				defer done.Add(1)
				defer wg.Done()

				if err := Index().Index(file.Path, file.Content); err != nil {
					log.Println(
						fmt.Errorf("Could not index the symbols in %s: %w", file.Path, err),
					)
					hasErrors = true
				}
			}(file)
		}
	}()

	if err := Wrkspc().Index(files, total, totalDone); err != nil {
		log.Println(
			fmt.Errorf(
				"Could not index the file content of root %s: %w",
				Wrkspc().Root(),
				err,
			),
		)
		hasErrors = true
	}

	<-wgDone
	wg.Wait()

	if hasErrors {
		return fmt.Errorf(
			"Parsing the project resulted in errors, check the logs for more details",
		)
	}

	return nil
}

func (p *Project) ParseWithoutProgress() error {
	done := &atomic.Uint64{}
	total := &atomic.Uint64{}
	totalDone := make(chan bool, 1)

	return p.Parse(done, total, totalDone)
}

func (p *Project) ParseFileUpdate(path string, content string) error {
	prevContent, err := Wrkspc().ContentOf(path)
	if err != nil && !errors.Is(err, wrkspc.ErrFileNotIndexed) {
		return fmt.Errorf("Could not read content of %s while updating: %w", path, err)
	}

	if err != nil {
		if strutil.RemoveWhitespace(content) == strutil.RemoveWhitespace(prevContent) {
			what.Happens("Skipping file update (only whitespace change) of %s", path)
			return nil
		}
	}

	// NOTE: order is important here.
	if err := Wrkspc().RefreshFrom(path, content); err != nil {
		return fmt.Errorf("Could not refresh indexed content of %s: %w", path, err)
	}

	if err := Index().Refresh(path, content); err != nil {
		return fmt.Errorf("Could not refresh indexed symbols of %s: %w", path, err)
	}

	return nil
}
