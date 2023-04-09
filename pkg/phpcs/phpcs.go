package phpcs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"syscall"
)

// Instance is a wrapper around the 'phpcs' cli for formatting code.
// This implementation preemptively starts an instance of the command,
// So that each call to Format does not need to wait for php and the phpcs bootstrap.
// This does mean that there will always be a 'phpcs' process instance running.
type Instance struct {
	startErr   error
	cmd        *exec.Cmd
	in         io.WriteCloser
	out        io.Reader
	executable string
	mu         sync.Mutex
}

// New initializes a new instance.
func New(executable string) *Instance {
	p := &Instance{executable: executable}
	p.Init()
	return p
}

func (p *Instance) Init() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.reset()
}

func (p *Instance) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cmd != nil {
		return
	}

	if err := p.cmd.Process.Kill(); err != nil {
		log.Println(fmt.Errorf("[ERROR]: killing phpcs process: %w", err))
	}

	if err := p.cmd.Wait(); err != nil {
		log.Println(fmt.Errorf("[ERROR]: waiting for cmd exit: %w", err))
	}
}

type Severity string

const (
	Warning Severity = "WARNING"
	Error   Severity = "ERROR"
)

type Report struct {
	Totals struct {
		Errors   int
		Warnings int
		Fixable  int
	}
	Files map[string]struct { // Stdin is at Files["STDIN"].
		Errors   int
		Warnings int
		Messages []*ReportMessage
	}
}

type ReportMessage struct {
	Msg      string `json:"message"`
	Source   string
	Severity int // TODO: enum
	Fixable  bool
	Type     Severity
	Row      int `json:"line"`
	Column   int
}

var ErrCancelled = errors.New("cancelled")

func (p *Instance) Check(ctx context.Context, code []byte) (*Report, error) {
	p.mu.Lock()

	if p.startErr != nil {
		p.mu.Unlock()
		return nil, p.startErr
	}

	// Calling reset in a goroutine so the results are returned before starting that.
	// Resulting in no time waiting on the reset and starting of phpcs.
	defer func() {
		go func() {
			p.reset()
			p.mu.Unlock()
		}()
	}()

	errorC := make(chan error)
	go func() {
		defer p.in.Close()
		if _, err := p.in.Write(code); err != nil {
			errorC <- fmt.Errorf("writing code to phpcs stdin: %w", err)
		}
	}()

	outC := make(chan []byte)
	go func() {
		out, err := io.ReadAll(p.out)
		if err != nil {
			errorC <- fmt.Errorf("reading phpcs stdout: %w", err)
			outC <- nil
			return
		}
		err = p.cmd.Wait()

		if aerr := checkPhpcsErr(out, err); aerr != nil {
			errorC <- aerr
			outC <- nil
			return
		}

		outC <- out
	}()

	var report *Report
	var errs []error
	doneC := ctx.Done()
Loop:
	for {
		select {
		case <-doneC:
			log.Println("[DEBUG]: killing phpcs process")
			if err := p.cmd.Process.Signal(syscall.SIGKILL); err != nil {
				errs = append(errs, fmt.Errorf("killing phpcs process because context done: %w", err))
			}
			// Make sure we only kill once.
			doneC = nil
		case err := <-errorC:
			errs = append(errs, err)
		case out := <-outC:
			if len(out) == 0 {
				break Loop
			}

			var err error
			report, err = newReport(out)
			if err != nil {
				errs = append(errs, err)
			}

			break Loop
		}
	}

	if len(errs) == 1 {
		return report, errs[0]
	}

	if len(errs) > 1 {
		return nil, fmt.Errorf("multiple errors during phpcs check: %w", errors.Join(errs...))
	}

	return report, nil
}

func newReport(out []byte) (*Report, error) {
	var report Report
	if err := json.Unmarshal(out, &report); err != nil {
		return nil, fmt.Errorf("parsing phpcs JSON report: %w", err)
	}

	if _, ok := report.Files["STDIN"]; !ok {
		return nil, fmt.Errorf("no report file for stdin returned in report: %v", report)
	}

	return &report, nil
}

type ExitCode int

const (
	ExitCodeKilled ExitCode = iota - 1
	ExitCodeNoIssues
	ExitCodeIssues
	ExitCodeFixableIssues
	ExitCodeProcessingError
)

func checkPhpcsErr(stdout []byte, err error) error {
	if err == nil {
		return nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		switch ExitCode(exitErr.ExitCode()) {
		case ExitCodeKilled:
			return ErrCancelled
		case ExitCodeNoIssues, ExitCodeIssues, ExitCodeFixableIssues:
			return nil
		default: // Any other codes are actual errors, error messages are output on stdout.
			return fmt.Errorf("phpcs error: %s, %w", stdout, err)
		}
	}

	// Error is not an ExitError, something has gone very wrong.
	return fmt.Errorf("unexpected error executing phpcs: %w", err)
}

func (p *Instance) reset() {
	var err error
	p.startErr = nil
	p.cmd = exec.Command( // nolint:gosec // p.executable is safe.
		p.executable,
		"-q",
		"--report=json",
		"-",
	)
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
		p.startErr = fmt.Errorf("starting phpcs: %w", err)
		return
	}
}
