package server

import (
	"context"
	"errors"
	"time"

	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
	"github.com/laytan/elephp/internal/project"
	"github.com/laytan/elephp/pkg/lsperrors"
	"github.com/laytan/elephp/pkg/position"
	log "github.com/sirupsen/logrus"
)

func (s *Server) Definition(
	ctx context.Context,
	params *protocol.DefinitionParams,
) (protocol.Definition, error) {
	start := time.Now()
	defer func() { log.Infof("Retrieving definition took %s\n", time.Since(start)) }()

	target := position.FromTextDocumentPositionParams(&params.Position, &params.TextDocument)
	pos, err := s.project.Definition(target)
	if err != nil {
		if errors.Is(err, project.ErrNoDefinitionFound) {
			log.Warn(err)
			return nil, nil
		}

		log.Error(err)
		return nil, lsperrors.ErrRequestFailed(err.Error())
	}

	return pos.ToLSPLocation(), nil
}
