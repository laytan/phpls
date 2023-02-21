package lsprogress

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
)

type Tracker struct {
	client protocol.Client

	mu         sync.Mutex
	inProgress map[protocol.ProgressToken]*WorkDone
}

func NewTracker(client protocol.Client) *Tracker {
	return &Tracker{
		client:     client,
		inProgress: make(map[protocol.ProgressToken]*WorkDone),
	}
}

func (t *Tracker) Start(
	ctx context.Context,
	title, message string,
	token protocol.ProgressToken,
) (*WorkDone, error) {
	wd := &WorkDone{
		client: t.client,
		token:  token,
	}

	if wd.token == nil {
		token = strconv.FormatInt(
			rand.Int63(), //nolint:gosec // Don't care about security with this int.
			10,
		)
		err := wd.client.WorkDoneProgressCreate(ctx, &protocol.WorkDoneProgressCreateParams{
			Token: token,
		})
		if err != nil {
			return nil, fmt.Errorf("starting work for %s: %w", title, err)
		}
		wd.token = token
	}

	// At this point we have a token that the client knows about. Store the token
	// before starting work.
	t.mu.Lock()
	t.inProgress[wd.token] = wd
	t.mu.Unlock()
	wd.cleanup = func() {
		t.mu.Lock()
		delete(t.inProgress, token)
		t.mu.Unlock()
	}
	err := wd.client.Progress(ctx, &protocol.ProgressParams{
		Token: wd.token,
		Value: &protocol.WorkDoneProgressBegin{
			Kind:    "begin",
			Message: message,
			Title:   title,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("reporting first progress for %s: %w", title, err)
	}

	return wd, nil
}

func (t *Tracker) Track(
	ctx context.Context,
	done, total func() float64,
	title string,
	interval time.Duration,
) (stop func(err error) error, err error) {
	progress, err := t.Start(
		ctx,
		title,
		"Started",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("starting progress track: %w", err)
	}

	var reportErr error

	timer := time.NewTicker(interval)
	stopped := make(chan any)
	go func() {
		for {
			select {
			case <-stopped:
				return
			case <-timer.C:
				d, t := done(), total()
				msg := fmt.Sprintf("%d/%d", int(d), int(t))
				if err := progress.Report(ctx, msg, (d / t * 100)); err != nil {
					reportErr = err
				}
			}
		}
	}()

	return func(err error) error {
		timer.Stop()
		stopped <- struct{}{}
		if err != nil {
			err = progress.End(ctx, err.Error())
			if err != nil {
				return fmt.Errorf("ending progress on error: %w", err)
			}

			return nil
		}

		if reportErr != nil {
			err := progress.End(ctx, reportErr.Error())
			if err != nil {
				return fmt.Errorf("multiple errors reporting and ending progress: %w", err)
			}

			return fmt.Errorf("reporting progress: %w", reportErr)
		}

		err = progress.End(ctx, "Completed")
		if err != nil {
			return fmt.Errorf("ending progress on success: %w", err)
		}

		return nil
	}, nil
}

// WorkDone represents a unit of work that is reported to the client via the
// progress API.
type WorkDone struct {
	client protocol.Client
	token  protocol.ProgressToken

	cleanup func()
}

func (wd *WorkDone) Token() protocol.ProgressToken {
	return wd.token
}

// Report reports an update on WorkDone report back to the client.
func (wd *WorkDone) Report(ctx context.Context, message string, percentage float64) error {
	err := wd.client.Progress(ctx, &protocol.ProgressParams{
		Token: wd.token,
		Value: &protocol.WorkDoneProgressReport{
			Kind:       "report",
			Message:    message,
			Percentage: uint32(percentage),
		},
	})
	if err != nil {
		return fmt.Errorf("reporting progress: %w", err)
	}

	return nil
}

// End reports a workdone completion back to the client.
func (wd *WorkDone) End(ctx context.Context, message string) error {
	err := wd.client.Progress(ctx, &protocol.ProgressParams{
		Token: wd.token,
		Value: &protocol.WorkDoneProgressEnd{
			Kind:    "end",
			Message: message,
		},
	})
	if err != nil {
		return fmt.Errorf("sending end work: %w", err)
	}

	if wd.cleanup != nil {
		wd.cleanup()
	}

	return nil
}
