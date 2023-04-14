package config

import (
	"github.com/laytan/phpls/pkg/connection"
	"github.com/laytan/phpls/pkg/phpversion"
)

type DiagnosticsMethod string

const (
	DiagnosticsOnSave   DiagnosticsMethod = "ON_SAVE"
	DiagnosticsOnChange DiagnosticsMethod = "ON_CHANGE"
)

type Schema struct {
	// TODO: implement usage of this.
	Php Php `json:"php,omitempty"`
	// TODO: implement usage of this.
	Phpcbf             Phpcbf      `json:"phpcbf,omitempty"`
	Diagnostics        Diagnostics `json:"diagnostics,omitempty"`
	Extensions         []string    `json:"extensions,omitempty"          uniqueItems:"true" minItems:"1" default:".php"              doc:"File extensions to consider PHP code."                                                    usage:"File extensions to consider PHP code."`
	IgnoredDirectories []string    `json:"ignored_directories,omitempty" uniqueItems:"true"              default:".git,node_modules" doc:"Directories to ignore completely, use when you have huge directories with non-php files." usage:"Directories to ignore completely, use when you have huge directories with non-php files." flag:"ignored-directories"`
	Server             Server      `json:"server,omitempty"`
	Statsviz           Statsviz    `json:"statsviz,omitempty"`
	CachePath          string      `json:"cache_path,omitempty"                                                                      doc:"Root directory for generated stubs and logs, defaults to the user cache directory."       usage:"Root directory for generated stubs and logs, defaults to the user cache directory."       flag:"cache-path"`
	DumpConfig         bool        `json:"dump_config,omitempty"                                         default:"false"             doc:"Dump the resolved config before validation, useful for debugging."                        usage:"Dump the resolved config before validation, useful for debugging."                        flag:"dump-config"`

	LogsPath   string                 `json:"-" flag:"-"`
	StubsPath  string                 `json:"-" flag:"-"`
	PhpVersion *phpversion.PHPVersion `json:"-" flag:"-"`
}

type Php struct {
	Binary  string `json:"binary,omitempty"  default:"php" example:"valet php" doc:"The php binary used to execute external commands like analyzers."         usage:"The php binary used to execute external commands like analyzers."`
	Version string `json:"version,omitempty"               example:"8.1"       doc:"The PHP version to use when parsing, defaults to the output of 'php -v'." usage:"The PHP version to use when parsing, defaults to the output of 'php -v'." pattern:"^[7-8](\\.[0-9]+){0,2}$"`
}

type Phpcbf struct {
	Enabled  bool     `json:"enabled,omitempty"  default:"true"                     doc:"Enable formatting using PHPCBF."                                                                       usage:"Enable formatting using phpcbf."`
	Binary   []string `json:"binary,omitempty"   default:"vendor/bin/phpcbf,phpcbf" doc:"The paths checked, in order, for the PHPCBF binary."                                                   usage:"The paths checked, in order, for the PHPCBF binary."                                                   uniqueItems:"true" minItems:"1"`
	Standard string   `json:"standard,omitempty"                                    doc:"The PHPCS standard to format according to, NOTE: if this is set, the project level config is ignored." usage:"The PHPCS standard to format according to, NOTE: if this is set, the project level config is ignored."`
}

type Diagnostics struct {
	Enabled bool    `json:"enabled,omitempty" default:"true"`
	Phpstan Phpstan `json:"phpstan,omitempty"`
	Phpcs   Phpcs   `json:"phpcs,omitempty"`
}

type Phpstan struct {
	Analyzer
	Binary []string `json:"binary,omitempty" default:"vendor/bin/phpstan,phpstan" uniqueItems:"true" minItems:"1" example:"phpstan" doc:"The paths checked, in order, for the PHPStan binary." usage:"The paths checked, in order, for the PHPStan binary."`
}

type Phpcs struct {
	Analyzer
	Binary []string `json:"binary,omitempty" default:"vendor/bin/phpcs,phpcs" uniqueItems:"true" minItems:"1" example:"phpcs" doc:"The paths checked, in order, for the PHPCS binary." usage:"The paths checked, in order, for the PHPCS binary."`
}

type Analyzer struct {
	Method  DiagnosticsMethod `json:"method,omitempty"  default:"ON_CHANGE" enum:"ON_SAVE,ON_CHANGE" doc:"When to run diagnostics, either ON_SAVE or ON_CHANGE." usage:"When to run diagnostics, either ON_SAVE or ON_CHANGE."`
	Enabled bool              `json:"enabled,omitempty" default:"true"`
}

type Server struct {
	Communication connection.ConnType `json:"communication,omitempty" default:"stdio"          enum:"stdio,ws,tcp" doc:"How to communicate: standard io, web sockets or tcp."             usage:"How to communicate: standard io, web sockets or tcp."`
	URL           string              `json:"url,omitempty"           default:"127.0.0.1:2001"                     doc:"The URL to use for the websocket or tcp server."                  usage:"The URL to use for the websocket or tcp server."`
	ClientPid     uint                `json:"client_pid,omitempty"                                                 doc:"A process ID to watch for exits, server will exit when it exits." usage:"A process ID to watch for exits, server will exit when it exits." min:"1" flag:"client-pid"`
}

type Statsviz struct {
	Enabled bool   `json:"enabled,omitempty" default:"false"                         doc:"Visualize the server's memory usage, cpu usage, threads and other stats. NOTE: comes with a performance cost." usage:"Visualize the server's memory usage, cpu usage, threads and other stats. NOTE: comes with a performance cost."`
	URL     string `json:"url,omitempty"     default:"localhost:6060" doc:"Where to serve the visualizations."                                                                            usage:"Where to serve the visualizations."`
}
