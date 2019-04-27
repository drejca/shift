package parser

import (
	"errors"
	"fmt"
	"github.com/drejca/shift/ast"
	"github.com/drejca/shift/lexer"
	"github.com/drejca/shift/token"
	"io"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	SUM   // +
	MINUS // -
	CALL
)

var precedences = map[token.Type]int{
	token.PLUS:  SUM,
	token.MINUS: MINUS,
	token.RPAREN: LOWEST,
	token.LPAREN:   CALL,
}

type Parser struct {
	l *lexer.Lexer

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.Type]prefixParseFn
	infixParseFns  map[token.Type]infixParseFn

	errors []error
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

func New(input io.Reader) *Parser {
	p := &Parser{l: lexer.New(input)}

	p.prefixParseFns = make(map[token.Type]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)

	p.infixParseFns = make(map[token.Type]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)

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
	case token.LET:
		return p.parseLetStatement()
	default:
		return p.parseExpressionStatement()
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

	if p.peekTokenIs(token.IDENT) {
		fn.InputParams = p.parseInputParameters()
	} else if !p.expectPeek(token.RPAREN) {
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

func (p *Parser) parseInputParameters() []*ast.Parameter {
	var inputParams []*ast.Parameter

	for {
		p.nextToken()

		if p.curTokenIs(token.RPAREN) {
			return inputParams
		}

		param := &ast.Parameter{}
		if p.curTokenIs(token.IDENT) {
			param.Ident = &ast.Identifier{Value: p.curToken.Lit}
		}
		p.nextToken()
		if p.curTokenIs(token.IDENT) {
			param.Type = p.curToken.Lit
		}
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
		inputParams = append(inputParams, param)
	}
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
	if p.curTokenIs(token.EOF) {
		p.peekError(token.RCURLY)
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

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Lit}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	if p.curTokenIs(token.IDENT) && p.peekTokenIs(token.ASSIGN) {
		stmt.Expression = p.parseAssignmentExpression()
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseAssignmentExpression() ast.Expression {
	stmt := &ast.AssignmentExpression{
		Token: p.curToken,
		Identifier: p.parseIdentifier(),
	}
	p.nextToken()
	p.nextToken()

	stmt.Expression = p.parseExpression(LOWEST)

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

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()

		leftExp = infix(leftExp)
	}
	return leftExp
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

func (p *Parser) parseExpressionList(end token.Type) []ast.Expression {
	var list []ast.Expression

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) peekPrecedence() int {
	if precedence, ok := precedences[p.peekToken.Type]; ok {
		return precedence
	}
	return LOWEST
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Lit,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) curPrecedence() int {
	if precedence, ok := precedences[p.curToken.Type]; ok {
		return precedence
	}
	return LOWEST
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Value: p.curToken.Lit}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Lit, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Lit)
		p.errors = append(p.errors, errors.New(msg))
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) noPrefixParseFnError() {
	err := fmt.Errorf("illegal symbol %s", p.curToken.Lit)
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
	err := fmt.Errorf("missing %s", token.Print(tokenType))
	p.errors = append(p.errors, err)
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) registerPrefix(tokenType token.Type, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.Type, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}
