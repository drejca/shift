package wasm

import (
	"fmt"
	"github.com/drejca/shiftlang/ast"
)

type Compiler struct {
	module *Module
	typeSection *TypeSection
	functionSection *FunctionSection
	exportSection *ExportSection
	codeSection *CodeSection

	buf    []byte
	errors []error

	symbolTable *SymbolTable
	scopeIndex  uint32
}

func New() *Compiler {
	return &Compiler{
		symbolTable: NewSymbolTable(),
		module: &Module{},
		typeSection: &TypeSection{},
		functionSection: &FunctionSection{},
		exportSection: &ExportSection{},
		codeSection: &CodeSection{},
	}
}

func (c *Compiler) Bytes() []byte {
	return c.buf
}

func (c *Compiler) Compile(node ast.Node) (Node, error) {
	switch node := node.(type) {
	case *ast.Program:
		for _, stmt := range node.Statements {
			c.Compile(stmt)
		}
		if c.typeSection.count > 0 {
			c.module.sections = append(c.module.sections, c.typeSection)
		}
		if c.functionSection.count > 0 {
			c.module.sections = append(c.module.sections, c.functionSection)
		}
		if c.codeSection.count > 0 {
			c.module.sections = append(c.module.sections, c.codeSection)
		}
		if c.exportSection.count > 0 {
			c.module.sections = append(c.module.sections, c.exportSection)
		}
		return c.module, nil
	case *ast.InfixExpression:
		_, err := c.Compile(node.Left)
		if err != nil {
			return nil, err
		}

		_, err = c.Compile(node.Right)
		if err != nil {
			return nil, err
		}

		switch node.Operator {
		case "+":
			c.sumTypes(node.Left, node.Right)
		case "-":
			c.subtractTypes(node.Left, node.Right)
		default:
			return nil, fmt.Errorf("unknown operator %s", node.Operator)
		}
	case *ast.Function:
		c.enterScope()

		c.typeSection.count++
		c.functionSection.count++
		c.codeSection.count++

		entry := &FuncType{}
		entry.name = node.Name

		for _, param := range node.InputParams {
			symbol := c.symbolTable.Define(param.Ident.Value, param.Type)

			c.functionSection.typesIdx = append(c.functionSection.typesIdx, symbol.Index)

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

		c.typeSection.entries = append(c.typeSection.entries, entry)

		if node.Name[0] >= 'A' && node.Name[0] <= 'Z' {
			if c.exportSection == nil {
				c.exportSection = &ExportSection{}
			}
			exportEntry := &ExportEntry{field: node.Name, index: c.exportSection.count}
			c.exportSection.entries = append(c.exportSection.entries, exportEntry)
			c.exportSection.count++
		}

		c.Compile(node.Body)
		c.codeSection.bodies = append(c.codeSection.bodies, c.symbolTable.scopeBody)

		c.leaveScope()
	case *ast.BlockStatement:
		for _, stmt := range node.Statements {
			c.Compile(stmt)
		}
	case *ast.ReturnStatement:
		expression, err := c.Compile(node.ReturnValue)
		if err != nil {
			return nil, err
		}
		return expression, nil
	case *ast.LetStatement:
		symbol := c.symbolTable.Define(node.Name.Value, c.inferType(node.Value))
		_, err := c.Compile(node.Value)
		if err != nil {
			return nil, err
		}

		if symbol.Scope == GlobalScope {
			setGlobal  := &SetGlobal{name: symbol.Name, globalIndex: symbol.Index}
			c.symbolTable.scopeBody.code = append(c.symbolTable.scopeBody.code, setGlobal)
			return setGlobal, nil
		} else {
			c.symbolTable.scopeBody.localCount++
			localEntry := &LocalEntry{count: 1, valueType: &ValueType{name: symbol.Name, typeName: symbol.Type}}
			c.symbolTable.scopeBody.locals = append(c.symbolTable.scopeBody.locals, localEntry)

			setLocal := &SetLocal{name: symbol.Name, localIndex: symbol.Index}
			c.symbolTable.scopeBody.code = append(c.symbolTable.scopeBody.code, setLocal)
			return setLocal, nil
        }
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return nil, fmt.Errorf("undefined variable %s", node.Value)
		}
		c.loadSymbol(symbol)
	case *ast.IntegerLiteral:
		constInt := &ConstInt{value: node.Value, typeName: "i32"}
		c.symbolTable.scopeBody.code = append(c.symbolTable.scopeBody.code, constInt)
	}
	return nil, nil
}

func valueType(param *ast.Parameter) *ValueType {
	return &ValueType{
		name: param.Ident.Value,
		typeName: param.Type,
	}
}

func (c *Compiler) inferType(expression ast.Expression) string {
	return "i32"
}

func (c *Compiler) sumTypes(left ast.Node, right ast.Node) {
	add := &Add{}
	c.symbolTable.scopeBody.code = append(c.symbolTable.scopeBody.code, add)
}

func (c *Compiler) subtractTypes(left ast.Node, right ast.Node) {
	sub := &Sub{}
	c.symbolTable.scopeBody.code = append(c.symbolTable.scopeBody.code, sub)
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
	case LocalScope:
		getLocal := &GetLocal{name: s.Name, localIndex: s.Index}
		c.symbolTable.scopeBody.code = append(c.symbolTable.scopeBody.code, getLocal)
	}
}

func (c *Compiler) Module() *Module {
	return c.module
}

func (c *Compiler) Errors() []error {
	return c.errors
}
