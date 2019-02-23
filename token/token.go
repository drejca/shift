package token

type Token struct {
	Type Type
	Lit string // token literal text
}

type Type int

const(
	EOF Type = iota
	ILLEGAL

	PROGRAM
	IDENT

	// Keywords
	FUNC
	RETURN

	// Delimiters
	LPAREN
	RPAREN
	COLON
	LCURLY
	RCURLY
	SEMICOLON
)

func Print(tokenType Type) string {
	switch tokenType {
	case EOF:
		return "EOF"
	case ILLEGAL:
		return "ILLEGAL"
	case PROGRAM:
		return "PROGRAM"
	case IDENT:
		return "IDENT"
	case FUNC:
		return "FUNC"
	case RETURN:
		return "RETURN"
	case LPAREN:
		return "LPAREN"
	case RPAREN:
		return "RPAREN"
	case COLON:
		return "COLON"
	case LCURLY:
		return "LCURLY"
	case RCURLY:
		return "RCURLY"
	case SEMICOLON:
		return "SEMICOLON"
	}
	return "unknown type"
}

func LookupIdent(ident string) Token {
	switch ident {
	case "fn":
		return Token{Type: FUNC, Lit: ident}
	case "return":
		return Token{Type: RETURN, Lit: ident}
	}
	return Token{Type: IDENT, Lit: ident}
}