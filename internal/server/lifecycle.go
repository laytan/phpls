package server

import (
	"context"
	"os"
	"strings"

	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
	"github.com/laytan/elephp/internal/project"
	"github.com/laytan/elephp/pkg/lsperrors"
)

// Entrypoint, must be requested first by the client.
func (s *server) Initialize(
	ctx context.Context,
	params *protocol.ParamInitialize,
) (*protocol.InitializeResult, error) {
	if err := s.isMethodAllowed("Initialize"); err != nil {
		return nil, err
	}

	if s.isInitialized {
		return nil, lsperrors.ErrRequestFailed("LSP Server is already initialized")
	}

	// OPTIM: might need to keep track of given 'processId' and exit when it dies.

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

	s.project = project.NewProject(string(s.root))
	go s.project.Parse()

	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync: &protocol.TextDocumentSyncOptions{
				// OPTIM: Full is easier for now, but Incremental would be better
				// OPTIM: for performance and a good improvement for later.
				Change: protocol.Full,

				OpenClose: true,
			},
			DefinitionProvider: true,
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
func (s *server) Initialized(context.Context, *protocol.InitializedParams) error {
	if err := s.isMethodAllowed("Initialized"); err != nil {
		return err
	}

	s.isInitialized = true
	return nil
}

// Starts the shutdown procedure, the client indicates it wants us to exit soon.
func (s *server) Shutdown(context.Context) error {
	if err := s.isMethodAllowed("Shutdown"); err != nil {
		return err
	}

	s.isShuttingDown = true
	return nil
}

// Exits with error code 0 when isShuttingDown, 1 otherwise.
func (s *server) Exit(context.Context) error {
	if err := s.isMethodAllowed("Exit"); err != nil {
		return err
	}

	if s.isShuttingDown {
		os.Exit(0)
		return nil
	}

	os.Exit(1)
	return nil
}
