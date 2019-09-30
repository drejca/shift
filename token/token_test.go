package token

import "testing"

func TestPrintToken(t *testing.T) {
	for tokType, tok := range tokens {
		tokenStr := Print(tokType)
		if tokenStr != tok {
			t.Errorf("token %d not found", tokType)
		}
	}
}

func TestPrintUnknownToken(t *testing.T) {
	tokenStr := Print(UNKNOWN)
	expected := "unknown token type"
	if tokenStr != expected {
		t.Errorf("expected %q got %q", expected, tokenStr)
	}
}

func TestLookupIdent(t *testing.T) {
	tests := []struct {
		ident       string
		expectToken Token
	}{
		{ident: "fn", expectToken: Token{Lit: "fn", Type: FUNC}},
		{ident: "return", expectToken: Token{Lit: "return", Type: RETURN}},
		{ident: "name", expectToken: Token{Lit: "name", Type: IDENT}},
		{ident: "import", expectToken: Token{Lit: "import", Type: IMPORT}},
		{ident: "if", expectToken: Token{Lit: "if", Type: IF}},
	}

	for _, test := range tests {
		tok := LookupIdent(test.ident)

		if tok.Type != test.expectToken.Type {
			t.Errorf("expected %q got %q", Print(test.expectToken.Type), Print(tok.Type))
		}
	}
}
