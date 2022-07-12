package server

import (
	"context"

	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
)

func NewServer(client protocol.ClientCloser) *server {
	return &server{
		client: client,
	}
}

type server struct {
	client protocol.ClientCloser
}

func (s *server) Initialize(
	context.Context,
	*protocol.ParamInitialize,
) (*protocol.InitializeResult, error) {
	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{},
		ServerInfo: struct {
			Name    string `json:"name"`
			Version string `json:"version,omitempty"`
		}{
			Name: "elephp",
			// TODO: version from env/go/anywhere
			Version: "0.0.1-dev",
		},
	}, nil
}
