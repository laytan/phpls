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

const (
	colorRed  = "\033[0;31m"
	colorGray = "\033[0;90m"
	colorNone = "\033[0m"
)

func main() {
	s := flag.String("input", "", "input to lex, instead of starting a repl")
	f := flag.String("file", "", "file to lex, instead of starting a repl")
	flag.Parse()
	if s != nil && len(*s) > 0 {
		l := lexer.New("stdin", strings.NewReader(*s))
		for t, _ := l.Next(); token.TokenType(t.Type) != token.EOF; t, _ = l.Next() {
			fmt.Printf("%+v\n", t)
		}
		return
	}

	if f != nil && len(*f) > 0 {
		file, err := os.Open(*f)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
			return
		}
		l := lexer.New(*f, file)
		line := 0
		fol := true
		for t, _ := l.Next(); token.TokenType(t.Type) != token.EOF; t, _ = l.Next() {
			if t.Pos.Line > line {
				line = t.Pos.Line
				fol = true
				fmt.Print("\n")
			}

			if fol {
				fol = false
				for i := 0; i < t.Pos.Column; i++ {
					fmt.Print(" ")
				}
			}

			tt := token.TokenType(t.Type)

			if tt == token.Illegal {
				fmt.Print(colorRed)
			}

			if tt == token.NonPHP || tt == token.BlockComment || tt == token.LineComment ||
				tt == token.PHPStart ||
				tt == token.PHPEchoStart ||
				tt == token.PHPEnd {
				fmt.Print(colorGray)
			}

			fmt.Printf("%s:\"%s\" ", token.TokenType(t.Type), t.Value)
			fmt.Print(colorNone)
		}
		fmt.Println()
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
		l := lexer.NewStartInPHP("repl", strings.NewReader(line))
		for t, _ := l.Next(); token.TokenType(t.Type) != token.EOF; t, _ = l.Next() {
			fmt.Printf("%#v\n", t)
		}
	}
}
