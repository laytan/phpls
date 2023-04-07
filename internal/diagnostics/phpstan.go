package diagnostics

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/laytan/elephp/pkg/functional"
	"github.com/laytan/elephp/pkg/phpstan"
	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
)

type PhpstanAnalyzer struct{}

func (p *PhpstanAnalyzer) Name() string {
	return "phpstan"
}

func (p *PhpstanAnalyzer) Analyze(
	ctx context.Context,
	path string,
	code []byte,
) ([]protocol.Diagnostic, error) {
	report, err := phpstan.Analyze(ctx, path, code)
	if errors.Is(err, phpstan.ErrCancelled) {
		log.Printf("[DEBUG]: phpstan cancelled: %v", err)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("analyzing %q with phpstan: %w", path, err)
	}

	return functional.Map(report, phpstanMessageToDiagnostic), nil
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
		// Source:  "elephp:phpstan",
		Message: m.Msg,
		// Tags:               []protocol.DiagnosticTag{},
		// RelatedInformation: []protocol.DiagnosticRelatedInformation{},
		// Data:               nil,
	}
}
