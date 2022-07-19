package config

import (
	"errors"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/laytan/elephp/pkg/connection"
)

type LogOutput string

const (
	LogOutputStderr LogOutput = "stderr"
	LogOutputFile   LogOutput = "file"
)

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

var ErrIncorrectConnTypeAmt = errors.New(
	"Elephp requires exactly one connection type to be selected",
)

func New() Config {
	return &lsConfig{
		Args: os.Args,
	}
}

type Config interface {
	Initialize() error
	ClientPid() (uint16, bool)
	ConnType() (connection.ConnType, error)
	ConnURL() string
	UseStatsviz() bool
	LogOutput() LogOutput
	LogLevel() LogLevel
}

type lsConfig struct {
	opts opts
	Args []string
}

func (c *lsConfig) Initialize() error {
	_, err := flags.ParseArgs(&c.opts, c.Args)
	return err
}

func (c *lsConfig) ClientPid() (uint16, bool) {
	isset := c.opts.ClientProcessId != 0
	return c.opts.ClientProcessId, isset
}

func (c *lsConfig) ConnType() (connection.ConnType, error) {
	connTypes := map[connection.ConnType]bool{
		connection.ConnStdio: c.opts.UseStdio,
		connection.ConnTcp:   c.opts.UseTcp,
		connection.ConnWs:    c.opts.UseWs,
	}

	var result connection.ConnType
	var found bool
	for connType, selected := range connTypes {
		if !selected {
			continue
		}

		if found {
			return "", ErrIncorrectConnTypeAmt
		}

		result = connType
		found = true
	}

	if !found {
		return "", ErrIncorrectConnTypeAmt
	}

	return result, nil
}

func (c *lsConfig) ConnURL() string {
	return c.opts.URL
}

func (c *lsConfig) UseStatsviz() bool {
	return c.opts.Statsviz
}

func (c *lsConfig) LogOutput() LogOutput {
	return c.opts.Log
}

func (c *lsConfig) LogLevel() LogLevel {
	return c.opts.LogLevel
}
