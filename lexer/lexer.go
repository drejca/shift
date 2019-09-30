package lexer

import (
	"bufio"
	"bytes"
	"io"

	"github.com/drejca/shift/token"
)

var eof = rune(token.EOF)

type Lexer struct {
	buf    *bufio.Reader
	ident  bytes.Buffer
	number bytes.Buffer

	pos      token.Position
	curRune  rune
	peekRune rune
}

func New(input io.Reader) *Lexer {
	lex := &Lexer{
		buf: bufio.NewReader(input),
		pos: token.Position{Line: 1, Column: 1},
	}
	return lex
}

func (l *Lexer) NextToken() token.Token {
	ch := l.read()

	if isNewLine(ch) {
		l.pos.Line++
		l.pos.Column = 1
		l.skipNewLine()
		return l.NextToken()
	}
	if ch == '\r' {
		l.pos.Line++
		l.pos.Column = 1
		ch = l.read()
		if ch == '\n' {
			return l.NextToken()
		}
	}
	if isWhitespace(ch) {
		l.skipWhitespace()
		return l.NextToken()
	}
	if isLetter(ch) {
		l.unread()
		return l.readIdentifier()
	}
	if isDigit(ch) {
		l.unread()
		return l.readNumber()
	}

	switch ch {
	case ',':
		return l.Token(token.COMMA, string(ch))
	case ':':
		if l.peek() == '=' {
			l.read()
			return l.Token(token.INIT_ASSIGN, string(":="))
		}
		return l.Token(token.COLON, string(ch))
	case ';':
		return l.Token(token.SEMICOLON, string(ch))
	case '(':
		return l.Token(token.LPAREN, string(ch))
	case ')':
		return l.Token(token.RPAREN, string(ch))
	case '{':
		return l.Token(token.LCURLY, string(ch))
	case '}':
		return l.Token(token.RCURLY, string(ch))
	case '+':
		return l.Token(token.PLUS, string(ch))
	case '-':
		return l.Token(token.MINUS, string(ch))
	case '!':
		if l.peek() == '=' {
			l.read()
			return l.Token(token.NOT_EQ, string("!="))
		}
		return l.Token(token.BANG, string(ch))
	case '=':
		return l.Token(token.ASSIGN, string(ch))
	case eof:
		return l.Token(token.EOF, string(ch))
	default:
		return l.Token(token.ILLEGAL, string(ch))
	}
}

func (l *Lexer) skipNewLine() {
	l.ident.Reset()
	ch := l.read()
	if ch != '\n' {
		l.unread()
	}
	l.ident.Reset()
}

func (l *Lexer) skipWhitespace() {
	l.ident.Reset()
	for {
		ch := l.read()
		if !isWhitespace(ch) {
			l.unread()
			return
		}
	}
}

func (l *Lexer) readIdentifier() token.Token {
	l.ident.Reset()
	for {
		ch := l.read()
		if !isLetter(ch) && !isDigit(ch) {
			l.unread()
			break
		}
		l.ident.WriteRune(ch)
	}
	tok := token.LookupIdent(l.ident.String())
	tok.Pos = l.pos
	tok.Pos.Column = tok.Pos.Column - len(tok.Lit)
	return tok
}

func (l *Lexer) readNumber() token.Token {
	l.number.Reset()
	for {
		ch := l.read()

		if ch == '.' {
			return l.readFloat()
		}

		if !isDigit(ch) {
			l.unread()
			break
		}
		l.number.WriteRune(ch)
	}
	tok := l.Token(token.INT, l.number.String())
	return tok
}

func (l *Lexer) readFloat() token.Token {
	l.number.WriteRune('.')
	for {
		ch := l.read()

		if !isDigit(ch) {
			l.unread()
			break
		}
		l.number.WriteRune(ch)
	}
	tok := l.Token(token.FLOAT, l.number.String())
	return tok
}

func (l *Lexer) Token(tokenType token.Type, literal string) token.Token {
	tok := token.Token{Type: tokenType, Lit: literal}
	tok.Pos = l.pos
	tok.Pos.Column = tok.Pos.Column - len(tok.Lit)
	return tok
}

func (l *Lexer) peek() rune {
	ch := l.read()
	l.unread()
	return ch
}

func (l *Lexer) read() rune {
	ch, _, _ := l.buf.ReadRune()
	l.pos.Column++
	return ch
}

func (l *Lexer) unread() {
	l.buf.UnreadRune()
	l.pos.Column--
}

func isNewLine(ch rune) bool    { return ch == '\n' || ch == '\r' }
func isWhitespace(ch rune) bool { return ch == ' ' || ch == '\t' }
func isLetter(ch rune) bool     { return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' }
func isDigit(ch rune) bool      { return '0' <= ch && ch <= '9' }
