package phpcbf

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
)

// Instance is a wrapper around the 'phpcbf' cli for formatting code.
// This implementation preemptively starts an instance of the command,
// So that each call to Format does not need to wait for php and the phpcbf bootstrap.
// This does mean that there will always be a 'phpcbf' process instance running.
type Instance struct {
	startErr   error
	cmd        *exec.Cmd
	in         io.WriteCloser
	out        io.Reader
	executable string
	mu         sync.Mutex
}

// NewInstance initializes a new instance.
func NewInstance() *Instance {
	p := &Instance{}
	p.Init()
	return p
}

func (p *Instance) HasExecutable() bool {
	return p.executable != ""
}

func (p *Instance) Init() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.executable = executable()
	if p.executable == "" {
		return
	}

	p.reset()
}

func (p *Instance) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd != nil {
		return
	}

	if err := p.cmd.Process.Kill(); err != nil {
		log.Println(fmt.Errorf("[ERROR]: killing phpcbf process: %w", err))
	}

	if err := p.cmd.Wait(); err != nil {
		log.Println(fmt.Errorf("[ERROR]: waiting for cmd exit: %w", err))
	}
}

func (p *Instance) Format(code []byte) ([]byte, error) {
	p.mu.Lock()

	if p.executable == "" {
		return nil, fmt.Errorf("no phpcbf executable found")
	}

	if p.startErr != nil {
		return nil, p.startErr
	}

	defer func() {
		go func() {
			defer p.mu.Unlock()
			p.reset()
		}()
	}()

	var stdinErr error
	go func() {
		defer p.in.Close()
		if _, err := p.in.Write(code); err != nil {
			stdinErr = fmt.Errorf("writing code to stdin: %w", err)
		}
	}()

	out, err := io.ReadAll(p.out)
	if err != nil {
		return nil, fmt.Errorf("reading stdout: %w", err)
	}

	var exitErr *exec.ExitError
	err = p.cmd.Wait()
	if err != nil {
		if !errors.As(err, &exitErr) {
			return nil, fmt.Errorf("returned non exitError error: %w", err)
		}

		if !exitErr.Success() && exitErr.ExitCode() != 1 {
			return nil, fmt.Errorf("%w: %s", exitErr, out)
		}
	}

	if stdinErr != nil {
		return nil, fmt.Errorf("writing code to stdin: %w", stdinErr)
	}

	return out, nil
}

func (p *Instance) FormatFileEdits(path string) ([]protocol.TextEdit, error) {
	code := wrkspc.Current.FContentOf(path)
	lines := len(strings.Split(code, "\n")) // TODO: ewh.
	formatted, err := p.Format([]byte(code))
	if err != nil {
		return nil, fmt.Errorf("formatting %q code: %w", path, err)
	}

	return []protocol.TextEdit{{
		Range: protocol.Range{
			Start: protocol.Position{
				Line:      0,
				Character: 0,
			},
			End: protocol.Position{
				Line:      uint32(lines + 1),
				Character: 0,
			},
		},
		NewText: string(formatted),
	}}, nil
}

func (p *Instance) reset() {
	var err error
	p.startErr = nil
	p.cmd = exec.Command(p.executable, "-q", "-") // nolint:gosec // p.executable is safe.
	p.in, err = p.cmd.StdinPipe()
	if err != nil {
		p.startErr = fmt.Errorf("connecting stdin: %w", err)
		return
	}

	p.out, err = p.cmd.StdoutPipe()
	if err != nil {
		p.startErr = fmt.Errorf("connecting stdout: %w", err)
		return
	}

	if err = p.cmd.Start(); err != nil {
		p.startErr = fmt.Errorf("starting phpcbf: %w", err)
		return
	}
}

func executable() string {
	localPath := filepath.Join("vendor", "bin", "phpcbf")
	p, err := exec.LookPath(localPath)
	if err == nil {
		return p
	}

	if !errors.Is(err, fs.ErrNotExist) {
		log.Println(fmt.Errorf("[WARN]: unexpected error checking %q: %w", localPath, err))
	}

	p, err = exec.LookPath("phpcbf")
	if err != nil {
		if !errors.Is(err, exec.ErrNotFound) {
			log.Println(fmt.Errorf("[WARN]: unexpected error checking path for phpcbf: %w", err))
		}

		return ""
	}

	return p
}
