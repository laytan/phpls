package server

import (
	"github.com/jdbaldry/go-language-server-protocol/jsonrpc2"
	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
	"github.com/laytan/elephp/internal/project"
	"github.com/laytan/elephp/pkg/lsperrors"
	log "github.com/sirupsen/logrus"
)

func NewServer(client protocol.ClientCloser) *Server {
	return &Server{
		client: client,
	}
}

type Server struct {
	client protocol.ClientCloser

	// Untill this is true, the server should only allow 'initialize' and 'initialized' requests.
	isInitialized bool

	// When this is true, the server should only allow 'exit' requests.
	// When the 'exit' request is sent, the server can exit completely.
	isShuttingDown bool

	// The currently open file, if any.
	openFile string

	root string

	project *project.Project
}

// OPTIM: Might make sense to use the state design pattern, eliminating the call
// OPTIM: to this method in every handler.
func (s *Server) isMethodAllowed(method string) error {
	// If we are shutting down, we only allow an exit request.
	if s.isShuttingDown && method != "Exit" {
		log.Errorf(
			"Method %s not allowed because the server is waiting for the exit method\n",
			method,
		)
		return jsonrpc2.ErrInvalidRequest
	}

	// When not initialized, we require initialization first.
	if !s.isInitialized && method != "Initialize" && method != "Initialized" {
		log.Errorf(
			"Method %s not allowed because the server is waiting for the initialization methods\n",
			method,
		)
		return lsperrors.ErrServerNotInitialized
	}

	return nil
}
