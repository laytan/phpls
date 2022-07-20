package config

import (
	"os"
	"testing"

	"github.com/jessevdk/go-flags"
	"github.com/laytan/elephp/pkg/connection"
	"github.com/matryer/is"
)

func newTestConfig(args []string) *lsConfig {
	return &lsConfig{
		Args: args,
	}
}

type connTypeTestInput struct {
	args     []string
	connType connection.ConnType
	err      error
}

func TestConfigConnType(t *testing.T) {
	is := is.New(t)

	expectations := []connTypeTestInput{
		{
			args:     []string{"--stdio"},
			connType: connection.ConnStdio,
		},
		{
			args:     []string{"--tcp"},
			connType: connection.ConnTcp,
		},
		{
			args:     []string{"--ws"},
			connType: connection.ConnWs,
		},
		{
			args: []string{"--ws", "--tcp"},
			err:  ErrIncorrectConnTypeAmt,
		},
		{
			args: []string{"--ws", "--tcp", "--stdio"},
			err:  ErrIncorrectConnTypeAmt,
		},
		{
			args: []string{},
			err:  ErrIncorrectConnTypeAmt,
		},
	}

	for _, test := range expectations {
		c := newTestConfig(test.args)
		err := c.Initialize()
		is.NoErr(err)

		connType, err := c.ConnType()
		is.Equal(connType, test.connType)
		is.Equal(err, test.err)
	}
}

type pidTestInput struct {
	args  []string
	pid   uint16
	ok    bool
	error bool
}

func TestClientPid(t *testing.T) {
	is := is.New(t)

	expectations := []pidTestInput{
		{
			args: []string{"--ws"},
		},
		{
			args: []string{"--clientProcessId=1"},
			pid:  1,
			ok:   true,
		},
		{
			args: []string{"--clientProcessId=\"1\""},
			pid:  1,
			ok:   true,
		},
		{
			args: []string{"--clientProcessId=65535"},
			pid:  65535,
			ok:   true,
		},
		{
			args:  []string{"--clientProcessId=65536"},
			error: true,
		},
	}

	for _, test := range expectations {
		config := newTestConfig(test.args)
		err := config.Initialize()
		if !test.error {
			is.NoErr(err)
		}

		pid, ok := config.ClientPid()
		is.Equal(pid, test.pid)
		is.Equal(ok, test.ok)
	}
}

func TestStatsviz(t *testing.T) {
	is := is.New(t)

	config := newTestConfig([]string{"--statsviz"})
	err := config.Initialize()
	is.NoErr(err)
	is.True(config.UseStatsviz())

	config = newTestConfig([]string{})
	err = config.Initialize()
	is.NoErr(err)
	is.Equal(config.UseStatsviz(), false)
}

func TestConnUrl(t *testing.T) {
	is := is.New(t)

	config := newTestConfig([]string{})
	err := config.Initialize()
	is.NoErr(err)
	is.Equal(config.ConnURL(), "127.0.0.1:2001")

	config = newTestConfig([]string{"--url=\"127.0.0.1:2003\""})
	err = config.Initialize()
	is.NoErr(err)
	is.Equal(config.ConnURL(), "127.0.0.1:2003")
}

func TestConstructor(t *testing.T) {
	is := is.New(t)

	config := New()
	inner, ok := config.(*lsConfig)
	is.Equal(ok, true)
	is.Equal(inner.Args, os.Args)
}

func TestHelp(t *testing.T) {
	is := is.New(t)

	expectations := [][]string{
		{"--help"},
		{"-h"},
	}

	for _, test := range expectations {
		config := newTestConfig(test)
		err := config.Initialize()
		flagError, ok := err.(*flags.Error)

		is.True(ok)
		is.Equal(flagError.Type, flags.ErrHelp)
	}
}
