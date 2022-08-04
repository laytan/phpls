package server

import (
	"context"
	"os"
	"strings"

	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
	"github.com/laytan/elephp/internal/project"
	"github.com/laytan/elephp/pkg/lsperrors"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/pkg/processwatch"
	log "github.com/sirupsen/logrus"
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

	phpv, err := phpversion.Get()
	if err != nil {
		log.Error(err)
		return nil, lsperrors.ErrRequestFailed("LSP Server " + err.Error())
	}

	log.Infof("Detected php version: %s\n", phpv.String())

	s.project = project.NewProject(string(s.root), phpv)

	// TODO: send this error to client, seems important to know about.
	go func() {
		if err := s.project.Parse(); err != nil {
			log.Error(err)
		}
	}()

	if params.ProcessID != 0 {
		processwatch.NewExiter(uint16(params.ProcessID))
		log.Infof("Monitoring process ID: %d\n", params.ProcessID)
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
			},
		},
		ServerInfo: serverInfo{
			Name: "elephp",
			// TODO: version from env/go/anywhere
			Version: "0.0.1-dev",
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
