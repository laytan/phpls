package server

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
	"github.com/laytan/elephp/internal/project"
	"github.com/laytan/elephp/pkg/functional"
	"github.com/laytan/elephp/pkg/lsperrors"
	"github.com/laytan/elephp/pkg/position"
)

func (s *Server) Definition(
	ctx context.Context,
	params *protocol.DefinitionParams,
) (protocol.Definition, error) {
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
