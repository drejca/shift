package ast

import (
	"bytes"
	"github.com/drejca/shift/token"
	"strings"
)

type Node interface {
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type ProgramError interface {
	Error() error
	Position() token.Position
}

type Program struct {
	Statements []Statement
}
func (p *Program) String() string {
	var out bytes.Buffer

	for _, stmt := range p.Statements {
		out.WriteString(stmt.String())
	}
	return out.String()
}

type Function struct {
	Name string
	InputParams []*Parameter
	ReturnParams *ReturnParamGroup
	Body *BlockStatement
}
func (f *Function) statementNode() {}
func (f *Function) String() string {
	var out bytes.Buffer

	out.WriteString("\n")
	out.WriteString("fn ")
	out.WriteString(f.Name)
	out.WriteString("(")
	for i, param := range f.InputParams {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(param.Ident.Value)
		out.WriteString(" ")
		out.WriteString(param.Type)
	}
	out.WriteString(")")

	if f.ReturnParams != nil {
		out.WriteString(f.ReturnParams.String())
	}

	out.WriteString(" {")
	if f.Body != nil {
		out.WriteString(f.Body.String())
	}
	out.WriteString("\n}\n")

	return out.String()
}

type ReturnParamGroup struct {
	Params []*Parameter
}
func (p *ReturnParamGroup) String() string {
	var out bytes.Buffer

	out.WriteString(" :")
	for _, param := range p.Params {
		out.WriteString(" ")
		out.WriteString(param.String())
	}

	return out.String()
}

type Parameter struct {
	Ident *Identifier
	Type string
}
func (p *Parameter) String() string {
	var out bytes.Buffer

	if p.Ident != nil {
		out.WriteString(p.Ident.String())
		out.WriteString(" ")
	}
	out.WriteString(p.Type)

	return out.String()
}

type BlockStatement struct {
	FirstToken token.Token
	Statements []Statement
}
func (b *BlockStatement) statementNode() {}
func (b *BlockStatement) String() string {
	var out bytes.Buffer

	for _, stmt := range b.Statements {
		out.WriteString("\n\t")
		out.WriteString(stmt.String())
	}
	return out.String()
}

type ReturnStatement struct {
	ReturnValue Expression
}
func (r *ReturnStatement) statementNode()       {}
func (r *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString("return ")

	if r.ReturnValue != nil {
		out.WriteString(r.ReturnValue.String())
	}

	return out.String()
}

type LetStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}
func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.Token.Lit + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}

	return out.String()
}

type ExpressionStatement struct {
	Token      token.Token // the first token of the expression
	Expression Expression
}
func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

type CallExpression struct {
	Token     token.Token // The '(' token
	Function  Expression
	Arguments []Expression
}
func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	var args []string
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

type AssignmentExpression struct {
	Token token.Token
	Identifier Expression
	Expression Expression
}
func (as *AssignmentExpression) expressionNode() {}
func (as *AssignmentExpression) String() string {
	var out bytes.Buffer

	out.WriteString(as.Identifier.String())
	out.WriteString(" = ")
	if as.Expression != nil {
		out.WriteString(as.Expression.String())
	}
	return out.String()
}

type InfixExpression struct {
	Token    token.Token // The operator token, e.g. +
	Left     Expression
	Operator string
	Right    Expression
}
func (oe *InfixExpression) expressionNode()      {}
func (oe *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(oe.Left.String())
	out.WriteString(" " + oe.Operator + " ")
	out.WriteString(oe.Right.String())
	out.WriteString(")")

	return out.String()
}

type Identifier struct {
	Token token.Token
	Value string
}
func (i *Identifier) expressionNode()      {}
func (i *Identifier) String() string       { return i.Value }

type IntegerLiteral struct {
	Token token.Token
	Value int64
}
func (i *IntegerLiteral) expressionNode()      {}
func (i *IntegerLiteral) String() string       { return i.Token.Lit }
