package phpstan

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type Report struct {
	Totals struct {
		Errors     int
		FileErrors int
	}
	Files map[string]struct {
		Errors   int
		Messages []*ReportMessage
	}
	Errors []any
}

type ReportMessage struct {
	Msg       string `json:"message"`
	Ln        int    `json:"line"`
	Ignorable bool
}

var ErrCancelled = errors.New("cancelled")

func Analyze(
	ctx context.Context,
	executable string,
	path string,
	content []byte,
) ([]*ReportMessage, error) {
	dir, name := filepath.Split(path)
	fh, err := os.CreateTemp(dir, ".phpstan-tmp."+name)
	if err != nil {
		return nil, fmt.Errorf("creating temp file for phpstan: %w", err)
	}

	if _, err := fh.Write(content); err != nil {
		return nil, fmt.Errorf("writing to temp file %q: %w", fh.Name(), err)
	}

	defer func() {
		go func() {
			p := fh.Name()
			if err := fh.Close(); err != nil {
				log.Println(fmt.Errorf("[ERROR]: phpstan closing temp file %q: %w", p, err))
			}
			if err := os.Remove(fh.Name()); err != nil {
				log.Println(fmt.Errorf("[ERROR]: phpstan removing temp file %q: %w", p, err))
			}
		}()
	}()

	return AnalyzePath(ctx, executable, fh.Name())
}

func AnalyzePath(ctx context.Context, executable string, path string) ([]*ReportMessage, error) {
	cmd := exec.CommandContext(
		ctx,
		executable,
		"analyze",
		"--error-format",
		"json",
		"--no-progress",
		path,
	)
	out, err := cmd.Output()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if exitErr.ExitCode() == -1 {
			return nil, ErrCancelled
		}
		if exitErr.ExitCode() > 1 {
			return nil, fmt.Errorf("running phpstan: %w: %s", err, exitErr.Stderr)
		}
	} else if err != nil {
		return nil, fmt.Errorf("running phpstan: %w", err)
	}

	var report Report
	if err := json.Unmarshal(out, &report); err != nil {
		// TODO: detect invalid configuration errors.

		// If the file did not contain any errors, the files key is an array,
		// we are then unmarshalling into an object (which it would be if there were errors).
		// So, failing this basically means there are no errors.
		return nil, nil
	}

	return report.Files[path].Messages, nil
}
