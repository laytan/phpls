// nolint: gocritic // hugeParam for dump not prio to address.
package ast

import "io"

var ComplexStrContentImps = []ComplexStrContent{
	BracedStringPart{},
	ComplexStringPart{},
	Expr{}, // The lexer already makes sure the expression is valid and delimited by string parts, which means that this is only ever a $foo[$xyz] or $foo->bar.
} // TODO: property access

// Anything that can be in a double quoted string, see `ast.ComplexStrContentImpls`.
type ComplexStrContent interface {
	Node
	complexStrContent()
}

type SimpleString struct {
	BaseNode

	Value string `parser:"@SimpleString"`
}

func (s SimpleString) value() {}

func (s SimpleString) Dump(w io.Writer, level int) {
	s.BaseDump(w, level, "SimpleString")
	level++
	writeStr(w, "value: %s", s.Value)
	writeEOL(w, level)
}

type ComplexString struct {
	BaseNode

	Content []ComplexStrContent `parser:"StringStart @@* StringEnd"`
}

func (c ComplexString) value() {}

func (c ComplexString) Dump(w io.Writer, level int) {
	c.BaseDump(w, level, "ComplexString")
	level++
	writeStr(w, "Content:")
	level++
	writeEOL(w, level)
	writeList(w, level, c.Content)
}

type ComplexStringPart struct {
	BaseNode

	Value string `parser:"@StringContent"`
}

func (c ComplexStringPart) complexStrContent() {}

func (c ComplexStringPart) Dump(w io.Writer, level int) {
	c.BaseDump(w, level, "ComplexStringPart")
	level++
	writeStr(w, "Value: %s", c.Value)
	writeEOL(w, level)
}

type BracedStringPart struct {
	BaseNode

	Value Expr `parser:"LBrace @@ RBrace"`
}

func (b BracedStringPart) complexStrContent() {}

func (b BracedStringPart) Dump(w io.Writer, level int) {
	b.BaseDump(w, level, "bracedStringPart")
	level++
	writeStr(w, "Value:")
	level++
	writeEOL(w, level)
	b.Value.Dump(w, level)
}
