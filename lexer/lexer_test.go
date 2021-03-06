package lexer_test

import (
	"strings"
	"testing"

	"github.com/drejca/shift/lexer"
	"github.com/drejca/shift/token"
)

func TestNextToken(t *testing.T) {
	t.Parallel()

	type output struct {
		tokenType token.Type
		literal   string
	}

	testCases := []struct {
		name    string
		input   string
		outputs []output
	}{
		{
			name: "function",
			input: `fn Add(a i32, b i32) : i32 {
				return a + b
			}`,
			outputs: []output{
				{tokenType: token.FUNC, literal: "fn"},
				{tokenType: token.IDENT, literal: "Add"},
				{tokenType: token.LPAREN, literal: "("},
				{tokenType: token.IDENT, literal: "a"},
				{tokenType: token.IDENT, literal: "i32"},
				{tokenType: token.COMMA, literal: ","},
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
				{tokenType: token.RCURLY, literal: "}"},
				{tokenType: token.EOF, literal: string(rune(token.EOF))},
			},
		},
		{
			name:  "illegal symbol",
			input: `~`,
			outputs: []output{
				{tokenType: token.ILLEGAL, literal: "~"},
				{tokenType: token.EOF, literal: string(rune(token.EOF))},
			},
		},
		{
			name:  "subtraction expression with semicolon",
			input: `2 - 1;`,
			outputs: []output{
				{tokenType: token.INT, literal: "2"},
				{tokenType: token.MINUS, literal: "-"},
				{tokenType: token.INT, literal: "1"},
				{tokenType: token.SEMICOLON, literal: ";"},
				{tokenType: token.EOF, literal: string(rune(token.EOF))},
			},
		},
		{
			name:  "assignment with initialization with addition expression",
			input: `a := 4 + 5`,
			outputs: []output{
				{tokenType: token.IDENT, literal: "a"},
				{tokenType: token.INIT_ASSIGN, literal: ":="},
				{tokenType: token.INT, literal: "4"},
				{tokenType: token.PLUS, literal: "+"},
				{tokenType: token.INT, literal: "5"},
				{tokenType: token.EOF, literal: string(rune(token.EOF))},
			},
		},
		{
			name:  "assignment with addition expression",
			input: `a = a + 2`,
			outputs: []output{
				{tokenType: token.IDENT, literal: "a"},
				{tokenType: token.ASSIGN, literal: "="},
				{tokenType: token.IDENT, literal: "a"},
				{tokenType: token.PLUS, literal: "+"},
				{tokenType: token.INT, literal: "2"},
				{tokenType: token.EOF, literal: string(rune(token.EOF))},
			},
		},
		{
			name:  "assignment with multiply expression",
			input: `a = a * 2`,
			outputs: []output{
				{tokenType: token.IDENT, literal: "a"},
				{tokenType: token.ASSIGN, literal: "="},
				{tokenType: token.IDENT, literal: "a"},
				{tokenType: token.ASTERISK, literal: "*"},
				{tokenType: token.INT, literal: "2"},
				{tokenType: token.EOF, literal: string(rune(token.EOF))},
			},
		},
		{
			name:  "assignment with initialization floating point number",
			input: `f := 0.6`,
			outputs: []output{
				{tokenType: token.IDENT, literal: "f"},
				{tokenType: token.INIT_ASSIGN, literal: ":="},
				{tokenType: token.FLOAT, literal: "0.6"},
				{tokenType: token.EOF, literal: string(rune(token.EOF))},
			},
		},
		{
			name:  "import external function",
			input: `import fn log(num i32)`,
			outputs: []output{
				{tokenType: token.IMPORT, literal: "import"},
				{tokenType: token.FUNC, literal: "fn"},
				{tokenType: token.IDENT, literal: "log"},
				{tokenType: token.LPAREN, literal: "("},
				{tokenType: token.IDENT, literal: "num"},
				{tokenType: token.IDENT, literal: "i32"},
				{tokenType: token.RPAREN, literal: ")"},
				{tokenType: token.EOF, literal: string(rune(token.EOF))},
			},
		},
		{
			name:  "if expression",
			input: `if a != f {}`,
			outputs: []output{
				{tokenType: token.IF, literal: "if"},
				{tokenType: token.IDENT, literal: "a"},
				{tokenType: token.NOT_EQ, literal: "!="},
				{tokenType: token.IDENT, literal: "f"},
				{tokenType: token.LCURLY, literal: "{"},
				{tokenType: token.RCURLY, literal: "}"},
				{tokenType: token.EOF, literal: string(rune(token.EOF))},
			},
		},
		{
			name:  "string assignment",
			input: `name := "shift"`,
			outputs: []output{
				{tokenType: token.IDENT, literal: "name"},
				{tokenType: token.INIT_ASSIGN, literal: ":="},
				{tokenType: token.STRING, literal: "shift"},
				{tokenType: token.EOF, literal: string(rune(token.EOF))},
			},
		},
		{
			name:  "string assignment with interpolated variable",
			input: `interpolate := "this $a variable"`,
			outputs: []output{
				{tokenType: token.IDENT, literal: "interpolate"},
				{tokenType: token.INIT_ASSIGN, literal: ":="},
				{tokenType: token.STRING, literal: "this $a variable"},
				{tokenType: token.EOF, literal: string(rune(token.EOF))},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lex := lexer.New(strings.NewReader(tc.input))

			for _, output := range tc.outputs {
				tok := lex.NextToken()

				if tok.Type != output.tokenType {
					t.Errorf("line %d, column %d - wrong type. Expected %q but got %q", tok.Pos.Line, tok.Pos.Column, token.Print(output.tokenType), token.Print(tok.Type))
				}

				if tok.Lit != output.literal {
					t.Errorf("line %d, column %d - wrong literal. Expected %q but got %q", tok.Pos.Line, tok.Pos.Column, output.literal, tok.Lit)
				}
			}
		})
	}
}

func TestTokenPosition(t *testing.T) {
	t.Parallel()

	input := `
fn Sub(a i32, b i32) : i32 {
	return a - b
}
`
	tests := []struct {
		tokenType token.Type
		literal   string
		pos       token.Position
	}{
		{tokenType: token.FUNC, literal: "fn", pos: token.Position{Line: 2, Column: 1}},
		{tokenType: token.IDENT, literal: "Sub", pos: token.Position{Line: 2, Column: 4}},
		{tokenType: token.LPAREN, literal: "(", pos: token.Position{Line: 2, Column: 7}},
		{tokenType: token.IDENT, literal: "a", pos: token.Position{Line: 2, Column: 8}},
		{tokenType: token.IDENT, literal: "i32", pos: token.Position{Line: 2, Column: 10}},
		{tokenType: token.COMMA, literal: ",", pos: token.Position{Line: 2, Column: 13}},
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
		{tokenType: token.RCURLY, literal: "}", pos: token.Position{Line: 4, Column: 1}},
		{tokenType: token.EOF, literal: string(rune(token.EOF)), pos: token.Position{Line: 5, Column: 1}},
	}

	lex := lexer.New(strings.NewReader(input))

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
