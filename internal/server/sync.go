package server

import (
	"context"
	"strings"
	"time"

	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
	"github.com/laytan/elephp/pkg/lsperrors"
	log "github.com/sirupsen/logrus"
)

func (s *Server) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	if err := s.isMethodAllowed("DidOpen"); err != nil {
		return err
	}

	path := strings.TrimPrefix(string(params.TextDocument.URI), "file://")

	if s.openFile != "" && s.openFile != path {
		log.Errorln("Can't track file because an open file is already being tracked")
		return lsperrors.ErrRequestFailed("LSP Server is already tracking an open file")
	}

	s.openFile = path
	log.Infof("Started tracking open file %s\n", s.openFile)
	return nil
}

func (s *Server) DidChange(
	ctx context.Context,
	params *protocol.DidChangeTextDocumentParams,
) error {
	if err := s.isMethodAllowed("DidChange"); err != nil {
		return err
	}

	if s.openFile == "" {
		log.Errorln("Can't process DidChange request because there is no open file being tracked")
		return lsperrors.ErrRequestFailed("LSP Server is not tracking an open file to be changed")
	}

	path := strings.TrimPrefix(string(params.TextDocument.URI), "file://")
	if path != s.openFile {
		log.Errorf(
			"Got DidChange request for file %s but we are tracking %s as open\n",
			path,
			s.openFile,
		)
		return lsperrors.ErrRequestFailed("LSP Server is tracking a different file as open")
	}

	for _, changes := range params.ContentChanges {
		if changes.Range != nil {
			log.Errorln("LSP Server does not support ranges in DidChange requests")
			return lsperrors.ErrRequestFailed(
				"LSP Server does not support ranges in DidChange requests",
			)
		}

		if err := s.project.ParseFileContent(
			path,
			[]byte(changes.Text),
			time.Now(),
		); err != nil {
			log.Error(err)
			return lsperrors.ErrRequestFailed(err.Error())
		}
	}

	log.Infof("Parsed changes for open file %s\n", s.openFile)
	return nil
}

func (s *Server) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	if err := s.isMethodAllowed("DidClose"); err != nil {
		return err
	}

	// OPTIM: This might not be necessary, the changes made are good enough.
	if err := s.project.ParseFile(
		strings.TrimPrefix(string(params.TextDocument.URI), "file://"),
		time.Now(),
	); err != nil {
		log.Error(err)
		return lsperrors.ErrRequestFailed(err.Error())
	}

	log.Infof("Closed file %s, started tracking it from the filesystem again\n", s.openFile)

	s.openFile = ""
	return nil
}
