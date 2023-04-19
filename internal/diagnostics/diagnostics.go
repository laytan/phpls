package diagnostics

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
	"github.com/laytan/phpls/internal/config"
	"github.com/laytan/phpls/pkg/lsprogress"
	"github.com/laytan/phpls/pkg/set"
)

// Implement closer to do clean up (stopping a daemon for example).
type Closer interface {
	Close()
}

type Analyzer interface {
	Name() string
	Analyze(ctx context.Context, path string, code []byte) ([]protocol.Diagnostic, error)
	AnalyzeSave(ctx context.Context, path string) ([]protocol.Diagnostic, error)
}

type Runner struct {
	client   protocol.Client
	progress *lsprogress.Tracker

	diagnostics   map[string]*fileDiagnostics
	diagnosticsMu sync.Mutex

	analyzers     []Analyzer
	saveAnalyzers []Analyzer

	watcher    *fsnotify.Watcher
	watching   map[string]*set.Set[string]
	watchingMu sync.Mutex
}

type fileDiagnostics struct {
	path            string
	diagnostics     []protocol.Diagnostic
	saveDiagnostics []protocol.Diagnostic
	version         int
	published       bool
	cancel          context.CancelFunc
}

func NewRunner(
	client protocol.Client,
	analyzers []Analyzer,
	saveAnalyzers []Analyzer,
) *Runner {
	return &Runner{
		client:        client,
		progress:      lsprogress.NewTracker(client),
		diagnostics:   make(map[string]*fileDiagnostics),
		analyzers:     analyzers,
		saveAnalyzers: saveAnalyzers,
		watching:      make(map[string]*set.Set[string]),
	}
}

func NewRunnerFromConfig(client protocol.Client) *Runner {
	if !config.Current.Diagnostics.Enabled {
		return nil
	}

	var analyzers []Analyzer
	var saveAnalyzers []Analyzer

	if config.Current.Diagnostics.Phpcs.Enabled {
		if executable, ok := findExec(config.Current.Diagnostics.Phpcs.Binary); ok {
			analyzer := MakePhpcs(executable)
			switch config.Current.Diagnostics.Phpcs.Method {
			case config.DiagnosticsOnSave:
				log.Printf(
					"[INFO]: phpcs diagnostics set up on save using binary at %q",
					executable,
				)
				saveAnalyzers = append(saveAnalyzers, analyzer)
			case config.DiagnosticsOnChange:
				log.Printf(
					"[INFO]: phpcs diagnostics set up on change using binary at %q",
					executable,
				)
				analyzers = append(analyzers, analyzer)
			}
		} else {
			log.Printf(
				"[ERROR]: phpcs is enabled but no executable found in the following configured places: %v",
				config.Current.Diagnostics.Phpcs.Binary,
			)
		}
	}

	if config.Current.Diagnostics.Phpstan.Enabled {
		if executable, ok := findExec(config.Current.Diagnostics.Phpstan.Binary); ok {
			analyzer := &PhpstanAnalyzer{Executable: executable}
			switch config.Current.Diagnostics.Phpstan.Method {
			case config.DiagnosticsOnSave:
				log.Printf(
					"[INFO]: phpstan diagnostics set up on save with binary at %q",
					executable,
				)
				saveAnalyzers = append(saveAnalyzers, analyzer)
			case config.DiagnosticsOnChange:
				log.Printf(
					"[INFO]: phpstan diagnostics set up on change with binary at %q",
					executable,
				)
				analyzers = append(analyzers, analyzer)
			}
		} else {
			log.Printf(
				"[ERROR]: phpstan is enabled but no executable found in the following configured places: %v",
				config.Current.Diagnostics.Phpstan.Binary,
			)
		}
	}

	return NewRunner(client, analyzers, saveAnalyzers)
}

func (r *Runner) Watch(path string) error {
	// No reason to watch if we don't have any save analyzers.
	if len(r.saveAnalyzers) == 0 {
		return nil
	}

	log.Printf("[INFO]: started watching %q for changes", path)

	// On initial watch, run one round of analyzers to get initial diagnostics.
	go r.runForSave(path)

	r.watchingMu.Lock()
	defer r.watchingMu.Unlock()

	if err := r.setupWatcher(); err != nil {
		return fmt.Errorf("setting up watcher: %w", err)
	}

	dir, fn := filepath.Split(path)

	if _, ok := r.watching[dir]; ok {
		r.watching[dir].Add(fn)
		return nil
	}

	// We are watching parent directories because of 2 reasons:
	//   1. An LSP didOpen request comes in when the file has just been opened, this does not mean it has actually been saved yet.
	//   1. The fsnotify package recommends watching directories because it is more stable, some tools do funky stuff.
	if err := r.watcher.Add(dir); err != nil {
		return fmt.Errorf("adding parent directory of %q to watcher: %w", path, err)
	}

	r.watching[dir] = set.New[string]()
	r.watching[dir].Add(fn)
	return nil
}

