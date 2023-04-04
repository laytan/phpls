package phpcs

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/functional"
	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
	"github.com/sourcegraph/go-diff/diff"
)

// TODO: Temporary, maybe a pool of them.
var instance = &phpcs{}

// TODO: won't work when only got the executable, needs work.
var phar = filepath.Join(pathutils.Root(), "bin", "formatter")

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

	// TODO: context should be passed in here, ideally from the start,
	// so that this stops when elephp stops.
	ctx, c := context.WithCancel(context.Background())
	p.cancelFunc = c
	cmd := exec.CommandContext(ctx, phar)

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
