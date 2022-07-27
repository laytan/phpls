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

	s.openFiles = append(s.openFiles, path)

	log.Infof("Started tracking open file %s\n", s.openFiles)
	return nil
}

func (s *Server) DidChange(
	ctx context.Context,
	params *protocol.DidChangeTextDocumentParams,
) error {
	if err := s.isMethodAllowed("DidChange"); err != nil {
		return err
	}

	path := strings.TrimPrefix(string(params.TextDocument.URI), "file://")

	isTrackingPath := false
	for _, checkingPath := range s.openFiles {
		if path == checkingPath {
			isTrackingPath = true
			break
		}
	}

	if !isTrackingPath {
		log.Errorf("Got DidChange request for file %s but we are tracking: %v\n", path, s.openFiles)
		return lsperrors.ErrRequestFailed("LSP Server is not tracking changed file")
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

	log.Infof("Parsed changes for open file %s\n", s.openFiles)
	return nil
}

func (s *Server) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	if err := s.isMethodAllowed("DidClose"); err != nil {
		return err
	}

	path := strings.TrimPrefix(string(params.TextDocument.URI), "file://")

	index := -1
	for i, checkingPath := range s.openFiles {
		if path == checkingPath {
			index = i
			break
		}
	}

	if index == -1 {
		log.Errorf("Trying to close file %s which is not in open files %v\n", path, s.openFiles)
		return lsperrors.ErrRequestFailed("Trying to close file which is not in tracked/open files")
	}

	s.openFiles = append(s.openFiles[:index], s.openFiles[index+1:]...)

	log.Infof("Closed file %s, started tracking it from the filesystem again\n", s.openFiles)
	return nil
}
