# Elephp

![Go build/test](https://github.com/laytan/elephp/actions/workflows/go-test.yml/badge.svg?branch=main)
![Go linting](https://github.com/laytan/elephp/actions/workflows/golangci-lint.yml/badge.svg?branch=main)

Elephp is a language server for PHP.

## Features

Below I have outlined all features that a LSP can theoretically do, and the progress/priority for it.

### In progress/done

- [x] Communication
    - [x] Stdio
    - [x] Websockets
    - [x] TCP
- [ ] Go to definition
    - [x] Standard PHP symbols
    - [x] Use statements
    - [x] Namespace statements
    - [x] ClassLike (classes, interfaces & traits)
    - [x] Extends statements
    - [x] Global variables
    - [x] Local variables
    - [x] Parameters
    - [ ] Methods
    - [ ] Properties
    - [ ] Constants (classLike & global)
    - [ ] Anonymous functions
    - [ ] Arrow functions

### Features to be implemented before v1

- Check if we can publish the binary/ls via packagist
- Go to declaration
- Go to type definition
- Go to implementation
- Find references
- Hover
- Document symbols
- Signature help
- Completion
- Workspace symbols
- Watching for changes outside editor (Did change watched files)

### Features for later

- Diagnostics (Publish & Pull)
- Highlight references (Document highlight)
- Rename
- Code lens
- Folding

### Features with even lower priority/that don't make sense to me

**Look into what these are:**  
- Inline value (is this a refactor?)
- Inlay hints (Looks like a completion?)
- Code action

**Others:**  
- Selection range
- Document link
- Semantic tokens
- Moniker
- Document Colors
- Formatting (Use third party formatter?)
    - Full formatting
    - Range formatting
    - On type formatting
- Linked editing range

## Installation

### Prebuilt

1. Download the zip file from the releases page
2. Unzip this somewhere
3. Symlink the executable to a folder that's in your path, example:
```bash
sudo ln -s /unzipped/folder/elephp /usr/local/bin/elephp
```

### Building yourself

Make sure you have git and go(1.18) installed.

```bash
git clone --recurse-submodules https://github.com/laytan/elephp.git
go build -o elephp cmd/main.go
```

Symlink the executable to a folder that's in your path, example:
```bash
sudo ln -s /unzipped/folder/elephp /usr/local/bin/elephp
```

Or if you're on windows, add it to your path.

## Running the language server

Running a language server is done using your IDE/editor most of the time,
but here are the options available:

```
elephp -h

Application Options:
      --clientProcessId=             Process ID that when terminated, terminates the language server
      --stdio                        Communicate over stdio
      --ws                           Communicate over websockets
      --tcp                          Communicate over TCP
      --url=                         The URL to listen on for tcp or websocket connections (default: 127.0.0.1:2001)
      --statsviz                     Visualize stats(CPU, memory etc.) on localhost:6060/debug/statsviz
      --log=[stderr|file]            Set the log output location (default: file)
      --level=[info|info|warn|error] The level of logs to output (default: info)

Help Options:
  -h, --help                         Show this help message
```

## Logs

If --log is set to file(the default) when running the language server, logs will be put in
the logs/elephp.log file.

Once this file gets to 100mb, the file is renamed to elephp-timestamp.log and the 
main log file gets recreated.
The maximum amount of log files is set to 2, so old log files will get deleted automatically.

You can keep an eye on the logs by running `elephp logs`, this will tail the log file for you.

## Development

Run all tests with coverage:
```bash
go test ./... --cover -timeout=1s
```

Linting:
```bash
# install at: https://golangci-lint.run/usage/install/
golangci-lint run
```

## License

Elephp is licensed with the [Apache 2](https://www.apache.org/licenses/LICENSE-2.0) license.

The project uses [JetBrains's phpstorm-stubs](https://github.com/JetBrains/phpstorm-stubs)
which is also licensed with the [Apache 2](https://www.apache.org/licenses/LICENSE-2.0) license and does not
make any significant changes to its licensed material.

[JetBrains's phpstorm-stubs](https://github.com/JetBrains/phpstorm-stubs) contains material by the PHP Documentation Group,
licensed with [CC-BY 3.0](https://www.php.net/manual/en/cc.license.php).

The project also uses a small subset of [VK.com's noverify](https://github.com/VKCOM/noverify)
licensed with the [MIT](https://raw.githubusercontent.com/VKCOM/noverify/v0.5.3/LICENSE) license.
