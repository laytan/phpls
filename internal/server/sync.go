package server

import (
	"context"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/conf"
	"github.com/VKCOM/php-parser/pkg/errors"
	"github.com/VKCOM/php-parser/pkg/parser"
	"github.com/VKCOM/php-parser/pkg/version"
	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
	"github.com/laytan/elephp/pkg/lsperrors"
)

func (s *server) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	if err := s.isMethodAllowed("DidOpen"); err != nil {
		return err
	}

	if s.openFile != nil {
		return lsperrors.ErrRequestFailed("LSP Server is already tracking an open file")
	}

	s.openFile = &params.TextDocument
	return nil
}

func (s *server) DidChange(
	ctx context.Context,
	params *protocol.DidChangeTextDocumentParams,
) error {
	if err := s.isMethodAllowed("DidChange"); err != nil {
		return err
	}

	if s.openFile == nil {
		return lsperrors.ErrRequestFailed("LSP Server is not tracking an open file to be changed")
	}

	if params.TextDocument.URI != s.openFile.URI {
		return lsperrors.ErrRequestFailed("LSP Server is tracking a different file as open")
	}

	for _, changes := range params.ContentChanges {
		if changes.Range != nil {
			return lsperrors.ErrRequestFailed(
				"LSP Server does not support ranges in DidChange requests",
			)
		}

		s.openFile.Text = changes.Text
	}

	return nil
}

func (s *server) DidClose(context.Context, *protocol.DidCloseTextDocumentParams) error {
	if err := s.isMethodAllowed("DidClose"); err != nil {
		return err
	}

	s.openFile = nil
	return nil
}
