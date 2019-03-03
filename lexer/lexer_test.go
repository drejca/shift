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
}~
2 - 1;
`

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
		{tokenType: token.INT, literal: "2"},
		{tokenType: token.MINUS, literal: "-"},
		{tokenType: token.INT, literal: "1"},
		{tokenType: token.SEMICOLON, literal: ";"},
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

func TestTokenPosition(t *testing.T) {
	input := `
fn Sub(a i32, b i32) : i32 {
	return a - b;
}
`
	tests := []struct{
		tokenType token.Type
		literal string
		pos token.Position
	}{
		{tokenType: token.FUNC, literal: "fn", pos: token.Position{Line: 2, Column: 1}},
		{tokenType: token.IDENT, literal: "Sub", pos: token.Position{Line: 2, Column: 4}},
		{tokenType: token.LPAREN, literal: "(", pos: token.Position{Line: 2, Column: 7}},
		{tokenType: token.IDENT, literal: "a", pos: token.Position{Line: 2, Column: 8}},
		{tokenType: token.IDENT, literal: "i32", pos: token.Position{Line: 2, Column: 10}},
		{tokenType: token.COLON, literal: ",", pos: token.Position{Line: 2, Column: 13}},
		{tokenType: token.IDENT, literal: "b", pos: token.Position{Line: 2, Column: 15}},
		{tokenType: token.IDENT, literal: "i32", pos: token.Position{Line: 2, Column: 17}},
		{tokenType: token.RPAREN, literal: ")", pos: token.Position{Line: 2, Column: 20}},
		{tokenType: token.COLON, literal: ":", pos: token.Position{Line: 2, Column: 22}},
		{tokenType: token.IDENT, literal: "i32", pos: token.Position{Line: 2, Column: 24}},
		{tokenType: token.LCURLY, literal: "{", pos: token.Position{Line: 2, Column: 28}},
		{tokenType: token.RETURN, literal: "return", pos: token.Position{Line: 3, Column: 2}},
		{tokenType: token.IDENT, literal: "a", pos: token.Position{Line: 3, Column: 9}},
		{tokenType: token.MINUS, literal: "-", pos: token.Position{Line: 3, Column: 11}},
		{tokenType: token.IDENT, literal: "b", pos: token.Position{Line: 3, Column: 13}},
		{tokenType: token.SEMICOLON, literal: ";", pos: token.Position{Line: 3, Column: 14}},
		{tokenType: token.RCURLY, literal: "}", pos: token.Position{Line: 4, Column: 1}},
		{tokenType: token.EOF, literal: string(rune(token.EOF)), pos: token.Position{Line: 5, Column: 1}},
	}

	lex := New(strings.NewReader(input))

	for i, test := range tests {
		tok := lex.NextToken()

		if tok.Type != test.tokenType {
			t.Errorf("tests[%d] %q - wrong type. Expected %q but got %q", i, test.literal, token.Print(test.tokenType), token.Print(tok.Type))
		}

		if tok.Lit != test.literal {
			t.Errorf("tests[%d] %q - wrong literal. Expected %q but got %q", i, test.literal, test.literal, tok.Lit)
		}

		if tok.Pos.Line != test.pos.Line {
			t.Errorf("tests[%d] %q - line number. Expected %d but got %d", i, test.literal, test.pos.Line, tok.Pos.Line)
		}

		if tok.Pos.Column != test.pos.Column {
			t.Errorf("tests[%d] %q - column number. Expected %d but got %d", i, test.literal, test.pos.Column, tok.Pos.Column)
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
