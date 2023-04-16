package server

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
	"github.com/laytan/phpls/pkg/functional"
	"github.com/laytan/phpls/pkg/position"
)

func (s *Server) Rename(
	ctx context.Context,
	params *protocol.RenameParams,
) (*protocol.WorkspaceEdit, error) {
	if err := s.isMethodAllowed("Rename"); err != nil {
		return nil, err
	}

	start := time.Now()
	defer func() {
		log.Printf("[INFO]: Rename took %s", time.Since(start))
	}()

	// TODO: check if new name is a valid name for this symbol.
	target := position.FromTextDocumentPositionParams(&params.Position, &params.TextDocument)
	references, err := s.references(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("finding references to rename: %w", err)
	}

	edit := protocol.WorkspaceEdit{}

	fileRefs := map[protocol.DocumentURI][]protocol.Location{}
	for _, reference := range references {
		if _, ok := fileRefs[reference.URI]; ok {
			fileRefs[reference.URI] = append(fileRefs[reference.URI], reference)
			continue
		}

		fileRefs[reference.URI] = []protocol.Location{reference}
	}

	for uri, references := range fileRefs {
		edits := functional.Map(
			references,
			func(reference protocol.Location) protocol.TextEdit {
				return protocol.TextEdit{
					Range:   reference.Range,
					NewText: params.NewName,
				}
			},
		)

		dc := protocol.DocumentChanges{
			TextDocumentEdit: &protocol.TextDocumentEdit{
				TextDocument: protocol.OptionalVersionedTextDocumentIdentifier{
					TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: uri},
				},
				Edits: edits,
			},
		}

		edit.DocumentChanges = append(edit.DocumentChanges, dc)
	}

	return &edit, nil
}
