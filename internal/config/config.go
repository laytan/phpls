package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jessevdk/go-flags"
	"github.com/laytan/elephp/pkg/connection"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/samber/do"
)

var ErrIncorrectConnTypeAmt = errors.New(
	"Elephp requires exactly one connection type to be selected",
)

const Usage = `<command> [-options]

Available commands:

empty   Run the language server
logs    Output the directory where logs are stored
stubs   Output the directory where generated stubs are stored`

func FromContainer() Config {
	return do.MustInvoke[Config](nil)
}

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
			LogsDir:         "",
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
	LogsDir() string
	PHPVersion() (*phpversion.PHPVersion, error)
}

type lsConfig struct {
	opts opts
	Args []string

	phpVersion   *phpversion.PHPVersion
	phpVersionMu sync.Mutex

	stubsDirMu sync.Mutex
	logsDirMu  sync.Mutex
}

func (c *lsConfig) Initialize() (disregardErr bool, err error) {
	parser := flags.NewParser(&c.opts, flags.Default)
	parser.Usage = Usage
	_, err = parser.ParseArgs(c.Args)
	if err != nil {
		var specificErr *flags.Error
		errors.As(err, &specificErr)
		helpShown := specificErr != nil && specificErr.Type == flags.ErrHelp

		return helpShown, fmt.Errorf("parsing arguments: %w", err)
	}

	return false, nil
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
	c.stubsDirMu.Lock()
	defer c.stubsDirMu.Unlock()

	if c.opts.StubsDir == "" {
		c.opts.StubsDir = c.cacheDir("stubs")
	}

	return c.opts.StubsDir
}

func (c *lsConfig) LogsDir() string {
	c.logsDirMu.Lock()
	defer c.logsDirMu.Unlock()

	if c.opts.LogsDir == "" {
		c.opts.LogsDir = c.cacheDir("logs")
	}

	return c.opts.LogsDir
}

func (c *lsConfig) PHPVersion() (*phpversion.PHPVersion, error) {
	c.phpVersionMu.Lock()
	defer c.phpVersionMu.Unlock()

	if c.phpVersion != nil {
		return c.phpVersion, nil
	}

	if c.opts.PHPVersion != "" {
		v, ok := phpversion.FromString(c.opts.PHPVersion)
		if !ok {
			return nil, fmt.Errorf("unable to parse config php version: %s", c.opts.PHPVersion)
		}

		c.phpVersion = v
		return v, nil
	}

	v, err := phpversion.Get()
	if err != nil {
		return nil, fmt.Errorf("retrieving current PHP version: %w", err)
	}

	c.phpVersion = v
	return v, nil
}

func (c *lsConfig) cacheDir(subfolder string) string {
	dir, err := os.UserCacheDir()
	if err != nil {
		panic(err)
	}

	cacheDir := filepath.Join(dir, c.Name(), c.Version(), subfolder)
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		panic(err)
	}

	return cacheDir
}
