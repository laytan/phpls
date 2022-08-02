package processwatch

import (
	"os"
	"runtime"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

const exitCheckInterval = time.Second * 30

func New(pid uint16, interval time.Duration, onExit func()) *Watcher {
	watcher := &Watcher{
		Pid:      pid,
		Interval: interval,
		OnExit:   onExit,
	}

	watcher.Start()

	return watcher
}

func NewExiter(pid uint16) *Watcher {
	watcher := &Watcher{
		Pid:      pid,
		Interval: exitCheckInterval,
		OnExit: func() {
			log.Warnf("Watched process with ID: %d has exited, exiting too\n", pid)
			os.Exit(1)
		},
	}

	watcher.Start()

	return watcher
}

type Watcher struct {
	Interval   time.Duration
	OnExit     func()
	Pid        uint16
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

func IsProcessRunning(pid uint16) bool {
	proc, err := os.FindProcess(int(pid))
	if err != nil {
		return false
	}

	if runtime.GOOS == "windows" {
		return true
	}

	err = proc.Signal(syscall.Signal(0))
	if err == nil {
		return true
	}

	if err.Error() == "os: process already finished" {
		return false
	}

	errno, ok := err.(syscall.Errno)
	if !ok {
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
