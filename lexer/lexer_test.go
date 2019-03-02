package lexer

import (
	"github.com/drejca/shiftlang/token"
	"strings"
	"testing"
)

func TestReadRune(t *testing.T) {
	input := `
fn Add(a i32, b i32) : i32 {
	return a + b;
}`
	lex := New(strings.NewReader(input))

	for _, ch := range input {
		rune := lex.read()
		if rune != ch {
			t.Errorf("expected to Read rune %q got %q", ch, rune)
		}
	}
}

func TestNextToken(t *testing.T) {
	input := `
fn Add(a i32, b i32) : i32 {
	return a + b;
}~`

	tests := []struct {
		tokenType token.Type
		literal string
	} {
		{tokenType: token.FUNC, literal: "fn"},
		{tokenType: token.IDENT, literal: "Add"},
		{tokenType: token.LPAREN, literal: "("},
		{tokenType: token.IDENT, literal: "a"},
		{tokenType: token.IDENT, literal: "i32"},
		{tokenType: token.COLON, literal: ","},
		{tokenType: token.IDENT, literal: "b"},
		{tokenType: token.IDENT, literal: "i32"},
		{tokenType: token.RPAREN, literal: ")"},
		{tokenType: token.COLON, literal: ":"},
		{tokenType: token.IDENT, literal: "i32"},
		{tokenType: token.LCURLY, literal: "{"},
		{tokenType: token.RETURN, literal: "return"},
		{tokenType: token.IDENT, literal: "a"},
		{tokenType: token.PLUS, literal: "+"},
		{tokenType: token.IDENT, literal: "b"},
		{tokenType: token.SEMICOLON, literal: ";"},
		{tokenType: token.RCURLY, literal: "}"},
		{tokenType: token.ILLEGAL, literal: "~"},
		{tokenType: token.EOF, literal: string(rune(token.EOF))},
	}

	lex := New(strings.NewReader(input))

	for i, test := range tests {
		tok := lex.NextToken()

		if tok.Type != test.tokenType {
			t.Errorf("tests[%d] - wrong type. Expected %q but got %q", i, token.Print(test.tokenType), token.Print(tok.Type))
		}

		if tok.Lit != test.literal {
			t.Errorf("tests[%d] - wrong literal. Expected %q but got %q", i, test.literal, tok.Lit)
		}
	}
}

func TestIsLetter(t *testing.T) {
	tests := []struct {
		ch rune
		isLetter bool
	} {
		{ch: 'a', isLetter: true},
		{ch: 'b', isLetter: true},
		{ch: '@', isLetter: false},
		{ch: 'z', isLetter: true},
		{ch: 'A', isLetter: true},
		{ch: 'Z', isLetter: true},
		{ch: '[', isLetter: false},
	}

	for i, test := range tests {
		if isLetter(test.ch) != test.isLetter {
			t.Errorf("tests[%d] - expected isLetter for ch(%q) to return %t got %t", i, test.ch, test.isLetter, isLetter(test.ch))
		}
	}
}
