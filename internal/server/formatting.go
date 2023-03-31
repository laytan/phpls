package server

import (
	"context"

	"github.com/laytan/elephp/internal/phpcs"
	"github.com/laytan/elephp/pkg/lsperrors"
	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
)

func (s *Server) Formatting(
	ctx context.Context,
	params *protocol.DocumentFormattingParams,
) ([]protocol.TextEdit, error) {
	edits, err := phpcs.FormatFileEdits(position.URIToFile(string(params.TextDocument.URI)))
	if err != nil {
		err := lsperrors.ErrRequestFailed(err.Error())
		go s.showAndLog(ctx, protocol.Error, err)
		return nil, err
	}

	return edits, nil
}
