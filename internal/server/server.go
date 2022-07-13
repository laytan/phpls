package server

import (
	"github.com/jdbaldry/go-language-server-protocol/jsonrpc2"
	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
	"github.com/laytan/elephp/pkg/lsperrors"
)

func NewServer(client protocol.ClientCloser) *server {
	return &server{
		client: client,
	}
}

type serverInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

type server struct {
	client protocol.ClientCloser

	// Untill this is true, the server should only allow 'initialize' and 'initialized' requests.
	isInitialized bool

	// When this is true, the server should only allow 'exit' requests.
	// When the 'exit' request is sent, the server can exit completely.
	isShuttingDown bool

	openFile *protocol.TextDocumentItem
}

// OPTIM: Might make sense to use the state design pattern, eliminating the call
// OPTIM: to this method in every handler.
func (s *server) isMethodAllowed(method string) error {
	// If we are shutting down, we only allow an exit request.
	if s.isShuttingDown && method != "Exit" {
		return jsonrpc2.ErrInvalidRequest
	}

	// When not initialized, we require initialization first.
	if !s.isInitialized && method != "Initialize" && method != "Initialized" {
		return lsperrors.ErrServerNotInitialized
	}

	return nil
}
