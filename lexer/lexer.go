package lexer

import (
	"bufio"
	"bytes"
	"github.com/drejca/shiftlang/token"
	"io"
)

var eof = rune(token.EOF)

type Lexer struct {
	buf *bufio.Reader
	ident bytes.Buffer
}

func New(input io.Reader) *Lexer {
	return &Lexer{buf: bufio.NewReader(input)}
}

func (l *Lexer) NextToken() token.Token {
	ch := l.read()

	if isNewLine(ch) {
		l.ident.Reset()
		return l.NextToken()
	}
	if isWhitespace(ch) {
		ch = l.skipWhitespace()
	}
	if isLetter(ch) || isNumber(ch) {
		l.unread()
		return l.readIdentifier()
	}

	switch ch {
	case ',':
		return token.Token{Type: token.COLON, Lit: string(ch)}
	case ';':
		return token.Token{Type: token.SEMICOLON, Lit: string(ch)}
	case '(':
		return token.Token{Type: token.LPAREN, Lit: string(ch)}
	case ')':
		return token.Token{Type: token.RPAREN, Lit: string(ch)}
	case '{':
		return token.Token{Type: token.LCURLY, Lit: string(ch)}
	case '}':
		return token.Token{Type: token.RCURLY, Lit: string(ch)}
	case ':':
		return token.Token{Type: token.COLON, Lit: string(ch)}
	case '+':
		return token.Token{Type: token.PLUS, Lit: string(ch)}
	case eof:
		return token.Token{Type: token.EOF, Lit: string(ch)}
	default:
		return token.Token{Type: token.ILLEGAL, Lit: string(ch)}
	}
}

func (l *Lexer) skipWhitespace() rune {
	l.ident.Reset()
	for {
		ch := l.read()
		if !isWhitespace(ch) {
			return ch
		}
	}
}

func (l *Lexer) readIdentifier() token.Token {
	l.ident.Reset()
	for {
		ch := l.read()
		if !isLetter(ch) && !isNumber(ch) {
			l.unread()
			break
		}
		l.ident.WriteRune(ch)
	}
	tok := token.LookupIdent(l.ident.String())
	return tok
}

func (l *Lexer) read() rune {
	ch, _, _ := l.buf.ReadRune()
	return ch
}

func (l *Lexer) unread() {
	l.buf.UnreadRune()
}

func isNewLine(ch rune) bool    { return ch == '\n' || ch == '\r' }
func isWhitespace(ch rune) bool { return ch == ' ' || ch == '\t' }
func isLetter(ch rune) bool { return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'}
func isNumber(ch rune) bool { return '0' <= ch && ch <= '9'}
