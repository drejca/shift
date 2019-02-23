package ast

import (
	"bytes"
	"github.com/drejca/shiftlang/token"
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
	ReturnParams *ReturnParamGroup
	Body *BlockStatement
}

func (f *Function) statementNode() {}
func (f *Function) String() string {
	var out bytes.Buffer

	out.WriteString("fn ")
	out.WriteString(f.Name)
	out.WriteString("()")

	if f.ReturnParams != nil {
		out.WriteString(f.ReturnParams.String())
	}

	out.WriteString(" {")
	if f.Body != nil {
		out.WriteString(f.Body.String())
	}
	out.WriteString("\n}")

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

	out.WriteString(";")

	return out.String()
}

type Identifier struct {
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) String() string       { return i.Value }
