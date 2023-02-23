package project

import (
	"fmt"
	"strings"
	"sync"

	"github.com/laytan/elephp/internal/config"
	"github.com/laytan/elephp/internal/throws"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/phplint"
	"golang.org/x/exp/slices"
)

func (p *Project) Diagnose(path string, content string) (issues []Issue, changed bool, err error) {
	if strings.HasPrefix(path, config.FromContainer().StubsDir()) {
		return
	}

	var phpIssues []*phplint.Issue
	var throwIssues []*throws.Violation

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		is, e := phplint.LintString([]byte(content))
		if e != nil {
			err = e
		}

		phpIssues = is
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		throwIssues = throws.Diagnose(wrkspc.NewRooter(path))
	}()

	wg.Wait()

	issues = slices.Grow(issues, len(phpIssues)+len(throwIssues))
	for _, i := range phpIssues {
		issues = append(issues, i)
	}
	for _, i := range throwIssues {
		issues = append(issues, i)
	}

	if p.diagnosticsChanged(path, issues) {
		p.diagnostics[path] = issues
		return issues, true, nil
	}

	return issues, false, nil
}

func (p *Project) DiagnoseFile(path string) (issues []Issue, changed bool, err error) {
	phpIssues, err := phplint.LintFile(path)
	if err != nil {
		return nil, false, fmt.Errorf("project.DiagnoseFile(%s): %w", path, err)
	}

	issues = slices.Grow(issues, len(phpIssues))
	for _, i := range phpIssues {
		issues = append(issues, i)
	}

	// NOTE: this does not work because at this point we might not have the
	// project indexed yet, need some kind of subscriber pattern to execute
	// the diagnostics when the index is ready.
	// issues = addThrowsIssues(path, phpIssues)

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

func (p *Project) diagnosticsChanged(path string, issues []Issue) bool {
	currIssues, ok := p.diagnostics[path]
	if !ok {
		return true
	}

	if len(currIssues) != len(issues) {
		return true
	}

	equal := true
Issues:
	for _, issue := range issues {
		for _, currIssue := range currIssues {
			if issue.Line() != currIssue.Line() || issue.Message() != currIssue.Message() ||
				issue.Code() != currIssue.Code() {
				continue Issues
			}
		}

		equal = false
		break
	}

	return !equal
}
