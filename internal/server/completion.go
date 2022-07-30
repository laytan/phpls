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

func (s *Server) Completion(
	ctx context.Context,
	params *protocol.CompletionParams,
) (*protocol.CompletionList, error) {
	start := time.Now()
	defer func() { log.Infof("Retrieving completion took %s\n", time.Since(start)) }()

	pos := position.FromTextDocumentPositionParams(&params.Position, &params.TextDocument)
	results, err := s.project.Complete(pos)
	if err != nil {
		if errors.Is(err, project.ErrNoCompletionResults) {
			log.Warn(err)
			return nil, nil
		}

		log.Error(err)
		return nil, lsperrors.ErrRequestFailed(err.Error())
	}

	completionItems := make([]protocol.CompletionItem, len(results))
	for i, result := range results {
		completionItems[i] = protocol.CompletionItem{
			Label: result,
		}
	}

	return &protocol.CompletionList{
		IsIncomplete: true,
		Items:        completionItems,
	}, nil
}
