# phpls

![Go build/test](https://github.com/laytan/phpls/actions/workflows/go-test.yml/badge.svg?branch=main)
![Go linting](https://github.com/laytan/phpls/actions/workflows/golangci-lint.yml/badge.svg?branch=main)

phpls is a language server for PHP.

## Features

- Go To Definition, mostly done, see [this milestone](https://github.com/laytan/phpls/milestone/1)
- Hover, basic implementation, shows the PHPDoc and signature for the symbol under cursor that can be defined using Go To Definition
- Completion, basic implementation, completes global functions, constants, classes, interfaces and traits, and auto-inserts a use-statement, see [this issue](https://github.com/laytan/phpls/issues/22)
- Diagnostics, basic diagnostics for syntax/parse errors, use PHPStan, PHPCS etc. if more diagnostics are wanted

## Installation

### Pre-built

1. Download the zip file from the releases page
2. Unzip this somewhere
3. Symlink the executable to a folder that's in your path, example:
```bash
sudo ln -s /unzipped/folder/phpls /usr/local/bin/phpls
```

### Building yourself

Make sure you have git and go(1.18) installed.

```bash
git clone --recurse-submodules https://github.com/laytan/phpls.git
make build
```

Symlink the executable to a folder that's in your path, example:
```bash
sudo ln -s /unzipped/folder/phpls /usr/local/bin/phpls
```

Or if you're on windows, add it to your path.

## Running the language server

Running a language server is done using your IDE/editor most of the time,
but here are the options available:

```bash
phpls -h

#Usage:
#  phpls [OPTIONS]
#
#Application Options:
#      --stdio            Communicate over stdio
#      --ws               Communicate over websockets
#      --tcp              Communicate over TCP
#  -v, --statsviz         Visualize stats(CPU, memory etc.) on localhost:6060/debug/statsviz
#  -p, --clientProcessId= Process ID that when terminated, terminates the language server
#      --url=             The URL to listen on for tcp or websocket connections (default: 127.0.0.1:2001)
#  -e, --fileExtensions=  Define file extensions to treat as PHP source files (default: php)
#  -i, --ignoredDirNames= Define directory names that should be ignored completely (default: .git, node_modules)
#      --stubsDir=        Where generated stubs should be stored, defaults to your OS's cache directory
#      --php=             The PHP version to parse and lint code with, defaults to output of php -v
#
#Help Options:
#  -h, --help             Show this help message
```

## Development

Run all tests:
```bash
make test
```

Linting:
```bash
# install at: https://golangci-lint.run/usage/install/
golangci-lint run
```

## License

phpls is licensed with the [Apache 2](https://www.apache.org/licenses/LICENSE-2.0) license.

The project uses [JetBrains's phpstorm-stubs](https://github.com/JetBrains/phpstorm-stubs)
which is also licensed with the [Apache 2](https://www.apache.org/licenses/LICENSE-2.0) license and does not
make any significant changes to its licensed material.

[JetBrains's phpstorm-stubs](https://github.com/JetBrains/phpstorm-stubs) contains material by the PHP Documentation Group,
licensed with [CC-BY 3.0](https://www.php.net/manual/en/cc.license.php).

The project also uses a small subset of [VK.com's noverify](https://github.com/VKCOM/noverify)
licensed with the [MIT](https://raw.githubusercontent.com/VKCOM/noverify/v0.5.3/LICENSE) license.
