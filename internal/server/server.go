package server

import (
	"context"
	"fmt"

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

	root string

	project *project.Project
}

// OPTIM: Might make sense to use the state design pattern, eliminating the call
// to this method in every handler.
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

func (s *Server) showAndLogErr(ctx context.Context, t protocol.MessageType, err error) {
	s.client.ShowMessage(ctx, &protocol.ShowMessageParams{
		Type:    t,
		Message: fmt.Sprintf("%v", err),
	})
	log.Error(err)
}

func (s *Server) showAndLogMsg(ctx context.Context, t protocol.MessageType, msg string) {
	s.client.ShowMessage(ctx, &protocol.ShowMessageParams{
		Type:    t,
		Message: msg,
	})

	switch t {
	case protocol.Error:
		log.Errorln(msg)
	case protocol.Warning:
		log.Warnln(msg)
	default:
		log.Infoln(msg)
	}
}
