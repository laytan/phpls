package server

import (
	"context"
	"log"
	"strings"
	"time"

	"appliedgo.net/what"
	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
	"github.com/laytan/elephp/internal/project"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/functional"
	"github.com/laytan/elephp/pkg/lsperrors"
	"github.com/laytan/elephp/pkg/strutil"
)

func (s *Server) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	if err := s.isMethodAllowed("DidOpen"); err != nil {
		return err
	}

	path := strings.TrimPrefix(string(params.TextDocument.URI), "file://")
	issues, isUpdated, err := s.project.DiagnoseFile(path)
	if err != nil {
		log.Println(err)
	} else if isUpdated {
		log.Printf("Opened file %s, found %d diagnostics", path, len(issues))

		err := s.client.PublishDiagnostics(
			ctx,
			s.convertIssuesToDiagnostics(issues, params.TextDocument.URI, params.TextDocument.Version),
		)
		if err != nil {
			log.Println(err)
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
	newContent := params.ContentChanges[len(params.ContentChanges)-1]

	prevContent := wrkspc.FromContainer().FContentOf(path)
	if strutil.RemoveWhitespace(newContent.Text) == strutil.RemoveWhitespace(prevContent) {
		what.Happens("Skipping file update (only whitespace change) of %s", path)
		return nil
	}

	start := time.Now()
	defer func() { log.Printf("File update took %s\n", time.Since(start)) }()

	if err := s.project.ParseFileUpdate(path, newContent.Text); err != nil {
		log.Println(err)
		return lsperrors.ErrRequestFailed(err.Error())
	}

	issues, isUpdated, err := s.project.Diagnose(path, newContent.Text)
	if err != nil {
		log.Println(err)
	} else if isUpdated {
		log.Printf("File change of %s resulted in %d updated diagnostics, pushing", path, len(issues))

		err := s.client.PublishDiagnostics(
			ctx,
			s.convertIssuesToDiagnostics(issues, params.TextDocument.URI, params.TextDocument.Version),
		)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}

func (s *Server) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	if err := s.isMethodAllowed("DidClose"); err != nil {
		return err
	}

	path := strings.TrimPrefix(string(params.TextDocument.URI), "file://")
	if s.project.HasDiagnostics(path) {
		log.Printf("File %s closed, clearing diagnostics", path)

		s.project.ClearDiagnostics(path)

		err := s.client.PublishDiagnostics(ctx, &protocol.PublishDiagnosticsParams{
			URI:         params.TextDocument.URI,
			Diagnostics: []protocol.Diagnostic{},
		})
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}

func (s *Server) convertIssuesToDiagnostics(
	issues []project.Issue,
	documentURI protocol.DocumentURI,
	documentVersion int32,
) *protocol.PublishDiagnosticsParams {
	return &protocol.PublishDiagnosticsParams{
		URI:     documentURI,
		Version: documentVersion,
		Diagnostics: functional.Map(issues, func(issue project.Issue) protocol.Diagnostic {
			return protocol.Diagnostic{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      uint32(issue.Line()) - 1,
						Character: 0,
					},
					End: protocol.Position{
						Line:      uint32(issue.Line()),
						Character: 0,
					},
				},
				Severity: protocol.SeverityError,
				Code:     issue.Code(),
				Source:   "PHP",
				Message:  issue.Message(),
			}
		}),
	}
}
