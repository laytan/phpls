package project

import (
	"fmt"

	"github.com/laytan/elephp/pkg/phplint"
)

func (p *Project) Diagnose(path string, content string) ([]*phplint.Issue, bool, error) {
	issues, err := phplint.LintString([]byte(content))
	if err != nil {
		return nil, false, fmt.Errorf("project.Diagnose(%s): %w", path, err)
	}

	if p.diagnosticsChanged(path, issues) {
		p.diagnostics[path] = issues
		return issues, true, nil
	}

	return issues, false, nil
}

func (p *Project) DiagnoseFile(path string) ([]*phplint.Issue, bool, error) {
	issues, err := phplint.LintFile(path)
	if err != nil {
		return nil, false, fmt.Errorf("project.DiagnoseFile(%s): %w", path, err)
	}

	if p.diagnosticsChanged(path, issues) {
		p.diagnostics[path] = issues
		return issues, true, nil
	}

	return issues, false, nil
}

func (p *Project) HasDiagnostics(path string) bool {
	issues, ok := p.diagnostics[path]
	return ok && len(issues) > 0
}

func (p *Project) ClearDiagnostics(path string) {
	delete(p.diagnostics, path)
}

func (p *Project) diagnosticsChanged(path string, issues []*phplint.Issue) bool {
	currIssues, ok := p.diagnostics[path]
	if !ok {
		return true
	}

	if len(currIssues) != len(issues) {
		return true
	}

	equal := true
	for i, issue := range issues {
		currIssue := currIssues[i]

		if !issue.Equals(currIssue) {
			equal = false
			break
		}
	}

	return !equal
}
