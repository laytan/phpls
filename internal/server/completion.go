package server

import (
	"context"
	"fmt"
	"time"

	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
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
	results, incomplete := s.project.Complete(pos)

	items := make([]protocol.CompletionItem, len(results))
	for i, res := range results {
		log.Infoln(res.Key)
		if res.Value.Namespace == "" {
			items[i] = protocol.CompletionItem{
				Label: res.Key,
			}
			continue
		}

		// TODO: Is this namespace already imported?
		// TODO: check where the last namespace declaration is and put it after it.

		items[i] = protocol.CompletionItem{
			Label:  res.Key,
			Detail: fmt.Sprintf(`%s\%s`, res.Value.Namespace, res.Key),
			AdditionalTextEdits: []protocol.TextEdit{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1},
						End:   protocol.Position{Line: 1},
					},
					NewText: fmt.Sprintf(`use %s\%s;`, res.Value.Namespace, res.Key),
				},
			},
		}
	}

	return &protocol.CompletionList{
		Items:        items,
		IsIncomplete: incomplete,
	}, nil
}
