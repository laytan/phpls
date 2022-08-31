package config

import (
	"errors"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/laytan/elephp/pkg/connection"
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
	Name() string
	Version() string
	Initialize() error
	ClientPid() (uint, bool)
	ConnType() (connection.ConnType, error)
	ConnURL() string
	UseStatsviz() bool
}

type lsConfig struct {
	opts opts
	Args []string
}

func (c *lsConfig) Initialize() error {
	_, err := flags.ParseArgs(&c.opts, c.Args)
	return err //nolint:wrapcheck
}

func (c *lsConfig) ClientPid() (uint, bool) {
	isset := c.opts.ClientProcessID != 0
	return c.opts.ClientProcessID, isset
}

func (c *lsConfig) ConnType() (connection.ConnType, error) {
	connTypes := map[connection.ConnType]bool{
		connection.ConnStdio: c.opts.UseStdio,
		connection.ConnTCP:   c.opts.UseTCP,
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

func (c *lsConfig) Name() string {
	return "elephp"
}

func (c *lsConfig) Version() string {
	return "0.0.1-dev"
}