func (r *Runner) StopWatching(path string) error {
	r.watchingMu.Lock()
	defer r.watchingMu.Unlock()

	if r.watcher == nil {
		return nil
	}

	log.Printf("[INFO]: stopped watching %q for changes", path)

	dir, fn := filepath.Split(path)
	if _, ok := r.watching[dir]; ok {
		r.watching[dir].Remove(fn)
	}

	if r.watching[dir].Size() == 0 {
		if err := r.watcher.Remove(dir); err != nil {
			return fmt.Errorf("removing parent directory of %q from watcher: %w", path, err)
		}
	}

	return nil
}

func (r *Runner) Run(ctx context.Context, version int, path string, code []byte) error {
	p, err := r.progress.Start(ctx, "diagnostics on change", "Started", nil)
	if err != nil {
		log.Printf("[ERROR]: starting progress tracking for diagnostics on change: %v", err)
	} else {
		defer func() {
			if err := p.End(ctx, "Done"); err != nil {
				log.Printf("[ERROR]: stopping progress tracking for diagnostics: %v", err)
			}
		}()
	}

	return r.run(
		ctx,
		r.analyzers,
		version,
		path,
		func(actx context.Context, a Analyzer) ([]protocol.Diagnostic, error) {
			return a.Analyze(actx, path, code) // nolint:wrapcheck // Already wrapped by run.
		},
		r.AddDiagnostics,
	)
}

func (r *Runner) AddDiagnostics(path string, version int, diagnostics []protocol.Diagnostic) {
	r.diagnosticsMu.Lock()
	defer r.diagnosticsMu.Unlock()

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

	log.Printf("[WARN]: trying to add outdated change diagnostics to %q", path)
}

func (r *Runner) AddSaveDiagnostics(path string, version int, diagnostics []protocol.Diagnostic) {
	r.diagnosticsMu.Lock()
	defer r.diagnosticsMu.Unlock()

	fd := r.diagnostics[path]

	if version == fd.version {
		if len(diagnostics) == 0 {
			return
		}

		fd.saveDiagnostics = append(fd.saveDiagnostics, diagnostics...)
		fd.published = false
		return
	}

	if version > fd.version {
		fd.saveDiagnostics = diagnostics
		fd.published = false
		fd.version = version
		return
	}

	log.Printf("[WARN]: trying to add outdated save diagnostics to %q", path)
}

func (r *Runner) Publish(ctx context.Context, path string) error {
	r.diagnosticsMu.Lock()
	defer r.diagnosticsMu.Unlock()

	if fd, ok := r.diagnostics[path]; ok {
		if fd.published {
			return nil
		}
		fd.published = true

		diagnostics := make([]protocol.Diagnostic, 0, len(fd.diagnostics)+len(fd.saveDiagnostics))
		diagnostics = append(diagnostics, fd.diagnostics...)
		diagnostics = append(diagnostics, fd.saveDiagnostics...)

		log.Printf("[INFO]: publishing diagnostic update: %d diagnostics", len(diagnostics))

		if err := r.client.PublishDiagnostics(ctx, &protocol.PublishDiagnosticsParams{
			URI:         protocol.DocumentURI("file://" + path),
			Version:     int32(fd.version),
			Diagnostics: diagnostics,
		}); err != nil {
			return fmt.Errorf("publishing diagnostics: %w", err)
		}
	}

	return nil
}

func (r *Runner) Close() {
	r.diagnosticsMu.Lock()
	r.watchingMu.Lock()
	defer r.diagnosticsMu.Unlock()
	defer r.watchingMu.Unlock()

	if err := r.watcher.Close(); err != nil {
		log.Printf("[ERROR]: closing watcher: %v", err)
	}
	for _, a := range r.analyzers {
		if closer, ok := a.(Closer); ok {
			closer.Close()
		}
	}
}

func (r *Runner) setupPath(path string, version int) {
	r.diagnosticsMu.Lock()
	defer r.diagnosticsMu.Unlock()

	if _, ok := r.diagnostics[path]; ok {
		return
	}

	r.diagnostics[path] = &fileDiagnostics{
		path:      path,
		version:   version,
		published: true,
	}
}

