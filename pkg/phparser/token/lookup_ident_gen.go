//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/laytan/elephp/pkg/phparser/token"
)

var fileTempl = template.Must(template.New("").Parse(`// Code generated by go generate; DO NOT EDIT.

// Add a token between tokenType.KeywordsStart and tokenType.KeywordsEnd and it will automatically be added.
// The keyword match is the lowercased version of the constant,
// You can change this keyword by adding a comment like this: Foo // literal:"bar".

package token

func LookupIdent(keyword string) TokenType {
    switch(keyword) { {{- range $case := .Cases}}
    case "{{$case.Keyword}}":
        return {{$case.TokenType}}{{end}}
    default:
        return Ident
    }
}

`))

type Case struct {
	Keyword   string
	TokenType string
}

type TemplData struct {
	Cases []Case
}

func main() {
	f, err := os.Create("lookup_ident.go")
	if err != nil {
		panic(fmt.Errorf("could not create lookup_ident.go: %w", err))
	}

	defer f.Close()

	source, err := ioutil.ReadFile("token.go")
	if err != nil {
		panic(fmt.Errorf("reading token.go: %w", err))
	}

	cases := []Case{}
	for i := token.KeywordsStart + 1; i < token.KeywordsEnd; i++ {
		cases = append(cases, Case{
			Keyword:   tokenTypeToKeyword(string(source), i),
			TokenType: i.String(),
		})
	}

	fileTempl.Execute(f, TemplData{cases})
}

func tokenTypeToKeyword(source string, t token.TokenType) string {
	// Match the token type followed by a space, followed by // literal:"foobar"
	rg := regexp.MustCompile(fmt.Sprintf(`\s%s \/\/ literal:"(.*)"`, t.String()))
	res := rg.FindStringSubmatch(source)
	if len(res) < 2 {
		// Otherwise, just lowercase it.
		keyword := strings.ToLower(t.String())
		// fmt.Printf("type '%s' matched by '%s'\n", t, keyword)
		return keyword
	}

	// fmt.Printf("type '%s' matched by '%s'\n", t, res[1])

	return res[1]
}
