package server

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
	"github.com/laytan/phpls/internal/project"
	"github.com/laytan/phpls/pkg/functional"
	"github.com/laytan/phpls/pkg/lsperrors"
	"github.com/laytan/phpls/pkg/position"
)

func (s *Server) Definition(
	ctx context.Context,
	params *protocol.DefinitionParams,
) ([]protocol.Location, error) {
	if err := s.isMethodAllowed("Definition"); err != nil {
		return nil, err
	}

	start := time.Now()
	defer func() { log.Printf("Retrieving definition took %s\n", time.Since(start)) }()

	target := position.FromTextDocumentPositionParams(&params.Position, &params.TextDocument)
	definitions, err := s.project.Definition(target)
	if err != nil {
		if errors.Is(err, project.ErrNoDefinitionFound) {
			log.Println(err)
			return nil, nil
		}

		log.Println(err)
		return nil, lsperrors.ErrRequestFailed(err.Error())
	}

	return functional.Map(definitions, func(def *position.Position) protocol.Location {
		return def.ToLSPLocation()
	}), nil
}
