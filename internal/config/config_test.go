package config_test

import (
	"fmt"
	"testing"

	"github.com/laytan/elephp/internal/config"
	"github.com/laytan/elephp/pkg/connection"
	"github.com/stretchr/testify/require"
)

func newTestConfig(args []string) config.Config {
	return config.NewWithArgs(args)
}

type connTypeTestInput struct {
	args     []string
	connType connection.ConnType
	err      error
}

func TestConfigConnType(t *testing.T) {
	t.Parallel()

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
			err:  config.ErrIncorrectConnTypeAmt,
		},
		{
			args: []string{"--ws", "--tcp", "--stdio"},
			err:  config.ErrIncorrectConnTypeAmt,
		},
		{
			args: []string{},
			err:  config.ErrIncorrectConnTypeAmt,
		},
	}

	for i, test := range expectations {
		i, test := i, test
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()
			c := newTestConfig(test.args)
			shownHelp, err := c.Initialize()
			require.False(t, shownHelp)
			require.NoError(t, err)

			connType, err := c.ConnType()
			require.Equal(t, connType, test.connType)
			require.Equal(t, err, test.err)
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
		i, test := i, test
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()
			config := newTestConfig(test.args)
			shownHelp, err := config.Initialize()
			require.False(t, shownHelp)

			if !test.error {
				require.NoError(t, err)
			}

			pid, ok := config.ClientPid()
			require.Equal(t, pid, test.pid)
			require.Equal(t, ok, test.ok)
		})
	}
}

func TestStatsviz(t *testing.T) {
	t.Parallel()

	config := newTestConfig([]string{"--statsviz"})
	shownHelp, err := config.Initialize()
	require.False(t, shownHelp)
	require.NoError(t, err)
	require.True(t, config.UseStatsviz())

	config = newTestConfig([]string{})
	shownHelp, err = config.Initialize()
	require.False(t, shownHelp)
	require.NoError(t, err)
	require.False(t, config.UseStatsviz())
}

func TestConnUrl(t *testing.T) {
	t.Parallel()

	config := newTestConfig([]string{})
	shownHelp, err := config.Initialize()
	require.False(t, shownHelp)
	require.NoError(t, err)
	require.Equal(t, config.ConnURL(), "127.0.0.1:2001")

	config = newTestConfig([]string{"--url=\"127.0.0.1:2003\""})
	shownHelp, err = config.Initialize()
	require.False(t, shownHelp)
	require.NoError(t, err)
	require.Equal(t, config.ConnURL(), "127.0.0.1:2003")
}

func TestHelp(t *testing.T) {
	t.Parallel()

	expectations := [][]string{
		{"--help"},
		{"-h"},
	}

	for i, test := range expectations {
		i, test := i, test
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()
			config := newTestConfig(test)
			shownHelp, _ := config.Initialize()
			require.True(t, shownHelp)
		})
	}
}
