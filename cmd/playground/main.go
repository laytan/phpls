package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/laytan/elephp/pkg/phparser/parser"
	"github.com/laytan/elephp/pkg/phparser/token"
)

func main() {
	http.HandleFunc("/api/parse", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		a, err := parser.Parser.Parse("body", r.Body)
		d := float64(time.Since(start)) / float64(time.Millisecond)
		w.Header().Add("Server-Timing", fmt.Sprintf("parse;dur=%f", d))
		if err != nil {
			w.Write([]byte(err.Error()))
			w.Write([]byte("\n\n"))
		}

		// res := struct {
		// 	Error error
		// 	Ast   *ast.Program
		// }{err, a}
		//
		// ms, err := json.MarshalIndent(res, "", "  ")
		// if err != nil {
		// 	panic(err)
		// }
		//
		// w.Header().Add("Content-Type", "application/json")
		// w.Write(ms)
		spew.Fdump(w, a)
	})

	http.HandleFunc("/api/lex", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		toks, err := parser.Parser.Lex("body", r.Body)
		d := float64(time.Since(start)) / float64(time.Millisecond)
		w.Header().Add("Server-Timing", fmt.Sprintf("parse;dur=%f", d))
		if err != nil {
			w.Write([]byte(err.Error()))
			w.Write([]byte("\n\n"))
		}

		line := 0
		fol := true
		for _, t := range toks {
			if t.Pos.Line > line {
				line = t.Pos.Line
				fol = true
				fmt.Fprint(w, "\n")
			}

			if fol {
				fol = false
				for i := 0; i < t.Pos.Column; i++ {
					fmt.Fprint(w, " ")
				}
			}

			tt := token.TokenType(t.Type)
			fmt.Fprintf(w, "%s:\"%s\" ", tt, t.Value)
		}
	})

	http.Handle(
		"/",
		http.FileServer(http.Dir(filepath.Join(pathutils.Root(), "www", "playground"))),
	)
	log.Println("Listening on http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	log.Println(err)
}
