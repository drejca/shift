package parser

import (
	"fmt"
	"github.com/drejca/shiftlang/ast"
	"github.com/drejca/shiftlang/lexer"
	"github.com/drejca/shiftlang/token"
	"io"
)

const (
	_ int = iota
	LOWEST
)

type Parser struct {
	l *lexer.Lexer

	curToken token.Token
	peekToken token.Token

	prefixParseFns map[token.Type]prefixParseFn

	errors []error
}

type (
	prefixParseFn func() ast.Expression
)

func New(input io.Reader) *Parser {
	p := &Parser{l: lexer.New(input)}

	p.prefixParseFns = make(map[token.Type]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)

	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) Parse() *ast.Program {
	program := &ast.Program{}

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.FUNC:
		return p.parseFunc()
	case token.RETURN:
		return p.parseReturn()
	default:
		return nil
	}
}

func (p *Parser) parseFunc() *ast.Function {
	p.nextToken()

	fn := &ast.Function{
		Name: p.curToken.Lit,
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if p.peekTokenIs(token.COLON) {
		p.nextToken()
		fn.ReturnParams = p.parseReturnParameters()
	} else if !p.expectPeek(token.LCURLY) {
		return nil
	}

	fn.Body = p.parseBlockStatement()

	return fn
}

func (p *Parser) parseReturnParameters() *ast.ReturnParamGroup {
	paramGroup := &ast.ReturnParamGroup{}

	for {
		p.nextToken()

		if p.curTokenIs(token.LCURLY) {
			return paramGroup
		}

		if p.curTokenIs(token.IDENT) {
			paramGroup.Params = append(paramGroup.Params, &ast.Parameter{Type: p.curToken.Lit})
		}
	}
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{FirstToken: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.RCURLY) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	return block
}

func (p *Parser) parseReturn() *ast.ReturnStatement {
	p.nextToken()

	stmt := &ast.ReturnStatement{}
	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError()
		return nil
	}
	leftExp := prefix()
	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Value: p.curToken.Lit}
}

func (p *Parser) noPrefixParseFnError() {
	err := fmt.Errorf("no prefix parse function for %s found", token.Print(p.curToken.Type))
	p.errors = append(p.errors, err)
}

func (p *Parser) expectPeek(tokenType token.Type) bool {
	if p.peekTokenIs(tokenType) {
		p.nextToken()
		return true
	} else {
		p.peekError(tokenType)
		return false
	}
}

func (p *Parser) curTokenIs(tokenType token.Type) bool {
	if p.curToken.Type == tokenType {
		return true
	}
	return false
}

func (p *Parser) peekTokenIs(tokenType token.Type) bool {
	return p.peekToken.Type == tokenType
}

func (p *Parser) peekError(tokenType token.Type) {
	err := fmt.Errorf("expected next token to be %s, got %s(%s) instead",
		token.Print(tokenType), token.Print(p.peekToken.Type), p.peekToken.Lit)
	p.errors = append(p.errors, err)
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) registerPrefix(tokenType token.Type, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}
