package server

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/laytan/elephp/internal/config"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/lsperrors"
	"github.com/laytan/elephp/pkg/strutil"
	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
)

func (s *Server) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	if err := s.isMethodAllowed("DidOpen"); err != nil {
		return err
	}

	path := strings.TrimPrefix(string(params.TextDocument.URI), "file://")
	if s.diag != nil && !inStubs(path) {
		code := wrkspc.Current.FContentOf(path)
		if err := s.diag.Run(ctx, int(params.TextDocument.Version), path, []byte(code)); err != nil {
			if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
				s.showAndLog(ctx, protocol.Error, err)
			}
		}

		if err := s.diag.Watch(path); err != nil {
			s.showAndLog(ctx, protocol.Error, err)
		}
	}

	return nil
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
	newContent := params.ContentChanges[len(params.ContentChanges)-1].Text
	prevContent := wrkspc.Current.FContentOf(path)
	if strutil.RemoveWhitespace(prevContent) == strutil.RemoveWhitespace(newContent) {
		return nil
	}

	// TODO: check if this file extension is in config file extensions, and return if not.

	go func() {
		if err := s.project.ParseFileUpdate(path, newContent); err != nil {
			s.showAndLog(ctx, protocol.Warning, err)
		}
	}()

	if s.diag != nil && !inStubs(path) {
		go func() {
			if err := s.diag.Run(ctx, int(params.TextDocument.Version), path, []byte(newContent)); err != nil {
				if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
					s.showAndLog(ctx, protocol.Error, err)
				}
			}
		}()
	}

	return nil
}

func (s *Server) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	if err := s.isMethodAllowed("DidClose"); err != nil {
		return err
	}

	if s.diag != nil {
		path := strings.TrimPrefix(string(params.TextDocument.URI), "file://")
		if err := s.diag.StopWatching(path); err != nil {
			s.showAndLog(ctx, protocol.Error, err)
		}
	}

	return nil
}

func inStubs(path string) bool {
	return strings.HasPrefix(path, config.Current.StubsPath)
}
