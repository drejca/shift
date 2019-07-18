package wasm

import (
	"fmt"
	"github.com/drejca/shift/ast"
	"reflect"
)

type Compiler struct {
	symbolTable *SymbolTable
	typeSection *TypeSection
	functionBody *FunctionBody

	errors []error
}

func NewCompiler() *Compiler {
	return &Compiler{
		symbolTable: NewSymbolTable(),
	}
}

func (c *Compiler) CompileProgram(program *ast.Program) *Module {
	module := &Module{
		functionSection: &FunctionSection{},
	}
	var functionIndex uint32 = 0

	for _, stmt := range program.Statements {
		function, ok := stmt.(*ast.Function)
		if ok {
			module.functionSection.count++
			module.functionSection.typesIdx = append(module.functionSection.typesIdx, functionIndex)

			funcType := c.compileFunctionSignature(function, functionIndex)

			if (funcType.name[0] >= 'A' && funcType.name[0] <= 'Z') || funcType.name == "main"  {
				appendExportEntry(module, funcType)
			}
			c.appendFunctionType(module, funcType)

			functionIndex++
		}
	}

	for _, stmt := range program.Statements {
		function, ok := stmt.(*ast.Function)
		if ok {
			funcBody := c.compileFunctionBody(function, module.typeSection)
			appendCodeSection(module, funcBody)
		}
	}

	return module
}

func (c *Compiler) compileFunctionSignature(function *ast.Function, functionIndex uint32) *FuncType {
	funcType := &FuncType{
		functionIndex: functionIndex,
		name: function.Name,
	}

	for _, param := range function.InputParams {
		valueType := &ValueType{name: param.Ident.Value, typeName: param.Type}
		funcType.paramTypes = append(funcType.paramTypes, valueType)
		funcType.paramCount++
	}

	if len(function.ReturnParams) > 1 {
		c.errors = append(c.errors, fmt.Errorf("fn %s(...) : (...) multiple return types is not implemented", function.Name))
	}

	for _, param := range function.ReturnParams {
		resultType := &ResultType{typeName: param.Type}
		funcType.resultType = resultType
		funcType.resultCount = 1
	}

	return funcType
}

func (c *Compiler) compileFunctionBody(function *ast.Function, typeSection *TypeSection) *FunctionBody {
	c.functionBody = &FunctionBody{}

	funcType, found := getFunctionType(typeSection, function.Name)
	if !found {
		c.errors = append(c.errors, fmt.Errorf("function type for %s not found", function.Name))
		return nil
	}
	c.functionBody.functionIndex = funcType.functionIndex

	c.enterScope()

	for _, param := range function.InputParams {
		c.symbolTable.Define(param.Ident.Value, param.Type)
	}

	for _, stmt := range function.Body.Statements {
		operations := c.compileExpression(stmt)
		c.functionBody.code = append(c.functionBody.code, operations...)
	}

	c.leaveScope()
	return c.functionBody
}

func (c *Compiler) compileExpression(node ast.Node) []Operation {
	switch node := node.(type) {
	case *ast.InfixExpression:
		return c.compileInfixExpression(node)
	case *ast.ReturnStatement:
		return c.compileExpression(node.ReturnValue)
	case *ast.LetStatement:
		return c.compileLetStatement(node, c.functionBody)
	case *ast.ExpressionStatement:
		return c.compileExpression(node.Expression)
	case *ast.CallExpression:
		return c.compileCallExpression(node, c.typeSection)
	case *ast.AssignmentExpression:
		return c.compileAssignmentExpression(node)
	case *ast.Identifier:
		return c.compileIdentifier(node)
	case *ast.IntegerLiteral:
		constInt := &ConstInt{value: node.Value, typeName: "i32"}
		return []Operation{constInt}
	}
	c.handleError(fmt.Errorf("unknown type %s", reflect.TypeOf(node)))
	return []Operation{}
}

func (c *Compiler) compileLetStatement(letStatement *ast.LetStatement, functionBody *FunctionBody) []Operation {
	var operations []Operation

	symbol := c.symbolTable.Define(letStatement.Name.Value, c.inferType(letStatement.Value))

	expressionOps := c.compileExpression(letStatement.Value)
	operations = append(operations, expressionOps...)

	if symbol.Scope == GlobalScope {
		setGlobal  := &SetGlobal{name: symbol.Name, globalIndex: symbol.Index}
		operations = append(operations, setGlobal)
	} else {
		functionBody.localCount++
		localEntry := &LocalEntry{count: 1, valueType: &ValueType{name: symbol.Name, typeName: symbol.Type}}
		functionBody.locals = append(functionBody.locals, localEntry)

		setLocal := &SetLocal{name: symbol.Name, localIndex: symbol.Index}
		operations = append(operations, setLocal)
	}
	return operations
}

