package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/project"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/lsperrors"
	"github.com/laytan/elephp/pkg/processwatch"
	"github.com/laytan/elephp/pkg/stubs"
	"github.com/samber/do"
	"golang.org/x/exp/slices"
)

// Entrypoint, must be requested first by the client.
func (s *Server) Initialize(
	ctx context.Context,
	params *protocol.ParamInitialize,
) (*protocol.InitializeResult, error) {
	if err := s.isMethodAllowed("Initialize"); err != nil {
		return nil, err
	}

	if s.isInitialized {
		return nil, lsperrors.ErrRequestFailed("LSP Server is already initialized")
	}

	// NOTE: Are we sending strings? There is a 'locale' param that we might want to support (translations).

	// TODO: store 'capabilities' of client and use when necessary, (maybe wrap in nice access methods).

	// OPTIM: support 'trace', when set, we need to send traces back specified by
	// OPTIM: the given trace severity.
	// NOTE: $/logTrace should be used for systematic trace reporting. For single debugging messages, the server should send window/logMessage notifications.

	// OPTIM: support 'workspaceFolders'.
	s.root = strings.TrimPrefix(string(params.RootURI), "file://")
	if s.root == "" {
		return nil, lsperrors.ErrRequestFailed("LSP Server requires RootURI to be set")
	}

	// TODO: do all the up-to-date clients support this or do we need to support
	// TODO: the other way too?
	if !slices.Contains(
		params.Capabilities.TextDocument.Completion.CompletionItem.ResolveSupport.Properties,
		"additionalTextEdits",
	) {
		return nil, lsperrors.ErrRequestFailed(
			`LSP Server requires the client to support the "additionalTextEdits" completion item resolve capability (textDocument.completion.completionItem.resolveSupport.properties).`,
		)
	}

	if !slices.Contains(params.Capabilities.TextDocument.Hover.ContentFormat, "markdown") {
		return nil, lsperrors.ErrRequestFailed(
			`LSP Server requires the client to support "markdown" hover results (textDocument.hover.contentFormat)`,
		)
	}

	stubsDir, err := s.initStubs(ctx)
	if err != nil {
		return nil, lsperrors.ErrRequestFailed(
			fmt.Errorf("Failed initializing stubs: %w", err).Error(),
		)
	}

	proj, err := s.createProject(stubsDir)
	if err != nil {
		return nil, lsperrors.ErrRequestFailed(
			fmt.Errorf("Failed initializing project: %w", err).Error(),
		)
	}

	s.project = proj

	go s.index()

	if params.ProcessID != 0 {
		processwatch.NewExiter(uint(params.ProcessID))
		log.Printf("Monitoring process ID: %d\n", uint(params.ProcessID))
	}

	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync: &protocol.TextDocumentSyncOptions{
				// OPTIM: Full is easier for now, but Incremental would be better
				// For performance and a good improvement for later.
				Change:    protocol.Full,
				OpenClose: true,
			},
			DefinitionProvider: true,
			CompletionProvider: protocol.CompletionOptions{
				TriggerCharacters: []string{"$", "-", ">"},
				ResolveProvider:   true,
			},
			HoverProvider: true,
		},
		ServerInfo: struct {
			Name    string `json:"name"`
			Version string `json:"version,omitempty"`
		}{
			Name:    Config().Name(),
			Version: Config().Version(),
		},
	}, nil
}

// The client has received our Initialize response and is going to start sending
// normal requests.
func (s *Server) Initialized(context.Context, *protocol.InitializedParams) error {
	if err := s.isMethodAllowed("Initialized"); err != nil {
		return err
	}

	s.isInitialized = true
	return nil
}

// Starts the shutdown procedure, the client indicates it wants us to exit soon.
func (s *Server) Shutdown(context.Context) error {
	if err := s.isMethodAllowed("Shutdown"); err != nil {
		return err
	}

	s.isShuttingDown = true

	log.Println("Received shutdown request, waiting for exit request")
	return nil
}

// Exits with error code 0 when isShuttingDown, 1 otherwise.
func (s *Server) Exit(context.Context) error {
	if err := s.isMethodAllowed("Exit"); err != nil {
		return err
	}

	log.Println("Received exit request, exiting")

	if s.isShuttingDown {
		os.Exit(0)
		return nil
	}

	os.Exit(1)
	return nil
}

func (s *Server) index() {
	ctx := context.Background()

	done := &atomic.Uint64{}
	total := &atomic.Uint64{}
	var finalTotal float64

	totalDoneChan := make(chan bool)
	go func() {
		<-totalDoneChan
		finalTotal = float64(total.Load())
	}()

	getTotal := func() float64 {
		if finalTotal != 0 {
			return finalTotal
		}

		return float64(total.Load())
	}

	stop, err := s.progress.Track(
		ctx,
		func() float64 { return float64(done.Load()) },
		getTotal,
		"indexing project",
		time.Millisecond*100,
	)
	if err != nil {
		s.showAndLogErr(ctx, protocol.Error, err)
		return
	}

	err = s.project.Parse(done, total, totalDoneChan)
	if err != nil {
		if err := stop(err); err != nil {
			s.showAndLogErr(ctx, protocol.Error, fmt.Errorf("stopping progress: %w", err))
			return
		}

		s.showAndLogErr(ctx, protocol.Info, err)
	}

	if err := stop(nil); err != nil {
		s.showAndLogErr(ctx, protocol.Error, err)
	}
}

func (s *Server) createProject(stubsDir string) (*project.Project, error) {
	phpv, err := Config().PHPVersion()
	if err != nil {
		return nil, fmt.Errorf("creating project: %w", err)
	}

	log.Printf("Detected php version: %s\n", phpv.String())

	i := index.New(phpv)
	w := wrkspc.New(phpv, string(s.root), stubsDir)
	do.ProvideValue(nil, i)
	do.ProvideValue(nil, w)

	return project.New(), nil
}

func (s *Server) initStubs(ctx context.Context) (string, error) {
	phpv, err := Config().PHPVersion()
	if err != nil {
		return "", fmt.Errorf("initializing stubs, getting php version: %w", err)
	}

	stubsDir, err := stubs.Path(Config().StubsDir(), phpv)
	if err == nil {
		return stubsDir, nil
	}

	if !errors.Is(err, stubs.ErrNotExists) {
		return "", fmt.Errorf("initializing stubs, getting stubs path: %w", err)
	}

	done := &atomic.Uint32{}
	stop, err := s.progress.Track(
		ctx,
		func() float64 { return float64(done.Load()) },
		func() float64 { return stubs.TotalStubs },
		fmt.Sprintf("generating stubs for PHP %s", phpv),
		time.Millisecond*50,
	)
	if err != nil {
		return "", fmt.Errorf("started tracking stub progress: %w", err)
	}

	stubsDir, err = stubs.Generate(Config().StubsDir(), phpv, done)

	if err != nil {
		err = fmt.Errorf("generating stubs: %w", err)
		if stopErr := stop(err); stopErr != nil {
			return "", fmt.Errorf("multiple errors stopping progress: %w", stopErr)
		}

		return "", err
	}

	if err = stop(nil); err != nil {
		return "", fmt.Errorf("stopping progress: %w", err)
	}

	return stubsDir, nil
}
