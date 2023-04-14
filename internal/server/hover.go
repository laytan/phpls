package server

import (
	"context"
	"log"
	"time"

	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
	"github.com/laytan/phpls/pkg/position"
)

func (s *Server) Hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	start := time.Now()
	defer func() {
		log.Printf("Retrieving hover took %s\n", time.Since(start))
	}()

	pos := position.FromTextDocumentPositionParams(&params.Position, &params.TextDocument)
	content := s.project.Hover(pos)

	if content == "" {
		return nil, nil
	}

	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  "markdown",
			Value: content,
		},
	}, nil
}