func (c *Compiler) compileCallExpression(callExpression *ast.CallExpression, typeSection *TypeSection) []Operation {
	var operations []Operation

	funcName := callExpression.Function.String()

	funcType, found := getFunctionType(typeSection, funcName)
	if !found {
		c.errors = append(c.errors, fmt.Errorf("function type for %s not found", funcName))
		return nil
	}

	call := &Call{functionIndex: funcType.functionIndex, name: funcName}

	for _, arg := range callExpression.Arguments {
		operations := c.compileExpression(arg)
		call.arguments = append(call.arguments, operations...)
	}
	operations = append(operations, call)
	return operations
}

func (c *Compiler) compileAssignmentExpression(assignmentExpression *ast.AssignmentExpression) []Operation {
	var operations []Operation

	symbol, ok := c.symbolTable.Resolve(assignmentExpression.Identifier.String())
	if !ok {
		c.handleError(fmt.Errorf("variable %s is undefined", assignmentExpression.Identifier.String()))
		return operations
	}

	expressionOperations := c.compileExpression(assignmentExpression.Expression)
	operations = append(operations, expressionOperations...)

	setLocal := &SetLocal{name: symbol.Name, localIndex: symbol.Index}
	operations = append(operations, setLocal)

	return operations
}

func (c *Compiler) compileInfixExpression(infixExpression *ast.InfixExpression) []Operation {
	var operations []Operation

	expressionOperations := c.compileExpression(infixExpression.Left)
	operations = append(operations, expressionOperations...)

	expressionOperations = c.compileExpression(infixExpression.Right)
	operations = append(operations, expressionOperations...)

	switch infixExpression.Operator {
	case "+":
		operation, err := sumTypes(infixExpression.Left, infixExpression.Right)
		c.handleError(err)
		operations = append(operations, operation)
	case "-":
		operation, err := subtractTypes(infixExpression.Left, infixExpression.Right)
		c.handleError(err)
		operations = append(operations, operation)
	default:
		c.handleError(fmt.Errorf("unknown operator %s", infixExpression.Operator))
	}
	return operations
}

func (c *Compiler) compileIdentifier(identifier *ast.Identifier) []Operation {
	symbol, ok := c.symbolTable.Resolve(identifier.Value)
	if !ok {
		c.handleError(fmt.Errorf("undefined variable %s", identifier.Value))
		return []Operation{}
	}
	operation := loadSymbol(symbol)
	return []Operation{operation}
}

func getFunctionType(typeSection *TypeSection, funcName string) (funcType *FuncType, found bool) {
	for _, entry := range typeSection.entries {
		if entry.name == funcName {
			return entry, true
		}
	}
	return nil, false
}

func appendExportEntry(module *Module, funcType *FuncType) {
	if module.exportSection == nil {
		module.exportSection = &ExportSection{}
	}
	exportEntry := &ExportEntry{index: funcType.functionIndex, field: funcType.name}

	module.exportSection.entries = append(module.exportSection.entries, exportEntry)
	module.exportSection.count++
}

func (c *Compiler) appendFunctionType(module *Module, funcType *FuncType) {
	if module.typeSection == nil {
		module.typeSection = &TypeSection{}
		c.typeSection = module.typeSection
	}
	module.typeSection.entries = append(module.typeSection.entries, funcType)
	module.typeSection.count++
}

func appendCodeSection(module *Module, funcBody *FunctionBody) {
	if module.codeSection == nil {
		module.codeSection = &CodeSection{}
	}
	module.codeSection.bodies = append(module.codeSection.bodies, funcBody)
	module.codeSection.count++
}

func (c *Compiler) handleError(err error) {
	if err != nil {
		c.errors = append(c.errors, err)
	}
}

func (c *Compiler) functionSymbol(callExpression ast.Node) (symbol Symbol, found bool) {
	switch node := callExpression.(type) {
	case *ast.Identifier:
		return c.symbolTable.Resolve(node.Value)
	default:
		return c.symbolTable.Resolve(callExpression.String())
	}
	return Symbol{}, false
}

func (c *Compiler) inferType(expression ast.Expression) string {
	switch node := expression.(type) {
	case *ast.CallExpression:
		funcType, found := getFunctionType(c.typeSection, node.Token.Lit)
		if !found {
			c.errors = append(c.errors, fmt.Errorf("function type for %s not found", node.Token.Lit))
			return "unknown"
		}
		return funcType.resultType.typeName
	case *ast.IntegerLiteral:
		return "i32"
	}
	return "unknown"
}

func sumTypes(left ast.Node, right ast.Node) (Operation, error) {
	return &Add{}, nil
}

func subtractTypes(left ast.Node, right ast.Node) (Operation, error) {
	return &Sub{}, nil
}

func (c *Compiler) enterScope() {
	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

func (c *Compiler) leaveScope() {
	c.symbolTable = c.symbolTable.Outer
}

func loadSymbol(s Symbol) Operation {
	switch s.Scope {
	case LocalScope:
		getLocal := &GetLocal{name: s.Name, localIndex: s.Index}
		return getLocal
	}
	return nil
}

func (c *Compiler) Errors() []error {
	return c.errors
}
