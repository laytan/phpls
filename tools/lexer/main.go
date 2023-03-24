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
		for t, _ := l.Next(); token.Type(t.Type) != token.EOF; t, _ = l.Next() {
			fprintf("%+v\n", t)
		}
		return
	}

	if f != nil && len(*f) > 0 {
		file, err := os.Open(*f)
		if err != nil {
			fprintln(err)
			os.Exit(1)
			return
		}
		l := lexer.New(*f, file)
		line := 0
		fol := true
		for t, _ := l.Next(); token.Type(t.Type) != token.EOF; t, _ = l.Next() {
			if t.Pos.Line > line {
				line = t.Pos.Line
				fol = true
				fprintln("")
			}

			if fol {
				fol = false
				for i := 0; i < t.Pos.Column; i++ {
					fprintf(" ")
				}
			}

			tt := token.Type(t.Type)

			if tt == token.Illegal {
				fprintf(colorRed)
			}

			if tt == token.NonPHP || tt == token.BlockComment || tt == token.LineComment ||
				tt == token.PHPStart ||
				tt == token.PHPEchoStart ||
				tt == token.PHPEnd {
				fprintf(colorGray)
			}

			fprintf("%s:\"%s\" ", token.Type(t.Type), t.Value)
			fprintf(colorNone)
		}

		fprintln("")
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
		l := lexer.NewStartInPHP("repl", strings.NewReader(line))
		for t, _ := l.Next(); token.Type(t.Type) != token.EOF; t, _ = l.Next() {
			fprintf("%#v\n", t)
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
