package parser

import (
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
}

type ParseError struct {
	position token.Position
	error error
}
func (p ParseError) Position() token.Position {
	return p.position
}
func (p ParseError) Error() error {
	return p.error
}

type (
	prefixParseFn func() (ast.Expression, token.CompileError)
	infixParseFn  func(ast.Expression) (ast.Expression, token.CompileError)
)

func New(input io.Reader) *Parser {
	p := &Parser{l: lexer.New(input)}

	p.prefixParseFns = make(map[token.Type]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)

	p.infixParseFns = make(map[token.Type]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)

	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) ParseProgram() (*ast.Program, token.CompileError) {
	program := &ast.Program{}

	for p.curToken.Type != token.EOF {
		stmt, err := p.parseGlobalStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program, nil
}

func (p *Parser) parseGlobalStatement() (ast.Statement, token.CompileError) {
	switch p.curToken.Type {
	case token.FUNC:
		return p.parseFunc()
	case token.LET:
		return p.parseLetStatement()
	}
	return nil, p.parseError(fmt.Errorf("non-declaration statement outside function body"), p.curToken, p.curToken.Pos.Column - 1)
}

func (p *Parser) parseLocalStatement() (ast.Statement, token.CompileError) {
	switch p.curToken.Type {
	case token.RETURN:
		return p.parseReturn()
	case token.LET:
		return p.parseLetStatement()
	}
	return p.parseExpressionStatement()
}

func (p *Parser) parseFunc() (*ast.Function, token.CompileError) {
	if !p.expectPeek(token.IDENT) {
		return nil, p.parseError(fmt.Errorf("missing function name"), p.curToken, p.curToken.Pos.Column + 2)
	}

	fn := &ast.Function{
		Name: p.curToken.Lit,
	}

	if !p.expectPeek(token.LPAREN) {
		return nil, p.peekError(token.LPAREN)
	}

	if p.peekTokenIs(token.IDENT) {
		inputParams, err := p.parseInputParameters()
		if err != nil {
			return nil, err
		}
		fn.InputParams = inputParams
	}

	if !p.expectPeek(token.RPAREN) {
		return nil, p.peekError(token.RPAREN)
	}

	if p.peekTokenIs(token.COLON) {
		p.nextToken()

		returnParams, err := p.parseReturnParameters()
		if err != nil {
			return nil, err
		}
		fn.ReturnParams = returnParams
	}

	if !p.expectPeek(token.LCURLY) {
		return nil, p.peekError(token.LCURLY)
	}

	stmt, err := p.parseBlockStatement()
	if err != nil {
		return nil, err
	}
	fn.Body = stmt

	return fn, nil
}

func (p *Parser) parseInputParameters() ([]*ast.Parameter, token.CompileError) {
	var inputParams []*ast.Parameter

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return inputParams, nil
	}

	param, err := p.parseInputParam()
	if err != nil {
		return nil, err
	}

	inputParams = append(inputParams, param)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()

		param, err := p.parseInputParam()
		if err != nil {
			return nil, err
		}

		inputParams = append(inputParams, param)
	}
	return inputParams, nil
}

func (p *Parser) parseInputParam() (*ast.Parameter, token.CompileError) {
	param := &ast.Parameter{}
	if p.peekTokenIs(token.IDENT) {
		param.Ident = &ast.Identifier{Value: p.peekToken.Lit}
	} else {
		return nil, p.parseError(fmt.Errorf("trailing comma in parameters"), p.curToken, p.curToken.Pos.Column - 1)
	}
	p.nextToken()

	if p.peekTokenIs(token.IDENT) {
		param.Type = p.peekToken.Lit
	} else {
		return nil, p.parseError(fmt.Errorf("missing function parameter type"), p.curToken, p.curToken.Pos.Column)
	}
	p.nextToken()

	return param, nil
}

func (p *Parser) parseReturnParameters() ([]*ast.Parameter, token.CompileError) {
	var params []*ast.Parameter

	p.nextToken()

	if p.curTokenIs(token.IDENT) {
		params = append(params, &ast.Parameter{Type: p.curToken.Lit})
	}

	if p.peekTokenIs(token.LCURLY) {
		return params, nil
	}
	return nil, p.peekError(token.LCURLY)
}

func (p *Parser) parseBlockStatement() (*ast.BlockStatement, token.CompileError) {
	block := &ast.BlockStatement{FirstToken: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.RCURLY) && !p.curTokenIs(token.EOF) {
		stmt, err := p.parseLocalStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	if p.curTokenIs(token.EOF) {
		return nil, p.peekError(token.RCURLY)
	}
	return block, nil
}

func (p *Parser) parseReturn() (*ast.ReturnStatement, token.CompileError) {
	p.nextToken()

	stmt := &ast.ReturnStatement{}
	expression, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	stmt.ReturnValue = expression

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt, nil
}

func (p *Parser) parseLetStatement() (*ast.LetStatement, token.CompileError) {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil, p.peekError(token.IDENT)
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Lit}
	stmt.Type = p.parseType()

	if !p.expectPeek(token.ASSIGN) {
		return nil, p.peekError(token.ASSIGN)
	}

	p.nextToken()

	expression, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	stmt.Value = expression

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt, nil
}

