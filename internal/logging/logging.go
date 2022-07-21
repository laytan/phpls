package logging

import (
	"os"

	"github.com/laytan/elephp/internal/config"
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

func Configure(con config.Config) {
	lvl := toLogrusLevel(con.LogLevel())
	logrus.SetLevel(lvl)
	if lvl >= logrus.InfoLevel {
		logrus.SetReportCaller(true)
	}

	logrus.SetFormatter(
		&logrus.TextFormatter{
			ForceColors:   true,
			DisableQuote:  true,
			PadLevelText:  true,
			FullTimestamp: true,
		},
	)

	switch con.LogOutput() {
	case config.LogOutputFile:
		logrus.SetOutput(&lumberjack.Logger{
			Filename: os.TempDir() + "elephp.log",
		})
	case config.LogOutputStderr:
		// Default configuration for logrus.
	}
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
