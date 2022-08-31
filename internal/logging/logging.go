package logging

import (
	"log"
	"os"
	"path"
	"strings"
	"time"
)

const (
	daysToKeep = 14
	hoursInDay = 24

	dateLayout = "2006-01-02"
	fileType   = ".log"

	filePerms = 0666
)

func Configure(root string, name string) (stop func()) {
	logsPath := getLogsPath(root, name)

	f, err := os.OpenFile(logsPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, filePerms)
	if err != nil {
		panic(err)
	}

	log.SetOutput(f)
	log.SetFlags(log.Ltime | log.LUTC | log.Lshortfile)

	go cleanLogs(root, name)

	return func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}
}

func getLogsPath(root, name string) string {
	filename := name + "-" + time.Now().Format(dateLayout) + fileType
	return path.Join(root, filename)
}

func cleanLogs(root, name string) {
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

			if err := os.Remove(path.Join(root, file.Name())); err != nil {
				panic(err)
			}
		}
	}
}
