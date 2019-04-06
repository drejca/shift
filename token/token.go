package token

type Token struct {
	Type Type
	Lit string // token literal text
	Pos Position // token position
}

type Position struct {
	Line     int    // line number, starting at 1
	Column   int    // column number, starting at 1
}

type Type int

const(
	EOF Type = iota
	ILLEGAL
	UNKNOWN

	PROGRAM

	// Identifiers + literals
	IDENT
	INT

	// Keywords
	FUNC
	RETURN
	LET

	// Delimiters
	LPAREN
	RPAREN
	COLON
	LCURLY
	RCURLY
	SEMICOLON

	// Operators
	PLUS
	MINUS
	ASSIGN
)

var tokens = map[Type]string {
	EOF: "EOF",
	ILLEGAL: "ILLEGAL",

	PROGRAM: "PROGRAM",

	// Identifiers + literals
	IDENT: "IDENT",
	INT: "INT",

	// Keywords
	FUNC: "FUNC",
	RETURN: "RETURN",
	LET: "LET",

	// Delimiters
	LPAREN: "(",
	RPAREN: ")",
	COLON: ",",
	LCURLY: "{",
	RCURLY: "}",
	SEMICOLON: ";",

	// Operators
	PLUS: "+",
	MINUS: "-",
	ASSIGN: "=",
}

func Print(tokenType Type) string {
	tokenStr, found := tokens[tokenType]
	if !found {
		return "unknown token type"
	}
	return tokenStr
}

func LookupIdent(ident string) Token {
	switch ident {
	case "fn":
		return Token{Type: FUNC, Lit: ident}
	case "return":
		return Token{Type: RETURN, Lit: ident}
	case "let":
		return Token{Type: LET, Lit: ident}
	}
	return Token{Type: IDENT, Lit: ident}
}