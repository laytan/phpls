package phpcs

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"

	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/functional"
	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
	"github.com/sourcegraph/go-diff/diff"
)

//go:generate stringer -type fixerExitCode
type fixerExitCode int

const (
	Ok             fixerExitCode = 0
	GeneralErr     fixerExitCode = 1
	InvalidSyntax  fixerExitCode = 4
	NeedsFixing    fixerExitCode = 8
	AppConfigErr   fixerExitCode = 16
	FixerConfigErr fixerExitCode = 32
	PHPException   fixerExitCode = 64
)

type fixResult struct {
	Files []struct {
		Name string
		Diff string
	}
	Time struct {
		Total float64
	}
	Memory float64
}

func FormatFileDiff(path string) (*diff.FileDiff, error) {
	return FormatCodeDiff([]byte(wrkspc.Current.FContentOf(path)))
}

func FormatCodeDiff(code []byte) (*diff.FileDiff, error) {
	cmd := exec.Command("php-cs-fixer", "fix", "--diff", "--format=json", "-")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("running php-cs-fixer: %w", err)
	}

	var stdinErr error
	go func() {
		defer stdin.Close()
		_, stdinErr = stdin.Write(code) // This blocks until run/output, so needs goroutine.
	}()

	out, err := cmd.Output()
	cmd.ProcessState.ExitCode()
	if err == nil { // If something can be fixed, an error is returned.
		return nil, nil
	}

	if stdinErr != nil {
		return nil, fmt.Errorf("writing to stdin: %w", stdinErr)
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return nil, fmt.Errorf("command err was not the expected type *exec.ExitError: %w", err)
	}

	ecode := fixerExitCode(exitErr.ExitCode())
	if ecode != NeedsFixing {
		return nil, fmt.Errorf(
			"expected exit code %d (needs fixing), got %d (%s)",
			NeedsFixing,
			ecode,
			ecode,
		)
	}

	var result fixResult
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("could not parse JSON output %q: %w", out, err)
	}
	if len(result.Files) != 1 {
		return nil, fmt.Errorf(
			"expected 1 file to be returned by php-cs-fixer, but got %v back",
			result.Files,
		)
	}

	d, err := diff.ParseFileDiff([]byte(result.Files[0].Diff))
	if err != nil {
		return nil, fmt.Errorf("parsing php-cs-fixer diff: %w", err)
	}

	return d, nil
}

func FormatFileEdits(path string) ([]protocol.TextEdit, error) {
	diff, err := FormatFileDiff(path)
	if err != nil {
		return nil, fmt.Errorf("computing file edits: %w", err)
	}

	return functional.Map(diff.Hunks, HunkToEdit), nil
}

func FormatCodeEdits(code []byte) ([]protocol.TextEdit, error) {
	diff, err := FormatCodeDiff(code)
	if err != nil {
		return nil, fmt.Errorf("computing code edits: %w", err)
	}

	return functional.Map(diff.Hunks, HunkToEdit), nil
}

// func FormatFile(path string) ([]byte, error) {
//     d, err := FormatFileDiff(path)
//     if err != nil {
//         return nil, fmt.Errorf("computing file diff: %w", err)
//     }
//     TODO: apply edits, got some code in some test file for this (new line test?).
// }

func HunkToEdit(hunk *diff.Hunk) protocol.TextEdit {
	text := bytes.Buffer{}
	scanner := bufio.NewScanner(bytes.NewReader(hunk.Body))
	for scanner.Scan() {
		line := scanner.Bytes()
		switch line[0] {
		case '+', ' ': // Unchanged or added.
			_, _ = text.Write(line[1:])
			_ = text.WriteByte('\n') // TODO: what about \r\n?
		case '-': // Deleted lines, just don't add to replacement.
		default:
			log.Panicf("unexpected start of line: %q", line)
		}
	}

	return protocol.TextEdit{
		Range: protocol.Range{
			Start: protocol.Position{
				Line:      uint32(hunk.OrigStartLine) - 1,
				Character: 0,
			},
			End: protocol.Position{
				Line:      uint32(hunk.OrigStartLine+hunk.OrigLines) - 1,
				Character: 0,
			},
		},
		NewText: text.String(),
	}
}
