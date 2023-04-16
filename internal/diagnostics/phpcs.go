package diagnostics

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
	"github.com/laytan/phpls/internal/wrkspc"
	"github.com/laytan/phpls/pkg/functional"
	"github.com/laytan/phpls/pkg/phpcs"
)

type PhpcsAnalyzer struct {
	instance *phpcs.Instance
}

var (
	_ Analyzer = &PhpcsAnalyzer{}
	_ Closer   = &PhpcsAnalyzer{}
)

func MakePhpcs(executable string) *PhpcsAnalyzer {
	return &PhpcsAnalyzer{
		instance: phpcs.New(executable),
	}
}

func (p *PhpcsAnalyzer) Analyze(
	ctx context.Context,
	path string,
	code []byte,
) ([]protocol.Diagnostic, error) {
	report, err := p.instance.Check(ctx, code)
	if errors.Is(err, phpcs.ErrCancelled) {
		log.Printf("[DEBUG]: phpcs cancelled: %v", err)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("analyzing %q with phpcs: %w", path, err)
	}

	// guaranteed to be there.
	return functional.Map(report.Files["STDIN"].Messages, phpcsMessageToDiagnostic), nil
}

func (p *PhpcsAnalyzer) AnalyzeSave(
	ctx context.Context,
	path string,
) ([]protocol.Diagnostic, error) {
	return p.Analyze(ctx, path, []byte(wrkspc.Current.ContentF(path)))
}

func (p *PhpcsAnalyzer) Name() string {
	return "phpcs"
}

func (p *PhpcsAnalyzer) Close() {
	p.instance.Close()
}

func phpcsMessageToDiagnostic(m *phpcs.ReportMessage) protocol.Diagnostic {
	pos := protocol.Position{Line: uint32(m.Row) - 1, Character: uint32(m.Column) - 1}
	return protocol.Diagnostic{
		Range:    protocol.Range{Start: pos, End: pos},
		Severity: phpcsToLSPSeverity(m.Type),
		Code:     m.Source,
		// Source:   "phpls",
		Message: m.Msg,

		// TODO: check if message is for deprecation/unnecessary, and add tags accordingly.
		// Tags:               []protocol.DiagnosticTag{},

		// CodeDescription:    &protocol.CodeDescription{},
		// RelatedInformation: []protocol.DiagnosticRelatedInformation{},
		//
		// TODO: if fixable, give user a code action to fix current file.
		// Data:               nil,
	}
}

func phpcsToLSPSeverity(s phpcs.Severity) protocol.DiagnosticSeverity {
	switch s {
	case phpcs.Warning:
		return protocol.SeverityWarning
	case phpcs.Error:
		return protocol.SeverityError
	default:
		panic("unreachable")
	}
}
