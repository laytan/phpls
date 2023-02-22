package config

type opts struct {
	UseStdio        bool     `long:"stdio"           description:"Communicate over stdio"`
	UseWs           bool     `long:"ws"              description:"Communicate over websockets"`
	UseTCP          bool     `long:"tcp"             description:"Communicate over TCP"`
	Statsviz        bool     `long:"statsviz"        description:"Visualize stats(CPU, memory etc.) on localhost:6060/debug/statsviz"             short:"v"`
	ClientProcessID uint     `long:"clientProcessId" description:"Process ID that when terminated, terminates the language server"                short:"p"`
	URL             string   `long:"url"             description:"The URL to listen on for tcp or websocket connections"                                    default:"127.0.0.1:2001"`
	FileExtensions  []string `long:"fileExtensions"  description:"Define file extensions to treat as PHP source files"                            short:"e" default:"php"`
    IgnoredDirNames []string `long:"ignoredDirNames" description:"Define directory names that should be ignored completely"                       short:"i" default:".git" default:"node_modules"` // nolint:gofumpt // Go fmt removes the duplicate default.
    StubsDir        string   `long:"stubsDir"        description:"Where generated stubs should be stored, defaults to your OS's cache directory"`
    LogsDir         string   `long:"logsDir"         description:"Where logs should be stored, defaults to your Os's cache directory"`
    PHPVersion      string   `long:"php"             description:"The PHP version to parse and lint code with, defaults to output of php -v"`
}
