package project

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/laytan/phpls/internal/index"
	"github.com/laytan/phpls/internal/wrkspc"
)

// This should only be called once at the beginning of the connection with a
// client.
func (p *Project) Parse(done *atomic.Uint64, total *atomic.Uint64) error {
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

				if err := index.Current.Index(file.Path, file.Content); err != nil {
					log.Println(
						fmt.Errorf("Could not index the symbols in %s: %w", file.Path, err),
					)
					hasErrors = true
				}
			}(file)
		}
	}()

	w := wrkspc.Current
	if err := w.Walk(files, total, wrkspc.WalkAll); err != nil {
		log.Println(
			fmt.Errorf(
				"Could not index the file content of root %s: %w",
				w.Root(),
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
	return p.Parse(done, total)
}

func (p *Project) ParseFileUpdate(path string, content string) error {
	if err := index.Current.Refresh(path, content); err != nil {
		return fmt.Errorf("Could not refresh indexed symbols of %s: %w", path, err)
	}

	return nil
}
