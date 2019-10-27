package token_test

import (
	"testing"

	"github.com/drejca/shift/token"
)

func TestPrintToken(t *testing.T) {
	for tokType, tok := range token.Tokens {
		tokenStr := token.Print(tokType)
		if tokenStr != tok {
			t.Errorf("token %d not found", tokType)
		}
	}
}

func TestPrintUnknownToken(t *testing.T) {
	tokenStr := token.Print(token.UNKNOWN)
	expected := "unknown token type"
	if tokenStr != expected {
		t.Errorf("expected %q got %q", expected, tokenStr)
	}
}

func TestLookupIdent(t *testing.T) {
	tests := []struct {
		ident       string
		expectToken token.Token
	}{
		{ident: "fn", expectToken: token.Token{Lit: "fn", Type: token.FUNC}},
		{ident: "return", expectToken: token.Token{Lit: "return", Type: token.RETURN}},
		{ident: "name", expectToken: token.Token{Lit: "name", Type: token.IDENT}},
		{ident: "import", expectToken: token.Token{Lit: "import", Type: token.IMPORT}},
		{ident: "if", expectToken: token.Token{Lit: "if", Type: token.IF}},
	}

	for _, test := range tests {
		tok := token.LookupIdent(test.ident)

		if tok.Type != test.expectToken.Type {
			t.Errorf("expected %q got %q", token.Print(test.expectToken.Type), token.Print(tok.Type))
		}
	}
}
