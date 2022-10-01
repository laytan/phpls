package phplint

// package phplint wraps `php -l` into a nice interface.

import (
	"bufio"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var outputWithCodeRgx = regexp.MustCompile(
	`^(.*)\(([A-Z_]+)\)(.*) in .* on line (\d+)$`,
)
var outputWithoutCodeRgx = regexp.MustCompile(`^(.*) in .* on line (\d+)$`)

type Issue struct {
	Message string
	Code    string
	Line    int
}

func (i *Issue) Equals(o *Issue) bool {
	return i.Message == o.Message && i.Code == o.Code && i.Line == o.Line
}

// LintString lints the code given in the parameter, returning any issues.
// NOTE: This silently returns no issues when `php` is not available.
func LintString(code []byte) ([]*Issue, error) {
	cmd := exec.Command("php", "-l")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("phplint.LintString: cmd.StdinPipe: %w", err)
	}

	var stdinErr error
	go func() {
		defer stdin.Close()

		if _, err := stdin.Write(code); err != nil {
			stdinErr = err
		}
	}()

	out, err := cmd.Output()
	// We don't care about exit errors.
	var exitErr *exec.ExitError
	if err != nil && !errors.As(err, &exitErr) {
		return nil, fmt.Errorf("phplint.LintString: cmd.Output: %w", err)
	}

	// See if stdin.Write has an error.
	// This is thread safe, because cmd.Output() blocks.
	if stdinErr != nil {
		return nil, fmt.Errorf("phplint.LintString: stdin.Write: %w", stdinErr)
	}

	issues, err := parseIssues(out)
	if err != nil {
		return nil, fmt.Errorf("phplint.LintString: %w", err)
	}

	return issues, nil
}

// LintFile lints the code in the file at the given path, returning any issues.
// NOTE: This silently returns no issues when `php` is not available.
// NOTE: This silently returns no issues when no file could be found at filepath.
func LintFile(filepath string) ([]*Issue, error) {
	cmd := exec.Command("php", "-l", filepath)

	out, err := cmd.Output()
	// We don't care about exit errors.
	var exitErr *exec.ExitError
	if err != nil && !errors.As(err, &exitErr) {
		return nil, fmt.Errorf("phplint.LintFile(%s): cmd.Output: %w", filepath, err)
	}

	issues, err := parseIssues(out)
	if err != nil {
		return nil, fmt.Errorf("phplint.LintFile(%s): %w", filepath, err)
	}

	return issues, nil
}

func parseIssues(out []byte) ([]*Issue, error) {
	issues := []*Issue{}

	reader := strings.NewReader(string(out))
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if match := outputWithCodeRgx.FindStringSubmatch(scanner.Text()); match != nil {
			line, err := strconv.Atoi(match[4])
			if err != nil {
				return nil, fmt.Errorf(
					"phplint.parseIssues: parsing output line: strconv.Atoi(%s): %w",
					match[4],
					err,
				)
			}

			issues = append(issues, &Issue{
				Message: match[1] + " " + match[3],
				Code:    match[2],
				Line:    line,
			})
			continue
		}

		if match := outputWithoutCodeRgx.FindStringSubmatch(scanner.Text()); match != nil {
			line, err := strconv.Atoi(match[2])
			if err != nil {
				return nil, fmt.Errorf(
					"phplint.parseIssues: parsing output line: strconv.Atoi(%s): %w",
					match[2],
					err,
				)
			}

			issues = append(issues, &Issue{
				Message: match[1],
				Code:    "",
				Line:    line,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("phplint.parseIssues: scanner.Err: %w", err)
	}

	return issues, nil
}
