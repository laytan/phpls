package parser_test

import (
	"testing"

	"appliedgo.net/what"
	"github.com/laytan/elephp/pkg/phparser/parser"
	. "github.com/onsi/gomega"
)

func TestParser(t *testing.T) {
	g := NewWithT(t)

	ast, err := parser.Parser.ParseString("test", `
testing 123
<?php

namespace Foo;

public class Bar extends Baz implements Test, Test2 {
}
    `)

	g.Expect(err).ToNot(HaveOccurred())

	what.Is(ast)
}
