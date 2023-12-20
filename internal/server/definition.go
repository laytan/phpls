package server

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
	"github.com/laytan/phpls/internal/project"
	"github.com/laytan/phpls/internal/project/definition"
	"github.com/laytan/phpls/internal/wrkspc"
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

	return functional.Map(definitions, func(def *definition.Definition) protocol.Location {
		content := wrkspc.Current.ContentF(def.Path)
		return position.FromIRPosition(def.Path, content, def.Position.StartPos).ToLSPLocation()
	}), nil
}
