package server

import (
	"context"
	"fmt"
	"log"

	"github.com/laytan/go-lsp-protocol/pkg/jsonrpc2"
	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
	"github.com/laytan/phpls/internal/diagnostics"
	"github.com/laytan/phpls/internal/project"
	"github.com/laytan/phpls/pkg/lsperrors"
	"github.com/laytan/phpls/pkg/lsprogress"
	"github.com/laytan/phpls/pkg/phpcs/phpcbf"
)

func NewServer(client protocol.ClientCloser) *Server {
	return &Server{
		client:   client,
		progress: lsprogress.NewTracker(client),
		phpcbf:   phpcbf.NewInstance(),
		diag:     diagnostics.NewRunnerFromConfig(client),
	}
}

type Server struct {
	client protocol.ClientCloser
	// Until this is true, the server should only allow 'initialize' and 'initialized' requests.
	isInitialized bool
	// When this is true, the server should only allow 'exit' requests.
	// When the 'exit' request is sent, the server can exit completely.
	isShuttingDown bool
	root           string
	project        *project.Project
	progress       *lsprogress.Tracker
	// NOTE: This can be nil if diagnostics are configured to be disabled!
	diag   *diagnostics.Runner
	phpcbf *phpcbf.Instance
}

var _ protocol.Server = &Server{}

// OPTIM: Might make sense to use the state design pattern, eliminating the call
// to this method in every handler.
func (s *Server) isMethodAllowed(method string) error {
	// If we are shutting down, we only allow an exit request.
	if s.isShuttingDown && method != "Exit" {
		log.Printf(
			"Method %s not allowed because the server is waiting for the exit method\n",
			method,
		)
		return jsonrpc2.ErrInvalidRequest
	}

	// When not initialized, we require initialization first.
	if !s.isInitialized && method != "Initialize" && method != "Initialized" {
		log.Printf(
			"Method %s not allowed because the server is waiting for the initialization methods\n",
			method,
		)
		return lsperrors.ErrServerNotInitialized
	}

	return nil
}

func (s *Server) showAndLog(ctx context.Context, t protocol.MessageType, msg any) {
	if err := s.client.ShowMessage(ctx, &protocol.ShowMessageParams{
		Type:    t,
		Message: fmt.Sprintf("%v", msg),
	}); err != nil {
		log.Println(err)
	}

	log.Println(msg)
}
