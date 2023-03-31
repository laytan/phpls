package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/laytan/elephp/internal/config"
)

const (
	daysToKeep = 5
	hoursInDay = 24

	dateLayout = "2006-01-02"
	fileType   = ".log"

	dirPerms  = 0o755
	filePerms = 0o666
)

func Configure(root string) (stop func()) {
	logsPath := LogsPath(root)

	if err := os.MkdirAll(root, dirPerms); err != nil {
		panic(fmt.Errorf("creating logs directory %s: %w", root, err))
	}

	f, err := os.OpenFile(logsPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, filePerms)
	if err != nil {
		panic(fmt.Errorf("opening log file %s: %w", logsPath, err))
	}

	log.SetOutput(f)
	log.SetFlags(log.Ltime | log.LUTC | log.Lshortfile)

	go cleanLogs(root)

	return func() {
		if err := f.Close(); err != nil {
			panic(fmt.Errorf("closing logs file: %w", err))
		}
	}
}

func LogsPath(root string) string {
	name := config.Current.Name()

	filename := name + "-" + time.Now().Format(dateLayout) + fileType
	return filepath.Join(root, filename)
}

func cleanLogs(root string) {
	name := config.Current.Name()

	minTime := time.Now().Add(-(time.Hour * hoursInDay * daysToKeep))

	files, err := os.ReadDir(root)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		date := file.Name()
		date = strings.TrimPrefix(date, name+"-")
		date = strings.TrimSuffix(date, fileType)

		if t, err := time.Parse(dateLayout, date); err == nil {
			if !t.Before(minTime) {
				continue
			}

			if err := os.Remove(filepath.Join(root, file.Name())); err != nil {
				panic(err)
			}
		}
	}
}
