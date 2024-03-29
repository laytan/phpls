package processwatch

import (
	"errors"
	"log"
	"os"
	"runtime"
	"syscall"
	"time"
)

const exitCheckInterval = time.Second * 30

func New(pid uint, interval time.Duration, onExit func()) *Watcher {
	watcher := &Watcher{
		Pid:      pid,
		Interval: interval,
		OnExit:   onExit,
	}

	watcher.Start()

	return watcher
}

func NewExiter(pid uint) *Watcher {
	watcher := &Watcher{
		Pid:      pid,
		Interval: exitCheckInterval,
		OnExit: func() {
			log.Printf("Watched process with ID: %d has exited, exiting too\n", pid)
			os.Exit(1)
		},
	}

	watcher.Start()

	return watcher
}

type Watcher struct {
	Interval   time.Duration
	OnExit     func()
	Pid        uint
	isWatching bool
}

func (w *Watcher) Start() {
	if w.isWatching {
		return
	}

	ticker := time.NewTicker(w.Interval)

	go func() {
		for {
			<-ticker.C

			if !IsProcessRunning(w.Pid) {
				w.OnExit()
			}
		}
	}()
}

func IsProcessRunning(pid uint) bool {
	proc, err := os.FindProcess(int(pid))
	if err != nil {
		return false
	}

	if runtime.GOOS == "windows" { // If we find a process on windows, it is always running.
		return true
	}

	err = proc.Signal(syscall.Signal(0))
	if err == nil {
		return true
	}

	if errors.Is(err, os.ErrProcessDone) {
		return false
	}

	var errno syscall.Errno
	if !errors.As(err, &errno) {
		return false
	}

	switch errno {
	case syscall.ESRCH:
		return false
	case syscall.EPERM:
		return true
	default:
	}

	return false
}
