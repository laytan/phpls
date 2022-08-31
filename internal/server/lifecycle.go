package server

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
	"github.com/laytan/elephp/internal/project"
	"github.com/laytan/elephp/pkg/lsperrors"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/pkg/processwatch"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

const (
	name    = "elephp"
	version = "0.0.1-dev"

	indexingProgressToken = "indexing"
	// Time between progress updates.
	indexingProgressInterval = 100 * time.Millisecond
	// The duration that the last progress message is shown, before end is sent.
	indexingDecayTime = 2 * time.Second
)

type serverInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

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
	if len(s.root) == 0 {
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

	phpv, err := phpversion.Get()
	if err != nil {
		log.Error(err)
		return nil, lsperrors.ErrRequestFailed("LSP Server " + err.Error())
	}

	log.Infof("Detected php version: %s\n", phpv.String())

	s.project = project.NewProject(string(s.root), phpv)

	go func() {
		filesDone := atomic.Uint32{}
		start := time.Now()

		message := func() string {
			doneAmt := filesDone.Load()

			return fmt.Sprintf(
				"Indexed %d source files in %s",
				doneAmt,
				time.Since(start).Round(time.Millisecond),
			)
		}
		log.Infoln(message())

		ticker := time.NewTicker(indexingProgressInterval)
		done := make(chan bool)
		go func() {
			s.client.Progress(ctx, &protocol.ProgressParams{
				Token: indexingProgressToken,
				Value: progress{
					Kind:    progressKindBegin,
					Title:   "Indexing project",
					Message: message(),
				},
			})

			for {
				select {
				case <-done:
					s.client.Progress(ctx, &protocol.ProgressParams{
						Token: indexingProgressToken,
						Value: progress{
							Kind:    progressKindReport,
							Message: message(),
						},
					})

					time.Sleep(indexingDecayTime)

					s.client.Progress(ctx, &protocol.ProgressParams{
						Token: indexingProgressToken,
						Value: progress{
							Kind: progressKindEnd,
						},
					})
					return

				case <-ticker.C:
					s.client.Progress(ctx, &protocol.ProgressParams{
						Token: indexingProgressToken,
						Value: progress{
							Kind:    progressKindReport,
							Message: message(),
						},
					})
				}
			}
		}()

		err := s.project.Parse(&filesDone)
		if err != nil {
			s.showAndLogErr(ctx, protocol.Warning, err)
			return
		}

		ticker.Stop()
		done <- true
		close(done)

		log.Infof(message())
	}()

	if params.ProcessID != 0 {
		processwatch.NewExiter(uint(params.ProcessID))
		log.Infof("Monitoring process ID: %d\n", uint(params.ProcessID))
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
		ServerInfo: serverInfo{
			Name:    name,
			Version: version,
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
