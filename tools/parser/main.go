package main

// parser is a command that starts a "repl" which parser given input per line.

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/laytan/elephp/pkg/phparser/parser"
)

func main() {
	p := parser.Parser

	s := flag.String("input", "", "input to parse, instead of starting a repl")
	f := flag.String("file", "", "file to parse, instead of starting a repl")
	flag.Parse()
	if s != nil && len(*s) > 0 {
		ast, err := p.ParseString("stdin", *s)
		spew.Dump(ast)
		if err != nil {
			fprintln(fmt.Errorf("parsing stdin: %w", err))
		}
	}

	if f != nil && len(*f) > 0 {
		file, err := os.Open(*f)
		if err != nil {
			fprintln(err)
			os.Exit(1)
			return
		}

		ast, err := p.Parse(*f, file)
		spew.Dump(ast)
		if err != nil {
			fprintln(err)
		}

		return
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fprintf(">> ")
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		ast, err := p.ParseString("repl", line)
		spew.Dump(ast)
		if err != nil {
			fprintln(fmt.Errorf("parsing line from repl: %w", err))
		}
	}
}

func fprintln(value any) {
	if _, err := fmt.Println(value); err != nil {
		panic(err)
	}
}

func fprintf(value string, args ...any) {
	if _, err := fmt.Printf(value, args...); err != nil {
		panic(err)
	}
}
