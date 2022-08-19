package logging

import (
	"fmt"
	"log"
	"path"

	"github.com/hpcloud/tail"
	"github.com/laytan/elephp/internal/config"
	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

const maxLogFiles = 2

func Configure(con config.Config) {
	lvl := toLogrusLevel(con.LogLevel())
	logrus.SetLevel(lvl)

	formatter := &logrus.TextFormatter{
		DisableQuote: true,
		PadLevelText: true,
	}

	switch con.LogOutput() {
	case config.LogOutputFile:
		logsPath := path.Join(pathutils.Root(), "logs", "elephp.log")
		fmt.Printf(
			"Logs are being printed to %s, run with `--log=stderr` if that is not desired\n",
			logsPath,
		)

		formatter.FullTimestamp = true

		logrus.SetOutput(&lumberjack.Logger{
			Filename:   logsPath,
			MaxBackups: maxLogFiles,
			LocalTime:  true,
		})

	case config.LogOutputStderr:
		// Default configuration for logrus.
	}

	logrus.SetFormatter(formatter)
}

// TODO: remove this feature.
func Tail() error {
	t, err := tail.TailFile(
		path.Join(pathutils.Root(), "logs", "elephp.log"),
		tail.Config{
			Follow:    true,
			ReOpen:    true,
			MustExist: true,
			// Start at the end of the file.
			Location: &tail.SeekInfo{Offset: 0, Whence: 2},
		},
	)
	if err != nil {
		return fmt.Errorf("Error configuring logs tail: %w", err)
	}

	for line := range t.Lines {
		log.Println(line.Text)
	}

	err = t.Wait()
	log.Println(err.Error())

	return nil
}

func toLogrusLevel(c config.LogLevel) logrus.Level {
	switch c {
	case config.LogLevelDebug:
		return logrus.DebugLevel
	case config.LogLevelInfo:
		return logrus.InfoLevel
	case config.LogLevelWarn:
		return logrus.WarnLevel
	case config.LogLevelError:
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel
	}
}
