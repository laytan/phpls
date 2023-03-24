package lexer_test

import (
	"strings"
	"testing"

	"appliedgo.net/what"
	plexer "github.com/alecthomas/participle/v2/lexer"
	"github.com/laytan/elephp/pkg/phparser/lexer"
	"github.com/laytan/elephp/pkg/phparser/token"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

func TestMain(m *testing.M) {
	format.UseStringerRepresentation = true
}

func TestNextToken(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	input := `$five = 5;
$ten = 0xa;

function add($x, $y) {
    return $x + $y;
}

$result = add($five, $ten); // Hello, world!

$a = 'test$test'
'test\''
'test\\'
'test\\\''
"test"
"$test"
"$"
"{"
"}"
"{}"
"hello$world"
"hello{$world}!"
"hello {$test"
yield from
yield frog
frogyield from
`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
		expectedLine    int
		expectedCol     int
	}{
		{token.Var, "$five", 0, 0},
		{token.Assign, "=", 0, 6},
		{token.Number, "5", 0, 8},
		{token.Semicolon, ";", 0, 9},

		{token.Var, "$ten", 1, 0},
		{token.Assign, "=", 1, 5},
		{token.Number, "0xa", 1, 7},
		{token.Semicolon, ";", 1, 10},

		{token.Function, "function", 3, 0},
		{token.Ident, "add", 3, 9},
		{token.LParen, "(", 3, 12},
		{token.Var, "$x", 3, 13},
		{token.Comma, ",", 3, 15},
		{token.Var, "$y", 3, 17},
		{token.RParen, ")", 3, 19},
		{token.LBrace, "{", 3, 21},

		{token.Return, "return", 4, 4},
		{token.Var, "$x", 4, 11},
		{token.Plus, "+", 4, 14},
		{token.Var, "$y", 4, 16},
		{token.Semicolon, ";", 4, 18},

		{token.RBrace, "}", 5, 0},

		{token.Var, "$result", 7, 0},
		{token.Assign, "=", 7, 8},
		{token.Ident, "add", 7, 10},
		{token.LParen, "(", 7, 13},
		{token.Var, "$five", 7, 14},
		{token.Comma, ",", 7, 19},
		{token.Var, "$ten", 7, 21},
		{token.RParen, ")", 7, 25},
		{token.Semicolon, ";", 7, 26},
		{token.LineComment, "// Hello, world!", 7, 28},

		{token.Var, "$a", 9, 0},
		{token.Assign, "=", 9, 3},
		{token.SimpleString, "'test$test'", 9, 5},

		{token.SimpleString, `'test\''`, 10, 0},

		{token.SimpleString, `'test\\'`, 11, 0},

		{token.SimpleString, `'test\\\''`, 12, 0},

		{token.StringStart, `"`, 13, 0},
		{token.StringContent, "test", 13, 1},
		{token.StringEnd, `"`, 13, 5},

		{token.StringStart, `"`, 14, 0},
		{token.Var, `$test`, 14, 1},
		{token.StringEnd, `"`, 14, 6},

		{token.StringStart, `"`, 15, 0},
		{token.StringContent, `$`, 15, 1},
		{token.StringEnd, `"`, 15, 2},

		{token.StringStart, `"`, 16, 0},
		{token.StringContent, `{`, 16, 1},
		{token.StringEnd, `"`, 16, 2},

		{token.StringStart, `"`, 17, 0},
		{token.StringContent, `}`, 17, 1},
		{token.StringEnd, `"`, 17, 2},

		{token.StringStart, `"`, 18, 0},
		{token.StringContent, `{}`, 18, 1},
		{token.StringEnd, `"`, 18, 3},

		// "hello$world"
		{token.StringStart, `"`, 19, 0},
		{token.StringContent, "hello", 19, 1},
		{token.Var, "$world", 19, 6},
		{token.StringEnd, `"`, 19, 12},

		// "hello{$world}!"
		{token.StringStart, `"`, 20, 0},
		{token.StringContent, "hello", 20, 1},
		{token.LBrace, "{", 20, 6},
		{token.Var, "$world", 20, 7},
		{token.RBrace, "}", 20, 13},
		{token.StringContent, "!", 20, 14},
		{token.StringEnd, `"`, 20, 15},

		// "hello {$test"
		{token.StringStart, `"`, 21, 0},
		{token.StringContent, "hello {", 21, 1},
		{token.Var, "$test", 21, 8},
		{token.StringEnd, `"`, 21, 13},

		// yield from
		{token.YieldFrom, "yield from", 22, 0},

		// yield frog
		{token.Yield, "yield", 23, 0},
		{token.Ident, "frog", 23, 6},

		// frogyield from
		{token.Ident, "frogyield", 24, 0},
		{token.Ident, "from", 24, 10},
	}

	l := lexer.NewStartInPHP("test", strings.NewReader(input))

	for _, tt := range tests {
		tok, _ := l.Next()

		what.Is(tok)
		what.Is(tt)

		g.Expect(tok.Type).To(Equal(plexer.TokenType(tt.expectedType)))
		g.Expect(tok.Value).To(Equal(tt.expectedLiteral))
		g.Expect(tok.Pos.Column).To(Equal(tt.expectedCol), "columns should be equal")
		g.Expect(tok.Pos.Line).To(Equal(tt.expectedLine), "rows should be equal")
	}
}

func TestStringVars(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tests := []struct {
		input  string
		output string
	}{
		{`"test"`, `"test"`},
		{`"$test"`, `""`},
		{`"$test."`, `"."`},
		{`"$test[0]"`, `""`},
		{`"teehee$test[0]"`, `"teehee"`},
		{`"$people->john then said hello to $people->jane."`, `" then said hello to ."`},
		{`"The ch at -2 is $string[-2]!"`, `"The ch at -2 is !"`},
		{`"He drank some $juices[koolaid1] juice."`, `"He drank some juice."`},
		{`"$test[0]->test"`, `"->test"`},
		{`"{$test[0]->test}"`, `""`},
		{`"test{$test}test"`, `"testtest"`},
		{`"This is { $great}"`, `"This is { }"`},
		{`"This is {$arr[4][3]}"`, `"This is"`},
		{`"$bar['foo']test"`, `"test"`},
		{`"hello{$world}!"`, `"hello!"`},
	}

	for _, tt := range tests {
		l := lexer.NewStartInPHP("test", strings.NewReader(tt.input))

		actOut := ""
		for tok, _ := l.Next(); tok.Type != plexer.TokenType(token.EOF); tok, _ = l.Next() {
			if tok.Type == plexer.TokenType(token.StringStart) ||
				tok.Type == plexer.TokenType(token.StringEnd) ||
				tok.Type == plexer.TokenType(token.StringContent) {
				actOut += tok.Value
			}
		}

		g.Expect(actOut).To(Equal(actOut))
	}
}
