package ast

import (
	"bytes"
	"strings"

	"github.com/drejca/shift/token"
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
	Signature *FunctionSignature
	Body      *BlockStatement
}

func (f *Function) statementNode() {}
func (f *Function) String() string {
	var out bytes.Buffer

	out.WriteString("\n")
	out.WriteString("fn ")

	out.WriteString(f.Signature.String())

	if f.Body != nil {
		out.WriteString(f.Body.String())
	}
	out.WriteString("\n")
	return out.String()
}

type FunctionSignature struct {
	Name         string
	InputParams  []*Parameter
	ReturnParams []*Parameter
}

func (f *FunctionSignature) statementNode() {}
func (f *FunctionSignature) String() string {
	var out bytes.Buffer

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

	if len(f.ReturnParams) > 0 {
		out.WriteString(" :")
	}
	for _, param := range f.ReturnParams {
		out.WriteString(" ")
		out.WriteString(param.String())
	}
	return out.String()
}

type Parameter struct {
	Ident *Identifier
	Type  string
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
	Depth      int
}

func (b *BlockStatement) statementNode() {}
func (b *BlockStatement) String() string {
	var out bytes.Buffer

	out.WriteString(" {")
	for _, stmt := range b.Statements {
		out.WriteString("\n")
		out.WriteString(indent(b.Depth))
		out.WriteString(stmt.String())
	}
	out.WriteString("\n")
	out.WriteString(indent(b.Depth - 1))
	out.WriteString("}")
	return out.String()
}

func indent(depth int) string {
	var out bytes.Buffer
	for i := 0; i < depth; i++ {
		out.WriteString("\t")
	}
	return out.String()
}

type ReturnStatement struct {
	ReturnValue Expression
}

func (r *ReturnStatement) statementNode() {}
func (r *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString("return ")

	if r.ReturnValue != nil {
		out.WriteString(r.ReturnValue.String())
	}

	return out.String()
}

type ImportStatement struct {
	FuncSignature *FunctionSignature
}

func (is *ImportStatement) statementNode() {}
func (is *ImportStatement) String() string {
	var out bytes.Buffer

	out.WriteString("\n")
	out.WriteString("import fn ")
	out.WriteString(is.FuncSignature.String())
	out.WriteString("\n")

	return out.String()
}

type ExpressionStatement struct {
	Token      token.Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode() {}
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

type InitAssignExpression struct {
	LeftExp Expression
	Type    string
	Value   Expression
}

func (ia *InitAssignExpression) expressionNode() {}
func (ia *InitAssignExpression) String() string {
	var out bytes.Buffer

	out.WriteString(ia.LeftExp.String())
	if ia.Type != "" {
		out.WriteString(" ")
		out.WriteString(ia.Type)
	}
	out.WriteString(" := ")

	if ia.Value != nil {
		out.WriteString(ia.Value.String())
	}

	return out.String()
}

type CallExpression struct {
	Token     token.Token // The '(' token
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode() {}
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
	Token      token.Token
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

type IfExpression struct {
	Condition Expression
	Body      *BlockStatement
}

func (ie *IfExpression) expressionNode() {}
func (ie *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("if ")
	out.WriteString(ie.Condition.String())
	out.WriteString(ie.Body.String())

	return out.String()
}

type InfixExpression struct {
	Token    token.Token // The operator token, e.g. +
	Left     Expression
	Operator string
	Right    Expression
}

func (oe *InfixExpression) expressionNode() {}
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

func (i *Identifier) expressionNode() {}
func (i *Identifier) String() string  { return i.Value }

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (i *IntegerLiteral) expressionNode() {}
func (i *IntegerLiteral) String() string  { return i.Token.Lit }

type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (f *FloatLiteral) expressionNode() {}
func (f *FloatLiteral) String() string  { return f.Token.Lit }

type String struct {
	Token token.Token
	Value string
}

func (s *String) expressionNode() {}
func (s *String) String() string  { return `"` + s.Value + `"` }
