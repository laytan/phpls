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
        - [x] On $this
        - [ ] On arbitrary variables
            - [x] On variables created using new X()
            - [x] On variables with @var phpdoc
        - [ ] Static
        - [ ] Final methods
        - [ ] @method PhpDoc
        - [ ] parent::__construct
    - [ ] Properties
        - [x] On $this
        - [x] On variables created using new X()
        - [x] On variables with a @var phpdoc
    - [ ] Constants (classLike & global)
    - [ ] Anonymous functions
    - [ ] Arrow functions
    - [ ] Class names in PhpDoc
- [ ] Completion
    - [x] Class-like (interface, class & trait) names
    - [x] Function names
    - [ ] Variables names
    - [ ] Local variables/functions
    - [ ] Parameters
    - [ ] Methods
    - [ ] Properties
    - [ ] Constants
    - [ ] Namespaces
    - [ ] PHP Keywords/language constructs
    - [x] Automatic use statement on complete
    - [ ] **Incomplete files** (with syntax errors), challenging because the parser disregards lines with syntax errors
        - [x] Done in a hacky way, parsing out current word ourselves
        - [ ] Context aware (if $this-> for example, add all members on the current class)
    - [x] Autocomplete namespace usage

### Features to be implemented before v1

- Send parse errors to client
- Hover
- Completion
    - Show details about completion items (use hover content?)

### Features for later

- Check if we can publish the binary/ls via packagist
- Go to declaration
- Go to type definition
- Go to implementation
- Find references
- Document symbols
- Signature help
    - Workspace symbols
    - Watching for changes outside editor (Did change watched files)
- Diagnostics (Publish & Pull)
- Highlight references (Document highlight)
- Code lens
- Folding
- Code action
    - Rename
    - Align associative array keys
    - Create interface from class
        - get all public functions in the class and create an interface with them
        - interface name is the current class name + Interface
    - Override (asks for method to override, can lsp ask for input?)
        - can lsp ask for input
        - maybe add the trigger to the current class name and populate it with 'override x, override y' etc.
    - Update interface
        - when on a method that is implementing an interface (incorrectly)
        - update the interface signature to the current method signature

### Features with even lower priority/that don't make sense to me

**Look into what these are:**  
- Inline value (is this a refactor?)
- Inlay hints (Looks like a completion?)

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
make build
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

Run all tests:
```bash
make test
```

Linting:
```bash
# install at: https://golangci-lint.run/usage/install/
golangci-lint run
```

## Use cases for seperate PHP process

Some use cases have arrived for spawning a php process on startup that keeps 
listening for requests outlined below for advanced features (that NEED php packages).

I don't think this would slow the server down, because these will at most be sent 1 file
which should be very fast, also having it open (not restarting the php process)
everytime should make it fast enough.

### PHPDoc

We could try to use phpstan/phpdoc-parser
(spawn seperate process with (tcp/pipe?) interface to Go code listening for parse requests)
and with that, allow definition etc for tokens in the phpdoc.

**I think this is a nice to have, hover can just send the raw phpdoc and signature back,
for typing it might have value, but most modern code is using native type hints anyway.
It would be a cool thing to have and set us apart from other PHP language servers though.**

### Completion

For completion we might want a seperate process, just like phpdoc listening for requests and
parsing using microsoft/tolerant-php-parser. This should gracefully handle syntax errors and
still give result that is workable.

**We should first see if just taking the file content and the line that the user is editting,
then splitting/parsing out the identifier that is being typed is enough though.**

## License

Elephp is licensed with the [Apache 2](https://www.apache.org/licenses/LICENSE-2.0) license.

The project uses [JetBrains's phpstorm-stubs](https://github.com/JetBrains/phpstorm-stubs)
which is also licensed with the [Apache 2](https://www.apache.org/licenses/LICENSE-2.0) license and does not
make any significant changes to its licensed material.

[JetBrains's phpstorm-stubs](https://github.com/JetBrains/phpstorm-stubs) contains material by the PHP Documentation Group,
licensed with [CC-BY 3.0](https://www.php.net/manual/en/cc.license.php).

The project also uses a small subset of [VK.com's noverify](https://github.com/VKCOM/noverify)
licensed with the [MIT](https://raw.githubusercontent.com/VKCOM/noverify/v0.5.3/LICENSE) license.
