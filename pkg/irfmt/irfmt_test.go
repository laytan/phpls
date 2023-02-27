package irfmt_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/laytan/elephp/pkg/irfmt"
	"github.com/laytan/elephp/pkg/parsing"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/matryer/is"
)

func TestFmt(t *testing.T) {
	is := is.New(t)

	in := `
<?php
/**
 * Test
 */
#[Test]
function test() {
    // Test.
    /** @var string */
    $test = '';
}

class Test {

    /**
     * @return string
     */
     public function test(): string|int|false {}

}
    `

	out := `
<?php
/**
 * Test
 */
#[Test]
function test() {
    // Test.
    /** @var string */
    $test = '';
}

class Test {

    /**
     * @return string
     */
    public function test(): string|int|false {}

}
    `

	parser := parsing.New(phpversion.EightOne())
	ir, err := parser.Parse(in)
	is.NoErr(err)

	o := bytes.NewBufferString("")
	irfmt.NewPrettyPrinter(o, "    ").Print(ir)

	aOut := strings.TrimSpace(out)
	aO := strings.TrimSpace(o.String())

	if aO != aOut {
		t.Errorf(
			"Result not as expected:\nWant: %v\nGot : %v\nDiff: %v",
			aOut,
			aO,
			diff.CharacterDiff(aOut, aO),
		)
	}
}
