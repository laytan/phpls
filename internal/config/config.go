package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
	"github.com/danielgtaylor/huma/schema"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/xeipuuv/gojsonschema"
)

//go:generate go run config_gen.go

var Current *Schema

const (
	Name           = "elephp"
	Version        = "0.0.1-dev"
	SchemaLocation = "https://raw.githubusercontent.com/laytan/elephp/main/internal/config/elephp.schema.json"
	ok             = 0
	err            = 1
	invalid        = 2
)

// filenames are the files we check for in each of the directories that are checked.
var filenames = []string{
	"./elephp.json",
	"./.elephp.json",
	"./elephp.yml",
	"./.elephp.yml",
	"./elephp.yaml",
	"./.elephp.yaml",
}

// Parse loads the configuration from all the config files, environment variables, cli flags, and defaults.
// Validates it and sets it co Current.
// Parse exits on its own when there were errors or invalid configuration values.
// args should be all the flags, not the program name or any subcommand name.
func Parse(args []string) {
	cfg := loadConfig(args)
	maybeDumpConfig(cfg)
	applyComputedDefaults(cfg)
	validate(cfg)
	setComputed(cfg)
	Current = cfg
}

func DefaultWithoutComputed() *Schema {
	cfg := &Schema{}
	loader := aconfig.LoaderFor(cfg, aconfig.Config{
		SkipFiles: true,
		SkipEnv:   true,
		SkipFlags: true,
	})
	if err := loader.Load(); err != nil {
		panic(err)
	}
	return cfg
}

func Default() *Schema {
	cfg := DefaultWithoutComputed()
	applyComputedDefaults(cfg)
	setComputed(cfg)
	return cfg
}

func wrapUsage(flags *flag.FlagSet) {
	var files strings.Builder
	for _, fn := range filenames {
		f := strings.TrimPrefix(fn, "./")
		_, _ = files.WriteString("  ")
		_, _ = files.WriteString(f)
		_, _ = files.WriteRune('\n')
	}

	var dirs strings.Builder
	for _, dir := range configDirs() {
		_, _ = dirs.WriteString("  ")
		_, _ = dirs.WriteString(dir)
		_, _ = dirs.WriteRune('\n')
	}

	cwd, err := os.Getwd()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	} else {
		_, _ = dirs.WriteString("  ")
		_, _ = dirs.WriteString(cwd)
		_, _ = dirs.WriteString(" (Your current working directory)\n")
	}

	usage := flags.Usage
	flags.Usage = func() {
		o := flags.Output()
		_, _ = fmt.Fprintf(
			o,
			`
ElePHP - The PHP language server

Available commands:
  default  Runs the language server
  logs     Outputs the directory where logs are saved
  stubs    Outputs the directory where stubs are saved

The following file names are seen as ElePHP configuration files:
%s
These files are recognized when they are in any of the following directories:
%s
These files are checked top to bottom, with later files overwriting the former.
See %s for the configuration schema.

Configuration files are then overwritten by environment variables, with the prefix 'ELEPHP_',
so setting the php version can for example be done with 'ELEPHP_PHP_VERSION=8'.

`, files.String(), dirs.String(), SchemaLocation,
		)
		usage()
	}
}

func loadConfig(args []string) *Schema {
	cfg := &Schema{}
	ymlDecoder := aconfigyaml.New()
	files := configPaths()
	loader := aconfig.LoaderFor(cfg, aconfig.Config{
		// SkipDefaults:       false,
		// SkipFiles:          false,
		// SkipEnv:            false,
		// SkipFlags:          false,
		EnvPrefix: "ELEPHP",
		// FlagPrefix:         "",
		FlagDelimiter: ".",
		// AllFieldRequired:   false,
		// AllowDuplicates:    false,
		// AllowUnknownFields: false,
		// AllowUnknownEnvs:   false,
		// AllowUnknownFlags:  false,
		// DontGenerateTags:   false,
		// FailOnFileNotFound: false,
		// FileSystem:         nil,
		MergeFiles: true,
		FileFlag:   "config",
		Files:      files,
		// Envs:               []string{},
		// Args:               []string{},
		FileDecoders: map[string]aconfig.FileDecoder{
			".yaml": ymlDecoder,
			".yml":  ymlDecoder,
		},
	})

	flags := loader.Flags()
	wrapUsage(flags)

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(ok)
		}

		os.Exit(invalid)
	}

	checkConfigFileExists(flags)

	if err := loader.Load(); err != nil {
		var numErr *strconv.NumError
		if errors.As(err, &numErr) {
			_, _ = fmt.Fprintf(
				os.Stderr,
				"Unable to %s %q because of %v\n",
				humanFunc(numErr.Func),
				numErr.Num,
				numErr.Err,
			)
		} else {
			_, _ = fmt.Fprintln(os.Stderr, err.Error()+"\n")
		}
		os.Exit(invalid)
	}

	return cfg
}

