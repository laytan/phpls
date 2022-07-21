# Elephp

![Go build/test](https://github.com/laytan/elephp/actions/workflows/go-test.yml/badge.svg?branch=main)
![Go linting](https://github.com/laytan/elephp/actions/workflows/golangci-lint.yml/badge.svg?branch=main)

Elephp is a language server for PHP.

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
git clone https://github.com/laytan/elephp.git
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
      --clientProcessId=              Process ID that when terminated, terminates the
                                      language server
      --stdio                         Communicate over stdio
      --ws                            Communicate over websockets
      --tcp                           Communicate over TCP
      --url=                          The URL to listen on for tcp or websocket
                                      connections (default: 127.0.0.1:2001)
      --statsviz                      Visualize stats(CPU, memory etc.) on
                                      localhost:6060/debug/statsviz
      --log=[stderr|file]             Set the log output location (default: stderr)
      --level=[debug|info|warn|error] The level of logs to output (default: warn)

Help Options:
  -h, --help                          Show this help message
```

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
