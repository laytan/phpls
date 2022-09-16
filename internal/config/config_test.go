package config

import (
	"fmt"
	"os"
	"testing"

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
	t.Parallel()
	is := is.New(t)

	expectations := []connTypeTestInput{
		{
			args:     []string{"--stdio"},
			connType: connection.ConnStdio,
		},
		{
			args:     []string{"--tcp"},
			connType: connection.ConnTCP,
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

	for i, test := range expectations {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()
			is := is.New(t)
			c := newTestConfig(test.args)
			shownHelp, err := c.Initialize()
			is.Equal(shownHelp, false)
			is.NoErr(err)

			connType, err := c.ConnType()
			is.Equal(connType, test.connType)
			is.Equal(err, test.err)
		})
	}
}

type pidTestInput struct {
	args  []string
	pid   uint
	ok    bool
	error bool
}

func TestClientPid(t *testing.T) {
	t.Parallel()
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
			args: []string{"--clientProcessId=65536"},
			pid:  65536,
			ok:   true,
		},
	}

	for i, test := range expectations {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()
			is := is.New(t)
			config := newTestConfig(test.args)
			shownHelp, err := config.Initialize()
			is.Equal(shownHelp, false)

			if !test.error {
				is.NoErr(err)
			}

			pid, ok := config.ClientPid()
			is.Equal(pid, test.pid)
			is.Equal(ok, test.ok)
		})
	}
}

func TestStatsviz(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	config := newTestConfig([]string{"--statsviz"})
	shownHelp, err := config.Initialize()
	is.Equal(shownHelp, false)
	is.NoErr(err)
	is.True(config.UseStatsviz())

	config = newTestConfig([]string{})
	shownHelp, err = config.Initialize()
	is.Equal(shownHelp, false)
	is.NoErr(err)
	is.Equal(config.UseStatsviz(), false)
}

func TestConnUrl(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	config := newTestConfig([]string{})
	shownHelp, err := config.Initialize()
	is.Equal(shownHelp, false)
	is.NoErr(err)
	is.Equal(config.ConnURL(), "127.0.0.1:2001")

	config = newTestConfig([]string{"--url=\"127.0.0.1:2003\""})
	shownHelp, err = config.Initialize()
	is.Equal(shownHelp, false)
	is.NoErr(err)
	is.Equal(config.ConnURL(), "127.0.0.1:2003")
}

func TestConstructor(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	config := New()
	inner, ok := config.(*lsConfig)
	is.Equal(ok, true)
	is.Equal(inner.Args, os.Args)
}

func TestHelp(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	expectations := [][]string{
		{"--help"},
		{"-h"},
	}

	for i, test := range expectations {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()
			is := is.New(t)
			config := newTestConfig(test)
			shownHelp, _ := config.Initialize()
			is.Equal(shownHelp, true)
		})
	}
}
