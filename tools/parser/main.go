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
			fmt.Println(fmt.Errorf("parsing stdin: %w", err))
		}
	}

	if f != nil && len(*f) > 0 {
		file, err := os.Open(*f)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
			return
		}

		ast, err := p.Parse(*f, file)
		spew.Dump(ast)
		if err != nil {
			fmt.Println(err)
		}

		return
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Printf(">> ")
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		ast, err := p.ParseString("repl", line)
		spew.Dump(ast)
		if err != nil {
			fmt.Println(fmt.Errorf("parsing line from repl: %w", err))
		}
	}
}
