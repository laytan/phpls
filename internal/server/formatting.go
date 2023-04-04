package server

import (
	"context"
	"fmt"
	"time"

	"github.com/laytan/elephp/pkg/lsperrors"
	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
)

func (s *Server) Formatting(
	ctx context.Context,
	params *protocol.DocumentFormattingParams,
) ([]protocol.TextEdit, error) {
	if !s.phpcbf.HasExecutable() {
		return nil, nil
	}

	start := time.Now()
	defer func() {
		go s.showAndLog(ctx, protocol.Info, fmt.Errorf("formatting took %s", time.Since(start)))
	}()

	edits, err := s.phpcbf.FormatFileEdits(position.URIToFile(string(params.TextDocument.URI)))
	if err != nil {
		err := lsperrors.ErrRequestFailed(err.Error())
		go s.showAndLog(ctx, protocol.Error, err)
		return nil, err
	}

	return edits, nil
}
