package token

type Token struct {
	Type Type
	Lit  string   // token literal text
	Pos  Position // token position
}

type Position struct {
	Line   int // line number, starting at 1
	Column int // column number, starting at 1
}

type CompileError interface {
	Error() error
	Position() Position
}

type Type int

const (
	EOF Type = iota
	ILLEGAL
	UNKNOWN

	PROGRAM

	// Identifiers + literals
	IDENT
	INT
	FLOAT
	STRING

	// Keywords
	FUNC
	RETURN
	IMPORT
	IF

	// Delimiters
	COMMA
	COLON
	SEMICOLON

	LPAREN
	RPAREN
	LCURLY
	RCURLY

	// Operators
	PLUS
	MINUS
	ASTERISK
	ASSIGN
	INIT_ASSIGN
	BANG
	NOT_EQ
)

var Tokens = map[Type]string{
	EOF:     "EOF",
	ILLEGAL: "ILLEGAL",

	PROGRAM: "PROGRAM",

	// Identifiers + literals
	IDENT:  "IDENT",
	INT:    "INT",
	FLOAT:  "FLOAT",
	STRING: "STRING",

	// Keywords
	FUNC:   "FUNC",
	RETURN: "RETURN",
	IMPORT: "IMPORT",
	IF:     "IF",

	// Delimiters
	COMMA:     ",",
	COLON:     ":",
	SEMICOLON: ";",

	LPAREN: "(",
	RPAREN: ")",
	LCURLY: "{",
	RCURLY: "}",

	// Operators
	PLUS:        "+",
	MINUS:       "-",
	ASTERISK:    "*",
	ASSIGN:      "=",
	INIT_ASSIGN: ":=",
	BANG:        "!",

	NOT_EQ: "!=",
}

// Print returns string name of token.Type
func Print(tokenType Type) string {
	tokenStr, found := Tokens[tokenType]
	if !found {
		return "unknown token type"
	}
	return tokenStr
}

// LookupIdent returns Token for ident
func LookupIdent(ident string) Token {
	switch ident {
	case "fn":
		return Token{Type: FUNC, Lit: ident}
	case "return":
		return Token{Type: RETURN, Lit: ident}
	case "import":
		return Token{Type: IMPORT, Lit: ident}
	case "if":
		return Token{Type: IF, Lit: ident}
	}
	return Token{Type: IDENT, Lit: ident}
}