func (r *Runner) cancelPrevRun(ctx context.Context, path string) context.Context {
	r.diagnosticsMu.Lock()
	defer r.diagnosticsMu.Unlock()

	fd := r.diagnostics[path]
	if fd.cancel != nil {
		log.Println("[INFO]: cancelling previous run")
		fd.cancel()
	}

	newCtx, cancelFunc := context.WithCancel(ctx)
	fd.cancel = cancelFunc
	return newCtx
}

func (r *Runner) run(
	ctx context.Context,
	analyzers []Analyzer,
	version int,
	path string,
	analyzeFunc func(context.Context, Analyzer) ([]protocol.Diagnostic, error),
	addFunc func(string, int, []protocol.Diagnostic),
) (reserr error) {
	logTime := timeDiagnostics("all")
	defer logTime()
	_, filename := filepath.Split(path)
	log.Printf("[INFO]: running analyzers for %q version %d", filename, version)

	r.setupPath(path, version)
	ctx = r.cancelPrevRun(ctx, path)

	errorC := make(chan error)
	resultsC := make(chan []protocol.Diagnostic)
	for _, a := range analyzers {
		go func(a Analyzer) {
			logTime := timeDiagnostics(a.Name())
			defer logTime()

			res, err := analyzeFunc(ctx, a)
			if err != nil {
				errorC <- fmt.Errorf("%s: %w", a.Name(), err)
				return
			}

			resultsC <- normalizeDiagnostics(a.Name(), res)
		}(a)
	}

	timer := time.NewTicker(time.Millisecond * 50)
	defer timer.Stop()
	waitingFor := len(analyzers)
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
			addFunc(path, version, res)

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

	return fmt.Errorf(
		"multiple errors running diagnostics for %q: %w",
		filename,
		errors.Join(errs...),
	)
}

func (r *Runner) runForSave(path string) {
	ctx := context.Background()
	p, err := r.progress.Start(ctx, "diagnostics on save", "Started", nil)
	if err != nil {
		log.Printf("[ERROR]: starting progress tracking for diagnostics on save: %v", err)
	} else {
		defer func() {
			if err := p.End(context.Background(), "Done"); err != nil {
				log.Printf("[ERROR]: stopping progress for diagnostics on save: %v", err)
			}
		}()
	}

	r.diagnosticsMu.Lock()
	var version int
	if fd, ok := r.diagnostics[path]; ok {
		version = fd.version + 1
	}
	r.diagnosticsMu.Unlock()

	err = r.run(
		context.Background(),
		r.saveAnalyzers,
		version,
		path,
		func(ctx context.Context, a Analyzer) ([]protocol.Diagnostic, error) {
			return a.AnalyzeSave( // nolint:wrapcheck // Already wrapped by run.
				ctx,
				path,
			)
		},
		r.AddSaveDiagnostics,
	)
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Printf("[ERROR]: running diagnostics on save error: %v", err)
		if err := r.client.LogMessage(ctx, &protocol.LogMessageParams{
			Type:    protocol.Error,
			Message: fmt.Sprintf("Running diagnostics on save error: %v", err),
		}); err != nil {
			log.Printf("[ERROR]: sending diagnostics on save error log to client: %v", err)
		}
	}
}

func (r *Runner) setupWatcher() error {
	if r.watcher != nil {
		return nil
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creating file watcher: %w", err)
	}

	r.watcher = watcher

	go func() {
		for {
			select {
			case event, ok := <-r.watcher.Events:
				if !ok {
					log.Println("[INFO]: file watcher channel closed")
					return
				}

				// Write for saves, Create for creation, a file that is watched might not have actually been created yet.
				if !event.Has(fsnotify.Write) && !event.Has(fsnotify.Create) {
					continue
				}

				// Check if we are actually watching this file.
				// r.watching[dir] should always be there at this point.
				dir, fn := filepath.Split(event.Name)
				r.watchingMu.Lock()
				if !r.watching[dir].Has(fn) {
					r.watchingMu.Unlock()
					continue
				}
				r.watchingMu.Unlock()

				log.Printf(
					"[INFO]: detected change in watched file %q, running analyzers",
					event.Name,
				)

				go r.runForSave(event.Name)
			case err, ok := <-r.watcher.Errors:
				if !ok {
					log.Println("[INFO]: file watcher error channel closed")
					return
				}

				log.Printf("[ERROR]: diagnostics file system watcher error: %v", err)
			}
		}
	}()

	return nil
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
		diagnostics[i].Source = "phpls-" + name
		diagnostics[i].Message = fmt.Sprintf("[%s] %s", name, diagnostics[i].Message)
	}

	return diagnostics
}

func findExec(tries []string) (string, bool) {
	for _, try := range tries {
		if path, err := exec.LookPath(try); err == nil {
			return path, true
		}
	}

	return "", false
}
