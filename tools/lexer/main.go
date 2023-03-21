package main

// lexer is a command that starts a "repl" which lexes given input per line.

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/laytan/elephp/pkg/phparser/lexer"
	"github.com/laytan/elephp/pkg/phparser/token"
)

func main() {
	s := flag.String("input", "", "input to lex, instead of starting a repl")
	flag.Parse()
	if s != nil && len(*s) > 0 {
		l := lexer.New(strings.NewReader(*s))
		for t := l.Next(); t.Type != token.EOF; t = l.Next() {
			fmt.Printf("%+v\n", t)
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
		l := lexer.NewStartInPHP(strings.NewReader(line))
		for t := l.Next(); t.Type != token.EOF; t = l.Next() {
			fmt.Printf("%+v\n", t)
		}
	}
}
