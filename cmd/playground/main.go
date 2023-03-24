package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/laytan/elephp/pkg/phparser/parser"
	"github.com/laytan/elephp/pkg/phparser/token"
)

func main() {
	http.HandleFunc("/api/parse", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		a, err := parser.Parser.Parse("body", r.Body)
		d := float64(time.Since(start)) / float64(time.Millisecond)
		w.Header().Add("Server-Timing", fmt.Sprintf("Parse;dur=%f", d))
		if err != nil {
			writeStr(w, err.Error())
			writeStr(w, "\n\n")
		}

		a.Dump(w, 0)
	})

	http.HandleFunc("/api/lex", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		toks, err := parser.Parser.Lex("body", r.Body)
		d := float64(time.Since(start)) / float64(time.Millisecond)
		w.Header().Add("Server-Timing", fmt.Sprintf("Lex;dur=%f", d))
		if err != nil {
			writeStr(w, err.Error())
			writeStr(w, "\n\n")
		}

		line := 0
		fol := true
		for _, t := range toks {
			if t.Pos.Line > line {
				line = t.Pos.Line
				fol = true
				fprintf(w, "\n %d | ", line+1)
			}

			if fol {
				fol = false
				for i := 0; i < t.Pos.Column; i++ {
					fprintf(w, " ")
				}
			}

			tt := token.Type(t.Type)
			fprintf(w, "%s:\"%s\" ", tt, t.Value)
		}
	})

	http.Handle(
		"/",
		http.FileServer(http.Dir(filepath.Join(pathutils.Root(), "web", "playground"))),
	)
	log.Println("Listening on http://localhost:8080")
	err := http.ListenAndServe( // nolint:gosec // This is a local server, no security needed.
		":8080",
		nil,
	)
	log.Println(err)
}

func writeStr(w io.Writer, value string) {
	if _, err := w.Write([]byte(value)); err != nil {
		panic(err)
	}
}

func fprintf(w io.Writer, format string, args ...any) {
	if _, err := fmt.Fprintf(w, format, args...); err != nil {
		panic(err)
	}
}