func (p* Parser) parseType() string {
	if p.peekTokenIs(token.IDENT) {
		p.nextToken()
		return p.curToken.Lit
	}
	return ""
}

func (p *Parser) parseExpressionStatement() (*ast.ExpressionStatement, token.CompileError) {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	if p.curTokenIs(token.IDENT) && p.peekTokenIs(token.ASSIGN) {
		expression, err :=  p.parseAssignmentExpression()
		if err != nil {
			return nil, err
		}
		stmt.Expression = expression
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt, nil
}

func (p *Parser) parseAssignmentExpression() (ast.Expression, token.CompileError) {
	expression, err := p.parseIdentifier()
	if err != nil {
		return nil, err
	}

	stmt := &ast.AssignmentExpression{
		Token: p.curToken,
		Identifier: expression,
	}
	p.nextToken()
	p.nextToken()

	expression, err = p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	stmt.Expression = expression

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt, nil
}

func (p *Parser) parseExpression(precedence int) (ast.Expression, token.CompileError) {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		return nil, p.parseError(fmt.Errorf("illegal symbol %s", p.curToken.Lit), p.curToken, p.curToken.Pos.Column)
	}
	leftExp, err := prefix()
	if err != nil {
		return nil, err
	}

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp, p.parseError(fmt.Errorf("illegal symbol %s", p.curToken.Lit), p.curToken, p.curToken.Pos.Column)
		}
		p.nextToken()

		expression, err := infix(leftExp)
		if err != nil {
			return nil, err
		}
		leftExp = expression
	}
	return leftExp, nil
}

func (p *Parser) parseCallExpression(function ast.Expression) (ast.Expression, token.CompileError) {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	expression, err := p.parseExpressionList(token.RPAREN)
	if err != nil {
		return nil, err
	}
	exp.Arguments = expression
	return exp, nil
}

func (p *Parser) parseExpressionList(end token.Type) ([]ast.Expression, token.CompileError) {
	var list []ast.Expression

	if p.peekTokenIs(end) {
		p.nextToken()
		return list, nil
	}

	p.nextToken()
	expression, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	list = append(list, expression)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()

		expression, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		list = append(list, expression)
	}

	if !p.expectPeek(end) {
		return nil, p.peekError(end)
	}

	return list, nil
}

func (p *Parser) peekPrecedence() int {
	if precedence, ok := precedences[p.peekToken.Type]; ok {
		return precedence
	}
	return LOWEST
}

func (p *Parser) parseInfixExpression(left ast.Expression) (ast.Expression, token.CompileError) {
	infixExpression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Lit,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()

	expression, err := p.parseExpression(precedence)
	if err != nil {
		return nil, err
	}

	infixExpression.Right = expression

	return infixExpression, nil
}

func (p *Parser) parseGroupedExpression() (ast.Expression, token.CompileError) {
	p.nextToken()

	exp, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}

	if !p.expectPeek(token.RPAREN) {
		return nil, p.peekError(token.RPAREN)
	}
	return exp, nil
}

func (p *Parser) curPrecedence() int {
	if precedence, ok := precedences[p.curToken.Type]; ok {
		return precedence
	}
	return LOWEST
}

func (p *Parser) parseIdentifier() (ast.Expression, token.CompileError) {
	return &ast.Identifier{Value: p.curToken.Lit}, nil
}

func (p *Parser) parseIntegerLiteral() (ast.Expression, token.CompileError) {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Lit, 0, 64)
	if err != nil {
		return nil, p.parseError(fmt.Errorf("could not parse %q as integer", p.curToken.Lit), p.curToken, p.curToken.Pos.Column)
	}
	lit.Value = value
	return lit, nil
}

func (p *Parser) parseFloatLiteral() (ast.Expression, token.CompileError) {
	lit := &ast.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Lit, 64)
	if err != nil {
		return nil, p.parseError(fmt.Errorf("could not parse %q as integer", p.curToken.Lit), p.curToken, p.curToken.Pos.Column)
	}
	lit.Value = value
	return lit, nil
}

func (p *Parser) expectToken(tokenType token.Type) bool {
	if p.curTokenIs(tokenType) {
		return true
	} else {
		p.error(tokenType)
		return false
	}
}

func (p *Parser) expectPeek(tokenType token.Type) bool {
	if p.peekTokenIs(tokenType) {
		p.nextToken()
		return true
	} else {
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

func (p *Parser) error(tokenType token.Type) token.CompileError {
	err := fmt.Errorf("missing %s", token.Print(tokenType))
	return ParseError{position: p.curToken.Pos, error: err}
}

func (p *Parser) peekError(tokenType token.Type) token.CompileError {
	err := fmt.Errorf("missing %s", token.Print(tokenType))
	return ParseError{position: p.peekToken.Pos, error: err}
}

func (p *Parser) parseError(err error, tok token.Token, column int) token.CompileError {
	pos := token.Position{Line: tok.Pos.Line, Column: column}
	return ParseError{position: pos, error: err}
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
