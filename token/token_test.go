package token

import "testing"

func TestPrintTokenType(t *testing.T) {
	tests := []struct{
		tokenType Type
		printVal string
	} {
		{tokenType: -1, printVal: "unknown type"},
		{tokenType: EOF, printVal: "EOF"},
		{tokenType: ILLEGAL, printVal: "ILLEGAL"},
		{tokenType: PROGRAM, printVal: "PROGRAM"},
		{tokenType: IDENT, printVal: "IDENT"},
		{tokenType: FUNC, printVal: "FUNC"},
		{tokenType: RETURN, printVal: "RETURN"},
		{tokenType: LPAREN, printVal: "LPAREN"},
		{tokenType: RPAREN, printVal: "RPAREN"},
		{tokenType: COLON, printVal: "COLON"},
		{tokenType: LCURLY, printVal: "LCURLY"},
		{tokenType: RCURLY, printVal: "RCURLY"},
		{tokenType: SEMICOLON, printVal: "SEMICOLON"},
	}

	for i, test := range tests {
		printVal := Print(test.tokenType)

		if printVal != test.printVal {
			t.Errorf("tests[%d] - wrong print value. Expected %q but got %q", i, test.printVal, printVal)
		}
	}
}

func TestLookupIdent(t *testing.T) {
	tests := []struct{
		ident string
		token Token
	} {
		{ident: "fn", token: Token{Type: FUNC, Lit: "fn"}},
		{ident: "main", token: Token{Type: IDENT, Lit: "main"}},
	}

	for i, test := range tests {
		tok := LookupIdent(test.ident)

		if tok.Type != test.token.Type {
			t.Errorf("tests[%d] - wrong literal. Expected %q but got %q", i, Print(test.token.Type), Print(tok.Type))
		}
		if tok.Lit != test.token.Lit {
			t.Errorf("tests[%d] - wrong literal. Expected %q but got %q", i, test.token.Lit, tok.Lit)
		}
	}
}
