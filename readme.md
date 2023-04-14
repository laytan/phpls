# phpls

![Go build/test](https://github.com/laytan/phpls/actions/workflows/go-test.yml/badge.svg?branch=main)
![Go linting](https://github.com/laytan/phpls/actions/workflows/golangci-lint.yml/badge.svg?branch=main)

phpls is a fast and smart language server for PHP written in Go.

## Features

- Very fast indexing: Responsive right away, fully indexed in a matter of seconds
- Smart go to definition: [WIP](https://github.com/laytan/phpls/issues?q=is:issue+is:open+label:%22Go+To+Definition%22)
	- Understands type hints
	- Understands PHPDoc tags (generics to-do)
	- Go to definition inside PHPDoc, (like on the @param type or @inheritdoc)
- Context-aware smart completion: [WIP](https://github.com/laytan/phpls/labels/Completion)
	- Suggests everything available from the current scope
	- Automatically adds use statements
	- Can complete nesting `$this->config->get()->version`
- Proxy your favorite tools, create an issue if yours is missing:
	- Efficiently ran by keeping processes running
	- Configurable on-change or on-save
	- Configurable binaries, php version, standards
	- [PHPCS](https://github.com/squizlabs/PHP_CodeSniffer)
	- [PHPStan](https://phpstan.org/)
- Basic hover, on the to-do list to greatly improve

## Installation

More installation methods will be added soon.

### Building yourself

Make sure you have git and go installed.

```bash
git clone --recurse-submodules https://github.com/laytan/phpls.git
go build -o phpls cmd/phpls/main.go
```

Symlink the executable to a folder that's in your path, example:
```bash
sudo ln phpls /usr/local/bin/phpls
```

Or if you're on windows, add it to your path.

## Editor setup

### Neovim LSP Config

These configs have to be manually added for now, will merge into the LSP config repo soon.

```lua
require('lspconfig.configs').phpls = {
  default_config = {
	cmd = { 'phpls' },
	filetypes = { 'php' },
	root_dir = function(pattern)
	  local cwd = vim.loop.cwd()
	  local root = util.root_pattern('.git')(pattern)

	  return util.path.is_descendant(cwd, root) and cwd or root
	end,
  },
}

require('lspconfig').phpls.setup({})
```

### VSCode extension

Coming soon.

Running a language server is done using your IDE/editor most of the time,
but here are the options available:

## Configuration

The following file names are seen as phpls configuration files:
`phpls.json`, `.phpls.json`, `phpls.yml`, `.phpls.yml`, `phpls.yaml`, and `.phpls.yaml`

These files are recognized when they are in any of the following directories:

- Linux and Mac OS: `~/.config/`, `~/`, `$(pwd)/`
- Windows: `%AppData%/`, `$(pwd)/`

These files are checked top to bottom, with later files overwriting the former.

Configuration files are then overwritten by environment variables, with the prefix 'phpls_',
so setting the php version can for example be done with 'phpls_PHP_VERSION=8'.

You can also provide command line flags instead of a config file or environment variables.

See [this file](https://raw.githubusercontent.com/laytan/phpls/main/internal/config/phpls.schema.json) for the configuration schema.
If possible, soon, we will submit this schema to the schema repository so editors will use it for completion and documentation.

### Command line usage

```
Usage:
  -cache-path string
        Root directory for generated stubs and logs, defaults to the user cache directory.
  -config string
        config file param
  -diagnostics.enabled string
         (default "true")
  -diagnostics.phpcs.binary string
        The paths checked, in order, for the PHPCS binary. (default "vendor/bin/phpcs,phpcs")
  -diagnostics.phpcs.enabled string
         (default "true")
  -diagnostics.phpcs.method string
        When to run diagnostics, either ON_SAVE or ON_CHANGE. (default "ON_CHANGE")
  -diagnostics.phpstan.binary string
        The paths checked, in order, for the PHPStan binary. (default "vendor/bin/phpstan,phpstan")
  -diagnostics.phpstan.enabled string
         (default "true")
  -diagnostics.phpstan.method string
        When to run diagnostics, either ON_SAVE or ON_CHANGE. (default "ON_CHANGE")
  -dump-config string
        Dump the resolved config before validation, useful for debugging. (default "false")
  -extensions string
        File extensions to consider PHP code. (default ".php")
  -ignored-directories string
        Directories to ignore completely, use when you have huge directories with non-php files. (default ".git,node_modules")
  -php.binary string
        The php binary used to execute external commands like analyzers. (default "php")
  -php.version string
        The PHP version to use when parsing, defaults to the output of 'php -v'.
  -phpcbf.binary string
        The paths checked, in order, for the PHPCBF binary. (default "vendor/bin/phpcbf,phpcbf")
  -phpcbf.enabled string
        Enable formatting using phpcbf. (default "true")
  -phpcbf.standard string
        The PHPCS standard to format according to, NOTE: if this is set, the project level config is ignored.
  -server.client-pid string
        A process ID to watch for exits, server will exit when it exits.
  -server.communication string
        How to communicate: standard io, web sockets or tcp. (default "stdio")
  -server.url string
        The URL to use for the websocket or tcp server. (default "127.0.0.1:2001")
  -statsviz.enabled string
        Visualize the server's memory usage, cpu usage, threads and other stats. NOTE: comes with a performance cost. (default "false")
  -statsviz.url string
        Where to serve the visualizations. (default "localhost:6060")
```

### Default configuration

Here is the default configuration, your config will be merged into this one to create the final one.
See above command line usage for descriptions of what each configuration does.

```json
{
    "cache_path": "",
    "diagnostics": {
        "enabled": true,
        "phpcs": {
            "binary": [
                "vendor/bin/phpcs",
                "phpcs"
            ],
            "enabled": true,
            "method": "ON_CHANGE"
        },
        "phpstan": {
            "binary": [
                "vendor/bin/phpstan",
                "phpstan"
            ],
            "enabled": true,
            "method": "ON_CHANGE"
        }
    },
    "dump_config": false,
    "extensions": [
        ".php"
    ],
    "ignored_directories": [
        ".git",
        "node_modules"
    ],
    "php": {
        "binary": "php",
        "version": ""
    },
    "phpcbf": {
        "binary": [
            "vendor/bin/phpcbf",
            "phpcbf"
        ],
        "enabled": true,
        "standard": ""
    },
    "server": {
        "client_pid": 0,
        "communication": "stdio",
        "url": "127.0.0.1:2001"
    },
    "statsviz": {
        "enabled": false,
        "url": "localhost:6060"
    }
}
```

## Development

See Taskfile.yml for building, testing and other configuration, there is not much going on besides that.

## License

phpls is licensed with the [Apache 2](https://www.apache.org/licenses/LICENSE-2.0) license.

The project uses [JetBrains's phpstorm-stubs](https://github.com/JetBrains/phpstorm-stubs)
which is licensed with the [Apache 2](https://www.apache.org/licenses/LICENSE-2.0) license and does not make any significant changes to its licensed material.

[JetBrains's phpstorm-stubs](https://github.com/JetBrains/phpstorm-stubs) contains material by the PHP Documentation Group,
licensed with [CC-BY 3.0](https://www.php.net/manual/en/cc.license.php).
