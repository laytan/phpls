// nolint:gosec // Lots of false positives.
package phpcs

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "embed"

	"github.com/laytan/elephp/internal/config"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/functional"
	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
	"github.com/sourcegraph/go-diff/diff"
)

var instance = &phpcs{}

//nolint:typecheck // File created during build.
//go:embed formatter.gz
var formatterSourceGzipped []byte

type phpcs struct {
	connectErr error

	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdoutC chan []byte // Line delimited stdout.
	stderrC chan []byte // Line delimited stderr.

	cancelFunc context.CancelFunc // The cancel for the command, called in Disconnect.

	mu sync.Mutex // Should not send new code to format before other is done.
}

func (p *phpcs) Connect() (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd != nil {
		return nil
	}

	if p.connectErr != nil {
		return p.connectErr
	}

	defer func() {
		p.connectErr = err
	}()

	if err := ExtractBinary(); err != nil {
		return fmt.Errorf("extracting php-cs-fixer-daemon binary: %w", err)
	}

	ctx, c := context.WithCancel(context.Background())
	p.cancelFunc = c
	cmd := exec.CommandContext(ctx, daemonPath())

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("connecting stdin: %w", err)
	}
	p.stdin = stdin

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("connecting stdout: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("connecting stderr: %w", err)
	}

	p.stderrC = make(chan []byte)
	go func() {
		s := bufio.NewScanner(stderr)
		for s.Scan() {
			log.Printf("PHPCS Fixer stderr: %s", s.Bytes())
			p.stderrC <- s.Bytes()
		}
	}()

	p.stdoutC = make(chan []byte)
	go func() {
		s := bufio.NewScanner(stdout)
		for s.Scan() {
			log.Printf("PHPCS Fixer stdout: %s", s.Bytes())
			p.stdoutC <- s.Bytes()
		}
	}()

	if err = cmd.Start(); err != nil {
		return fmt.Errorf("starting formatter: %w", err)
	}

	p.cmd = cmd
	return nil
}

func (p *phpcs) Disconnect() {
	if p.cmd == nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.cancelFunc()
	if err := p.cmd.Wait(); err != nil {
		log.Println(err)
	}
	p.cmd = nil
}

func (p *phpcs) Format(code string) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	enc, err := json.Marshal(code)
	if err != nil {
		return "", fmt.Errorf("json encoding code: %w", err)
	}

	enc = append(enc, byte('\n'))
	if _, err := p.stdin.Write(enc); err != nil {
		return "", fmt.Errorf("writing code to stdin: %w", err)
	}

	timeout := time.NewTimer(time.Second * 5)
	select {
	case errMsg := <-p.stderrC:
		return "", fmt.Errorf("formatter returned error: %s", errMsg)
	case outMsg := <-p.stdoutC:
		var res string
		if err := json.Unmarshal(outMsg, &res); err != nil {
			return "", fmt.Errorf("decoding json response: %w", err)
		}
		return res, nil
	case <-timeout.C:
		// 5 seconds no response, maybe daemon is fucked?
		go CloseDaemon()
		return "", fmt.Errorf("formatting exceeded timeout of 5s")
	}
}

// ExtractBinary extracts and unzips the formatter binary into `daemonPath()`.
func ExtractBinary() error {
	defer func() { formatterSourceGzipped = nil }() // No need to keep in memory.
	if formatterSourceGzipped == nil {
		log.Printf("php-cs-fixer-daemon has already been extracted.")
		return nil
	}

	path := daemonPath()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o770)
	if err != nil {
		if os.IsExist(err) { // Binary exists, that's ok.
			log.Printf("%q exists, not extracting formatter binary.", path)
			return nil
		}

		return fmt.Errorf("creating file %q: %w", path, err)
	}
	defer f.Close()

	log.Printf("Extracting %q", path)
	r, err := gzip.NewReader(bytes.NewReader(formatterSourceGzipped))
	if err != nil {
		return fmt.Errorf("creating gzip reader: %w", err)
	}

	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("unzipping and writing binary to %q: %w", path, err)
	}

	return nil
}

func CloseDaemon() {
	instance.Disconnect()
}

func FormatFileDiff(path string) (*diff.FileDiff, error) {
	return FormatCodeDiff(wrkspc.Current.FContentOf(path))
}

func FormatCodeDiff(code string) (*diff.FileDiff, error) {
	if err := instance.Connect(); err != nil {
		return nil, fmt.Errorf("connecting to formatter: %w", err)
	}

	formatted, err := instance.Format(code)
	if err != nil {
		return nil, fmt.Errorf("formatting code: %w", err)
	}

	if strings.TrimSpace(formatted) == "" {
		return &diff.FileDiff{}, nil
	}

	d, err := diff.ParseFileDiff([]byte(formatted))
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

func FormatCodeEdits(code string) ([]protocol.TextEdit, error) {
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

func daemonPath() string {
	dir := config.Current.BinDir()
	return filepath.Join(dir, "php-cs-fixer-daemon")
}
