package diagnostics

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
	"github.com/laytan/phpls/pkg/functional"
	"github.com/laytan/phpls/pkg/phpstan"
)

type PhpstanAnalyzer struct {
	Executable string
}

var _ Analyzer = &PhpstanAnalyzer{}

func (p *PhpstanAnalyzer) Name() string {
	return "phpstan"
}

func (p *PhpstanAnalyzer) Analyze(
	ctx context.Context,
	path string,
	code []byte,
) ([]protocol.Diagnostic, error) {
	return transformPhpstanResult(phpstan.Analyze(ctx, p.Executable, path, code))
}

func (p *PhpstanAnalyzer) AnalyzeSave(
	ctx context.Context,
	path string,
) ([]protocol.Diagnostic, error) {
	return transformPhpstanResult(phpstan.AnalyzePath(ctx, p.Executable, path))
}

func phpstanMessageToDiagnostic(m *phpstan.ReportMessage) protocol.Diagnostic {
	return protocol.Diagnostic{
		Range: protocol.Range{
			Start: protocol.Position{
				Line:      uint32(m.Ln - 1),
				Character: 0,
			},
			End: protocol.Position{
				Line:      uint32(m.Ln),
				Character: 0,
			},
		},
		Severity: protocol.SeverityError,
		// Code:               nil,
		// CodeDescription:    &protocol.CodeDescription{},
		// Source:  "phpls:phpstan",
		Message: m.Msg,
		// Tags:               []protocol.DiagnosticTag{},
		// RelatedInformation: []protocol.DiagnosticRelatedInformation{},
		// Data:               nil,
	}
}

func transformPhpstanResult(
	msgs []*phpstan.ReportMessage,
	err error,
) ([]protocol.Diagnostic, error) {
	if errors.Is(err, phpstan.ErrCancelled) {
		log.Printf("[DEBUG]: phpstan cancelled: %v", err)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("analyzing with phpstan: %w", err)
	}

	return functional.Map(msgs, phpstanMessageToDiagnostic), nil
}
