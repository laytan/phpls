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
	Initialize() error
	ClientPid() (uint16, bool)
	ConnType() (connection.ConnType, error)
	ConnURL() string
	UseStatsviz() bool
}

type opts struct {
	ClientProcessId uint16 `long:"clientProcessId" description:"Process ID that when terminated, terminates the language server"`
	UseStdio        bool   `long:"stdio"           description:"Communicate over stdio"`
	UseWs           bool   `long:"ws"              description:"Communicate over websockets"`
	UseTcp          bool   `long:"tcp"             description:"Communicate over TCP"`
	URL             string `long:"url"             description:"The URL to listen on for tcp or websocket connections"              default:"127.0.0.1:2001"`
	Statsviz        bool   `long:"statsviz"        description:"Visualize stats(CPU, memory etc.) on localhost:6060/debug/statsviz"`
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
			return 0, ErrIncorrectConnTypeAmt
		}

		result = connType
		found = true
	}

	if !found {
		return 0, ErrIncorrectConnTypeAmt
	}

	return result, nil
}

func (c *lsConfig) ConnURL() string {
	return c.opts.URL
}

func (c *lsConfig) UseStatsviz() bool {
	return c.opts.Statsviz
}
