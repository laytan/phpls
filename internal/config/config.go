package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/kirsle/configdir"
	"github.com/laytan/elephp/pkg/connection"
	"github.com/laytan/elephp/pkg/phpversion"
)

var ErrIncorrectConnTypeAmt = errors.New(
	"Elephp requires exactly one connection type to be selected",
)

func New() Config {
	return &lsConfig{Args: os.Args}
}

func NewWithArgs(args []string) Config {
	return &lsConfig{Args: args}
}

func Default() Config {
	return &lsConfig{
		opts: opts{
			UseStdio:        true,
			UseWs:           false,
			UseTCP:          false,
			Statsviz:        false,
			ClientProcessID: 0,
			URL:             "",
			FileExtensions:  []string{"php"},
			IgnoredDirNames: []string{".git", "node_modules"},
			StubsDir:        "",
			PHPVersion:      "",
		},
	}
}

type Config interface {
	Name() string
	Version() string
	Initialize() (disregardErr bool, err error)
	ClientPid() (uint, bool)
	ConnType() (connection.ConnType, error)
	ConnURL() string
	UseStatsviz() bool
	FileExtensions() []string
	IgnoredDirNames() []string
	StubsDir() string
	PHPVersion() (*phpversion.PHPVersion, error)
}

type lsConfig struct {
	opts opts
	Args []string

	phpVersion *phpversion.PHPVersion
}

func (c *lsConfig) Initialize() (shownHelp bool, err error) {
	_, err = flags.ParseArgs(&c.opts, c.Args)
	if err == nil {
		return false, nil
	}

	var specificErr *flags.Error
	if !errors.As(err, &specificErr) {
		return false, fmt.Errorf("unexpected error parsing flags: %w", err)
	}

	if specificErr.Type == flags.ErrHelp {
		return true, nil
	}

	return false, fmt.Errorf("Could not initialize config: %w", specificErr)
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

func (c *lsConfig) FileExtensions() []string {
	exts := make([]string, 0, len(c.opts.FileExtensions))
	for _, ext := range c.opts.FileExtensions {
		exts = append(exts, "."+strings.TrimSpace(ext))
	}

	return exts
}

func (c *lsConfig) IgnoredDirNames() []string {
	dirs := make([]string, 0, len(c.opts.IgnoredDirNames))
	for _, dir := range c.opts.IgnoredDirNames {
		dirs = append(dirs, strings.TrimSpace(dir))
	}

	return dirs
}

func (c *lsConfig) StubsDir() string {
	if c.opts.StubsDir == "" {
		c.opts.StubsDir = configdir.LocalCache("elephp", c.Version(), "stubs")
	}

	return c.opts.StubsDir
}

func (c *lsConfig) PHPVersion() (*phpversion.PHPVersion, error) {
	if c.phpVersion != nil {
		return c.phpVersion, nil
	}

	if c.opts.PHPVersion == "" {
		v, err := phpversion.Get()
		if err != nil {
			return nil, fmt.Errorf("retrieving current PHP version: %w", err)
		}

		c.phpVersion = v
		return v, nil
	}

	v, err := phpversion.Get()
	if err != nil {
		return nil, fmt.Errorf("getting current php version: %w", err)
	}

	return v, nil
}