func validate(cfg *Schema) {
	jsonSchema, err := schema.Generate(reflect.TypeOf(&Schema{}))
	if err != nil {
		panic(err) // programmer error.
	}

	sLoader := gojsonschema.NewGoLoader(jsonSchema)
	cfgLoader := gojsonschema.NewGoLoader(cfg)
	s, err := gojsonschema.NewSchema(sLoader)
	if err != nil {
		panic(err) // programmer error.
	}

	res, err := s.Validate(cfgLoader)
	if err != nil {
		panic(err) // programmer error.
	}

	errs := res.Errors()
	if len(errs) > 0 {
		_, _ = fmt.Fprintf(os.Stderr, "Invalid configuration detected:\n")
		for _, err := range errs {
			_, _ = fmt.Fprintf(os.Stderr, "  - %v\n", err)
		}
		os.Exit(invalid)
	}
}

// applyComputedDefaults applies any defaults that can't be put in struct tags.
func applyComputedDefaults(cfg *Schema) {
	if cfg.Php.Version == "" {
		phpv, err := phpversion.Get()
		if err != nil {
			_, _ = fmt.Fprintf(
				os.Stderr,
				"Unable to get php version automatically, specify the version manually or make sure `php -v` works: %v\n",
				err,
			)
			os.Exit(invalid)
		}

		cfg.Php.Version = phpv.String()
	}

	if cfg.CachePath == "" {
		dir, err := os.UserCacheDir()
		if err != nil {
			_, _ = fmt.Fprintf(
				os.Stderr,
				"Unable to get cache path automatically, try specifying the path manually: %v\n",
				err,
			)
			os.Exit(invalid)
		}

		cfg.CachePath = dir
	}
}

func setComputed(cfg *Schema) {
	cfg.LogsPath = filepath.Join(cfg.CachePath, "elephp", Version, "logs")
	if err := os.MkdirAll(cfg.LogsPath, 0o755); err != nil {
		_, _ = fmt.Fprintf(
			os.Stderr,
			"Could not create logs directory at %q, please set the cache path configuration to a writable directory: %v\n",
			cfg.LogsPath,
			err,
		)
		os.Exit(invalid)
	}

	cfg.StubsPath = filepath.Join(cfg.CachePath, "elephp", Version, "stubs")
	if err := os.MkdirAll(cfg.LogsPath, 0o755); err != nil {
		_, _ = fmt.Fprintf(
			os.Stderr,
			"Could not create stubs directory at %q, please set the cache path configuration to a writable directory: %v\n",
			cfg.StubsPath,
			err,
		)
		os.Exit(invalid)
	}

	v, ok := phpversion.FromString(cfg.Php.Version)
	if !ok {
		_, _ = fmt.Fprintf(os.Stderr, "Invalid PHP Version string %q", cfg.Php.Version)
		os.Exit(invalid)
	}
	cfg.PhpVersion = v
}

func configDirs() (dirs []string) {
	configDir, err := configDir()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	} else {
		dirs = append(dirs, configDir)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	} else {
		dirs = append(dirs, homeDir)
	}

	return dirs
}

// configPaths collects all paths to check for config files.
func configPaths() (paths []string) {
	for _, dir := range configDirs() {
		paths = append(paths, addFileNames(filepath.Join(dir, "elephp"))...)
	}
	paths = append(paths, filenames...)
	return paths
}

func addFileNames(dir string) []string {
	paths := make([]string, 0, len(filenames))
	for _, fn := range filenames {
		paths = append(paths, filepath.Join(dir, fn))
	}
	return paths
}

func configDir() (string, error) {
	// MacOS returns /Library/Application Support normally, but this is not really used for dev
	// based configuration, that is the same as unix systems most of the time.
	if runtime.GOOS == "darwin" {
		dir := os.Getenv("XDG_CONFIG_HOME")
		if dir == "" {
			dir = os.Getenv("HOME")
			if dir == "" {
				dir, err := os.UserConfigDir()
				if err != nil {
					return "", fmt.Errorf("getting darwin config dir: %w", err)
				}

				return dir, nil
			}
			dir += "/.config"
		}

		return dir, nil
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("getting config dir: %w", err)
	}

	return dir, nil
}

func humanFunc(f string) string {
	switch f {
	case "ParseBool":
		return "parse boolean"
	case "ParseInt":
		return "parse integer"
	case "ParseUint":
		return "parse unsigned integer"
	case "ParseFloat", "ParseComplex":
		return "parse number"
	default:
		return f
	}
}

func maybeDumpConfig(cfg *Schema) {
	if !cfg.DumpConfig {
		return
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "    ")
	err := enc.Encode(cfg)
	if err != nil {
		panic(err) // programmer error.
	}

	os.Exit(ok)
}

func checkConfigFileExists(flags *flag.FlagSet) {
	f := flags.Lookup("config")
	configFile := f.Value.String()
	if configFile != "" {
		_, err := os.Stat(configFile)
		_, _ = fmt.Fprintf(
			os.Stderr,
			"Non default config file (provided through the -config flag) could not be accessed: %v\n",
			err,
		)
		os.Exit(invalid)
	}
}
