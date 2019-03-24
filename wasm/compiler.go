package wasm

import (
	"fmt"
	"github.com/drejca/shiftlang/ast"
)

type Compiler struct {
	module *Module
	currentBody *FunctionBody

	buf    []byte
	errors []error

	symbolTable *SymbolTable
	scopeIndex  uint32
}

func New() *Compiler {
	return &Compiler{symbolTable: NewSymbolTable()}
}

func (c *Compiler) Bytes() []byte {
	return c.buf
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		c.module = &Module{
			typeSection: &TypeSection{},
			functionSection: &FunctionSection{},
			exportSection: &ExportSection{},
			codeSection: &CodeSection{},
		}

		for _, stmt := range node.Statements {
			c.Compile(stmt)
		}
	case *ast.InfixExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.sumTypes(node.Left, node.Right)
		case "-":
			c.subtractTypes(node.Left, node.Right)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}
		c.loadSymbol(symbol)
	case *ast.IntegerLiteral:
	case *ast.Function:
		c.enterScope()

		c.module.typeSection.count++
		c.module.functionSection.count++
		c.module.codeSection.count++

		entry := &FuncType{}
		entry.name = node.Name

		c.currentBody = &FunctionBody{}

		for _, param := range node.InputParams {
			symbol := c.symbolTable.Define(param.Ident.Value, param.Type)

			c.module.functionSection.typesIdx = append(c.module.functionSection.typesIdx, symbol.Index)

			entry.paramTypes = append(entry.paramTypes, valueType(param))
			entry.paramCount++
		}

		if len(node.ReturnParams.Params) > 0 {
			for _, param := range node.ReturnParams.Params {
				resultType := &ResultType{typeName: param.Type}
				entry.resultType = resultType
				entry.resultCount++
			}
		}

		c.module.typeSection.entries = append(c.module.typeSection.entries, entry)

		if node.Name[0] >= 'A' && node.Name[0] <= 'Z' {
			exportEntry := &ExportEntry{field: node.Name, index: c.module.exportSection.count}
			c.module.exportSection.entries = append(c.module.exportSection.entries, exportEntry)
			c.module.exportSection.count++
		}

		c.Compile(node.Body)
		c.module.codeSection.bodies = append(c.module.codeSection.bodies, c.currentBody)

		c.leaveScope()
	case *ast.BlockStatement:
		for _, stmt := range node.Statements {
			c.Compile(stmt)
		}
	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return err
		}
	}
	return nil
}

func valueType(param *ast.Parameter) *ValueType {
	return &ValueType{
		name: param.Ident.Value,
		typeName: param.Type,
	}
}

func getType(typeName string) byte {
	switch typeName {
	case "i32":
		return TYPE_I32
	}
	return ZERO
}

func (c *Compiler) sumTypes(left ast.Node, right ast.Node) {
	add := &Add{}
	c.currentBody.code = append(c.currentBody.code, add)
}

func (c *Compiler) subtractTypes(left ast.Node, right ast.Node) {
	sub := &Sub{}
	c.currentBody.code = append(c.currentBody.code, sub)
}

func (c *Compiler) getType(left ast.Node, right ast.Node) (varType string) {
	leftType :=  c.resolveSymbol(left)
	rightType := c.resolveSymbol(right)

	if leftType != rightType {
		c.errors = append(c.errors, fmt.Errorf("can not subtract %s and %s", leftType, rightType))
		return ""
	}
	return leftType
}

func (c *Compiler) sumTypeOpCode(varType string) byte {
	switch varType {
	case "i32":
		return I32_ADD
	default:
		c.errors = append(c.errors, fmt.Errorf("can not sum user defined type %s", varType))
		return ZERO
	}
}

func (c *Compiler) subtractTypeOpCode(varType string) byte {
	switch varType {
	case "i32":
		return I32_SUB
	default:
		c.errors = append(c.errors, fmt.Errorf("can not subtract user defined type %s", varType))
		return ZERO
	}
}

func (c *Compiler) resolveSymbol(node ast.Node) (varType string) {
	ident, ok := node.(*ast.Identifier)
	if ok {
		symbol, ok := c.symbolTable.Resolve(ident.Value)
		if !ok {
			c.errors = append(c.errors, fmt.Errorf("undefined variable %s", ident.Value))
			return
		}
		c.loadSymbol(symbol)
		return symbol.Type
	}

	integer, ok := node.(*ast.IntegerLiteral)
	if ok {
		//c.emit(CONST_I32)
		//c.emit(byte(int32(integer.Value)))
		return integer.Token.Lit
	}

	c.errors = append(c.errors, fmt.Errorf("symbol could not be loaded for %s", node.String()))
	return
}

func (c *Compiler) enterScope() {
	c.scopeIndex++
	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

func (c *Compiler) leaveScope() {
	c.scopeIndex--
	c.symbolTable = c.symbolTable.Outer
}

func (c *Compiler) loadSymbol(s Symbol) {
	switch s.Scope {
	case GlobalScope:
		//c.emit(GET_GLOBAL)
	case LocalScope:
		getLocal := &GetLocal{name: s.Name, localIndex: s.Index}
		c.currentBody.code = append(c.currentBody.code, getLocal)
	}
}

func (c *Compiler) Module() *Module {
	return c.module
}

func (c *Compiler) Errors() []error {
	return c.errors
}
