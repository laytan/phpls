package config

type opts struct {
	UseStdio        bool     `long:"stdio"           description:"Communicate over stdio"`
	UseWs           bool     `long:"ws"              description:"Communicate over websockets"`
	UseTCP          bool     `long:"tcp"             description:"Communicate over TCP"`
	Statsviz        bool     `long:"statsviz"        description:"Visualize stats(CPU, memory etc.) on localhost:6060/debug/statsviz" short:"v"`
	ClientProcessID uint     `long:"clientProcessId" description:"Process ID that when terminated, terminates the language server"    short:"p"`
	URL             string   `long:"url"             description:"The URL to listen on for tcp or websocket connections"                        default:"127.0.0.1:2001"`
	FileExtensions  []string `long:"fileExtensions"  description:"Define file extensions to treat as PHP source files"                short:"e" default:"php"`
}
