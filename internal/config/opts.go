package config

type opts struct {
	ClientProcessID uint      `long:"clientProcessId" description:"Process ID that when terminated, terminates the language server"`
	UseStdio        bool      `long:"stdio"           description:"Communicate over stdio"`
	UseWs           bool      `long:"ws"              description:"Communicate over websockets"`
	UseTCP          bool      `long:"tcp"             description:"Communicate over TCP"`
	URL             string    `long:"url"             description:"The URL to listen on for tcp or websocket connections"              default:"127.0.0.1:2001"`
	Statsviz        bool      `long:"statsviz"        description:"Visualize stats(CPU, memory etc.) on localhost:6060/debug/statsviz"`
	// NOTE: Go fmt removes duplicate struct tags which breaks this, be careful.
        Log             LogOutput `long:"log"             description:"Set the log output location"                                        default:"file"           choice:"stderr" choice:"file"`
        LogLevel        LogLevel  `long:"level"           description:"The level of logs to output"                                        default:"info"           choice:"info"   choice:"info" choice:"warn" choice:"error"`
}

