// nolint: gocritic // hugeParam for dump not prio to address.
package ast

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/laytan/elephp/pkg/phparser/token"
)

type Node interface {
	Position() lexer.Position
	Dump(w io.Writer, level int)
}

type BaseNode struct {
	Pos    lexer.Position
    Tokens []lexer.Token // TODO: when done, this can prob go for more performance.
}

func (b BaseNode) Position() lexer.Position {
	return lexer.Position{}
}

func (b BaseNode) BaseDump(w io.Writer, level int, name string) {
	writeStr(w, "%s:", name)
	level++
	writeEOL(w, level)
	writeStr(w, "%d Tokens (%d-%d): ", len(b.Tokens), b.Pos.Line, b.Pos.Column)
	for _, t := range b.Tokens {
		writeStr(w, "%s ", token.Type(t.Type))
	}
	writeEOL(w, level)
}

type Program struct {
	BaseNode

	Statements []Statement `parser:"@@*"`
}

var _ Node = Program{}

func (p Program) Dump(w io.Writer, level int) {
	p.BaseDump(w, level, "Program")
	level++
	writeStr(w, "Statements:")
	level++
	writeEOL(w, level)
	writeList(w, level, p.Statements)
}

type Extends struct {
	BaseNode

	Name Name `parser:"Extends @@"`
}

var _ Node = Extends{}

func (e Extends) Dump(w io.Writer, level int) {
	e.BaseDump(w, level, "Extends")
	level++
	writeStr(w, "Name:")
	level++
	writeEOL(w, level)
	e.Name.Dump(w, level)
}

type Implements struct {
	BaseNode

	Names []Name `parser:"Implements @@ ( Comma @@ )*"`
}

var _ Node = Implements{}

func (i Implements) Dump(w io.Writer, level int) {
	i.BaseDump(w, level, "Implements")
	level++
	writeStr(w, "Names:")
	level++
	writeEOL(w, level)
	writeList(w, level, i.Names)
}

type Name struct {
	BaseNode

	Value string `parser:"@Ident"`
}

var _ Node = Name{}

func (n Name) Dump(w io.Writer, level int) {
	n.BaseDump(w, level, "Name")
	level++
	writeStr(w, "Value: %s", n.Value)
	writeEOL(w, level)
}

func writeStr(w io.Writer, val string, format ...any) {
	if _, err := fmt.Fprintf(w, val, format...); err != nil {
		log.Println(fmt.Errorf("writing \"%s\" to ast dump writer: %w", val, err))
	}
}

func writeEOL(w io.Writer, level int) {
	writeStr(w, "\n"+strings.Repeat("    ", level))
}

func writeList[T Node](w io.Writer, level int, list []T) {
	for i, e := range list {
		writeStr(w, fmt.Sprintf("%d) ", i+1))
		e.Dump(w, level)
		writeEOL(w, level)
	}
}

func writeStrList(w io.Writer, level int, list []string) {
	for i, e := range list {
		writeStr(w, fmt.Sprintf("%d) %s", i+1, e))
		writeEOL(w, level)
	}
}

func defaultStr(str *string) string {
	if str == nil {
		return ""
	}

	return *str
}
