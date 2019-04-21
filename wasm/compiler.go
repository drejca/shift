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
			c.module.typeSection = c.typeSection
		}
		if c.functionSection.count > 0 {
			c.module.sections = append(c.module.sections, c.functionSection)
		}
		if c.codeSection.count > 0 {
			c.module.codeSection = c.codeSection
		}
		if c.exportSection.count > 0 {
			c.module.sections = append(c.module.sections, c.exportSection)
		}
		return c.module, nil
	case *ast.InfixExpression:
		operation, err := c.Compile(node.Left)
		if err != nil {
			return nil, err
		}
		if operation != nil {
			c.symbolTable.scopeBody.code = append(c.symbolTable.scopeBody.code, operation)
		}

		operation, err = c.Compile(node.Right)
		if err != nil {
			return nil, err
		}
		if operation != nil {
			c.symbolTable.scopeBody.code = append(c.symbolTable.scopeBody.code, operation)
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
		symbol := c.symbolTable.Define(node.Name, "func")

		c.enterScope()

		c.typeSection.count++
		c.functionSection.count++
		c.codeSection.count++

		entry := &FuncType{
			functionIndex: symbol.Index,
			name: node.Name,
		}

		c.functionSection.typesIdx = append(c.functionSection.typesIdx, symbol.Index)

		for _, param := range node.InputParams {
			c.symbolTable.Define(param.Ident.Value, param.Type)

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
			exportEntry := &ExportEntry{field: node.Name, index: symbol.Index}
			c.exportSection.entries = append(c.exportSection.entries, exportEntry)
			c.exportSection.count++
		}

		c.symbolTable.scopeBody.functionIndex = symbol.Index

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
		operation, err := c.Compile(node.Value)
		if err != nil {
			return nil, err
		}
		c.symbolTable.scopeBody.code = append(c.symbolTable.scopeBody.code, operation)

		if symbol.Scope == GlobalScope {
			setGlobal  := &SetGlobal{name: symbol.Name, globalIndex: symbol.Index}
			c.addOperation(setGlobal)

			return setGlobal, nil
		} else {
			c.symbolTable.scopeBody.localCount++
			localEntry := &LocalEntry{count: 1, valueType: &ValueType{name: symbol.Name, typeName: symbol.Type}}
			c.symbolTable.scopeBody.locals = append(c.symbolTable.scopeBody.locals, localEntry)

			setLocal := &SetLocal{name: symbol.Name, localIndex: symbol.Index}
			c.addOperation(setLocal)

			return setLocal, nil
        }
	case *ast.CallExpression:
		symbol, ok := c.FunctionSymbol(node.Function)
		if !ok {
			return nil, fmt.Errorf("function symbol not found")
		}

		call := &Call{functionIndex: symbol.Index, name: node.Function.String()}

		for _, arg := range node.Arguments {
			expression, err := c.Compile(arg)
			if err != nil {
				return nil, err
			}

			operation, ok := expression.(Operation)
			if !ok {
				return nil, fmt.Errorf("its not operation")
			}
			call.arguments = append(call.arguments, operation)
		}
		return call, nil
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return nil, fmt.Errorf("undefined variable %s", node.Value)
		}
		operation := c.loadSymbol(symbol)
		return operation, nil
	case *ast.IntegerLiteral:
		constInt := &ConstInt{value: node.Value, typeName: "i32"}
		return constInt, nil
	}
	return nil, nil
}

func valueType(param *ast.Parameter) *ValueType {
	return &ValueType{
		name: param.Ident.Value,
		typeName: param.Type,
	}
}

func (c *Compiler) FunctionSymbol(callExpression ast.Node) (symbol Symbol, found bool) {
	switch node := callExpression.(type) {
	case *ast.Identifier:
		return c.symbolTable.Resolve(node.Value)
	default:
		return c.symbolTable.Resolve(callExpression.String())
	}
	return Symbol{}, false
}

func (c *Compiler) inferType(expression ast.Expression) string {
	return "i32"
}

func (c *Compiler) sumTypes(left ast.Node, right ast.Node) {
	c.addOperation(&Add{})
}

func (c *Compiler) subtractTypes(left ast.Node, right ast.Node) {
	c.addOperation(&Sub{})
}

func (c *Compiler) addOperation(operation Operation) {
	c.symbolTable.scopeBody.code = append(c.symbolTable.scopeBody.code, operation)
}

func (c *Compiler) enterScope() {
	c.scopeIndex++
	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

func (c *Compiler) leaveScope() {
	c.scopeIndex--
	c.symbolTable = c.symbolTable.Outer
}

func (c *Compiler) loadSymbol(s Symbol) Operation {
	switch s.Scope {
	case LocalScope:
		getLocal := &GetLocal{name: s.Name, localIndex: s.Index}
		return getLocal
	}
	return nil
}

func (c *Compiler) Module() *Module {
	return c.module
}

func (c *Compiler) Errors() []error {
	return c.errors
}
