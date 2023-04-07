package diagnostics

// TODO: add configuration, enable or disable analyzers, set if they run on change or on save, provide php binary, provide analyzer binary.
// TODO: if running on save, configure phpstan to not create a temporary file.

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
)

// Implement closer to do clean up (stopping a daemon for example).
type Closer interface {
	Close()
}

type Analyzer interface {
	Name() string
	Analyze(ctx context.Context, path string, code []byte) ([]protocol.Diagnostic, error)
}

type Runner struct {
	client      protocol.Client
	diagnostics map[string]*fileDiagnostics
	analyzers   []Analyzer
}

type fileDiagnostics struct {
	path        string
	diagnostics []protocol.Diagnostic
	version     int
	published   bool
	cancel      context.CancelFunc
}

func NewRunner(client protocol.Client, analyzers []Analyzer) *Runner {
	return &Runner{
		client:      client,
		diagnostics: make(map[string]*fileDiagnostics),
		analyzers:   analyzers,
	}
}

func (r *Runner) SetupPath(path string, version int) {
	if _, ok := r.diagnostics[path]; ok {
		return
	}

	r.diagnostics[path] = &fileDiagnostics{
		path:      path,
		version:   version,
		published: true,
	}
}

func (r *Runner) CancelPrevRun(ctx context.Context, path string) context.Context {
	fd := r.diagnostics[path]
	if fd.cancel != nil {
		log.Println("[INFO]: cancelling previous run")
		fd.cancel()
	}

	newCtx, cancelFunc := context.WithCancel(ctx)
	fd.cancel = cancelFunc
	return newCtx
}

// TODO: cancel diagnostics for earlier versions of this file.
func (r *Runner) Run(ctx context.Context, version int, path string, code []byte) (reserr error) {
	logTime := timeDiagnostics("all")
	defer logTime()
	_, filename := filepath.Split(path)
	log.Printf("[INFO]: running analyzers for %q version %d", filename, version)

	r.SetupPath(path, version)
	ctx = r.CancelPrevRun(ctx, path)

	errorC := make(chan error)
	resultsC := make(chan []protocol.Diagnostic)
	for _, a := range r.analyzers {
		go func(a Analyzer) {
			logTime := timeDiagnostics(a.Name())
			defer logTime()

			res, err := a.Analyze(ctx, path, code)
			if err != nil {
				errorC <- fmt.Errorf("%s: %w", a.Name(), err)
				return
			}

			resultsC <- normalizeDiagnostics(a.Name(), res)
		}(a)
	}

	timer := time.NewTicker(time.Millisecond * 50)
	defer timer.Stop()
	waitingFor := len(r.analyzers)
	var errs []error
Loop:
	for {
		select {
		case err := <-errorC:
			errs = append(errs, err)

			waitingFor--
			if waitingFor == 0 {
				if err := r.Publish(ctx, path); err != nil {
					errs = append(errs, err)
				}
				break Loop
			}
		case res := <-resultsC:
			r.AddDiagnostics(path, version, res)

			waitingFor--
			if waitingFor == 0 {
				if err := r.Publish(ctx, path); err != nil {
					errs = append(errs, err)
				}
				break Loop
			}
		case <-timer.C:
			if err := r.Publish(ctx, path); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) == 0 {
		return nil
	}

	if len(errs) == 1 {
		return errs[0]
	}

	return fmt.Errorf("multiple errors running diagnostics for %q: %w", filename, errors.Join(errs...))
}

func (r *Runner) AddDiagnostics(path string, version int, diagnostics []protocol.Diagnostic) {
	fd := r.diagnostics[path]

	if version == fd.version {
		if len(diagnostics) == 0 {
			return
		}

		fd.diagnostics = append(fd.diagnostics, diagnostics...)
		fd.published = false
		return
	}

	if version > fd.version {
		fd.diagnostics = diagnostics
		fd.published = false
		fd.version = version
		return
	}

	log.Printf("[WARN]: trying to add outdated diagnostics to %q", path)
}

func (r *Runner) Publish(ctx context.Context, path string) error {
	if fd, ok := r.diagnostics[path]; ok {
		if fd.published {
			return nil
		}
		fd.published = true

		log.Printf("[INFO]: publishing diagnostic update: %d diagnostics", len(fd.diagnostics))

		if err := r.client.PublishDiagnostics(ctx, &protocol.PublishDiagnosticsParams{
			URI:         protocol.DocumentURI("file://" + path),
			Version:     int32(fd.version),
			Diagnostics: fd.diagnostics,
		}); err != nil {
			return fmt.Errorf("publishing diagnostics: %w", err)
		}
	}

	return nil
}

func (r *Runner) Close() {
	for _, a := range r.analyzers {
		if closer, ok := a.(Closer); ok {
			closer.Close()
		}
	}
}

func timeDiagnostics(name string) func() {
	start := time.Now()
	return func() {
		log.Printf("[INFO]: %s diagnostics took %s", name, time.Since(start))
	}
}

// TODO: consistent casing and consistent ending . or not.
func normalizeDiagnostics(name string, diagnostics []protocol.Diagnostic) []protocol.Diagnostic {
	for i := range diagnostics {
		diagnostics[i].Source = "elephp-" + name
		diagnostics[i].Message = fmt.Sprintf("[%s] %s", name, diagnostics[i].Message)
	}

	return diagnostics
}
