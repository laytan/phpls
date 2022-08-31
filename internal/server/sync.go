package server

import (
	"context"
	"log"
	"strings"

	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
	"github.com/laytan/elephp/pkg/lsperrors"
)

func (s *Server) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	return s.isMethodAllowed("DidOpen")
}

func (s *Server) DidChange(
	ctx context.Context,
	params *protocol.DidChangeTextDocumentParams,
) error {
	if err := s.isMethodAllowed("DidChange"); err != nil {
		return err
	}

	for _, changes := range params.ContentChanges {
		if changes.Range != nil {
			log.Println("LSP Server does not support ranges in DidChange requests")
			return lsperrors.ErrRequestFailed(
				"LSP Server does not support ranges in DidChange requests",
			)
		}
	}

	path := strings.TrimPrefix(string(params.TextDocument.URI), "file://")
	newContent := params.ContentChanges[len(params.ContentChanges)-1]

	if err := s.project.ParseFileUpdate(path, newContent.Text); err != nil {
		log.Println(err)
		return lsperrors.ErrRequestFailed(err.Error())
	}

	log.Printf("Parsed changes for file %s\n", path)
	return nil
}

func (s *Server) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	return s.isMethodAllowed("DidClose")
}
