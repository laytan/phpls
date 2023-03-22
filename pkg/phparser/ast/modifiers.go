package ast

import (
	"fmt"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/laytan/elephp/pkg/phparser/token"
	"github.com/laytan/elephp/pkg/phprivacy"
)

type Modifiers struct {
	Visibility *string
	Others     []string
}

var _ participle.Parseable = &Modifiers{}

func isVisibilityModifier(t token.TokenType) bool {
	return t == token.Public || t == token.Private || t == token.Protected
}

func isOtherModifier(t token.TokenType) bool {
	return t == token.Readonly || t == token.Abstract || t == token.Static || t == token.Final
}

func isModifier(t token.TokenType) bool {
	return isVisibilityModifier(t) || isOtherModifier(t)
}

// TODO: classes and functions have different rules.
func (m *Modifiers) Parse(lex *lexer.PeekingLexer) error {
	tok := lex.Peek()
	typ := token.TokenType(tok.Type)
	for isModifier(typ) {
		if isVisibilityModifier(typ) {
			if m.Visibility != nil {
				return &participle.ParseError{
					Msg: fmt.Sprintf(
						"multiple visibility modifiers are not allowed, found \"%s\" but already got \"%s\"",
						tok.Value,
						*m.Visibility,
					),
					Pos: tok.Pos,
				}
			}

			v := lex.Next()
			m.Visibility = &v.Value
		} else if isOtherModifier(typ) {
			for _, other := range m.Others {
				if other == tok.Value {
					return &participle.ParseError{
						Msg: fmt.Sprintf("duplicate modifiers are not allowed, found \"%s\" twice", tok.Value),
						Pos: tok.Pos,
					}
				} else if typ == token.Final && other == "abstract" {
					return &participle.ParseError{
						Msg: "cannot mark as final and abstract",
						Pos: tok.Pos,
					}
				} else if typ == token.Abstract && other == "final" {
					return &participle.ParseError{
						Msg: "cannot mark as abstract and final",
						Pos: tok.Pos,
					}
				}
			}

			m.Others = append(m.Others, lex.Next().Value)
		}

		tok = lex.Peek()
		typ = token.TokenType(tok.Type)
	}

	return nil
}

func (v *Modifiers) Privacy() phprivacy.Privacy {
	if v.Visibility == nil {
		return phprivacy.PrivacyPublic
	}

    p, err := phprivacy.FromString(*v.Visibility)
    if err != nil {
        panic(fmt.Errorf("should not happen: parsing modifiers visibility: %w", err))
    }
    return p
}
