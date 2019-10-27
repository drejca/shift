package wasm

import (
	"fmt"
	"reflect"

	"github.com/drejca/shift/ast"
)

type Compiler struct {
	module        *Module
	symbolTable   *SymbolTable
	functionBody  *FunctionBody
	typeIndex     uint32
	functionIndex uint32
	dataIndex     uint32
	dataOffset    int32

	errors []error
}

func NewCompiler() *Compiler {
	return &Compiler{
		symbolTable: NewSymbolTable(),
	}
}

func (c *Compiler) CompileProgram(program *ast.Program) *Module {
	c.module = &Module{
		typeSection:     &TypeSection{},
		importSection:   &ImportSection{},
		functionSection: &FunctionSection{},
		memorySection:   &MemorySection{},
		exportSection:   &ExportSection{},
		codeSection:     &CodeSection{},
		dataSection:     &DataSection{},
	}

	for _, stmt := range program.Statements {
		switch stmt := stmt.(type) {
		case *ast.Function:
			funcType := c.compileFunctionSignature(stmt.Signature)

			c.appendFunction(funcType)

			if (funcType.name[0] >= 'A' && funcType.name[0] <= 'Z') || funcType.name == "main" {
				funcType.exported = true
				c.appendExportEntry(funcType)
			}
		case *ast.ImportStatement:
			funcType := c.compileFunctionSignature(stmt.FuncSignature)

			c.appendImport(funcType)
		}
	}

	for _, stmt := range program.Statements {
		function, ok := stmt.(*ast.Function)
		if ok {
			funcBody := c.compileFunctionBody(function)
			c.appendCodeSection(funcBody)
		}
	}

	if c.module.dataSection.count > 0 {
		c.module.memorySection.count = 1
		memoryType := MemoryType{
			flags:         uint32(0),
			initialLength: uint32(1),
		}
		c.module.memorySection.entries = append(c.module.memorySection.entries, &memoryType)
	}

	return c.module
}

func (c *Compiler) compileFunctionSignature(functionSignature *ast.FunctionSignature) *FuncType {
	funcType := &FuncType{
		name: functionSignature.Name,
	}

	for _, param := range functionSignature.InputParams {
		valueTypes := c.compileFuncInputParam(param)

		for _, valueType := range valueTypes {
			funcType.paramTypes = append(funcType.paramTypes, valueType)
			funcType.paramCount++
		}
	}

	if len(functionSignature.ReturnParams) > 1 {
		c.errors = append(c.errors, fmt.Errorf("fn %s(...) : (...) multiple return types is not implemented", functionSignature.Name))
	}

	for _, param := range functionSignature.ReturnParams {
		resultType := &ResultType{typeName: param.Type}
		funcType.resultType = resultType
		funcType.resultCount = 1
	}

	if foundFuncType, found := c.findFunctionType(funcType.paramTypes, funcType.resultType); found {
		funcType.typeIndex = foundFuncType.typeIndex
		funcType.functionIndex = c.functionIndex
	} else {
		funcType.typeIndex = c.typeIndex
		funcType.functionIndex = c.functionIndex
		c.appendType(funcType)
	}

	return funcType
}

func (c *Compiler) compileFuncInputParam(param *ast.Parameter) []*ValueType {
	switch param.Type {
	case "string":
		offset := &ValueType{name: "offset", typeName: "i32"}
		strLength := &ValueType{name: "length", typeName: "i32"}
		return []*ValueType{offset, strLength}
	default:
		valueType := &ValueType{name: param.Ident.Value, typeName: param.Type}
		return []*ValueType{valueType}
	}
}

func (c *Compiler) findFunctionType(paramTypes []*ValueType, resultType *ResultType) (funcType *FuncType, found bool) {
	for _, typeEntry := range c.module.typeSection.entries {
		switch node := typeEntry.(type) {
		case *FuncType:
			if c.matchParams(node.paramTypes, paramTypes) && c.matchResult(node.resultType, resultType) {
				return node, true
			}
		}
	}
	return nil, false
}

func (c *Compiler) matchParams(paramTypesA []*ValueType, paramTypesB []*ValueType) (isMatch bool) {
	if len(paramTypesA) != len(paramTypesB) {
		return false
	}

	for i, paramType := range paramTypesA {
		if paramType.typeName != paramTypesB[i].typeName {
			return false
		}
	}
	return true
}

func (c *Compiler) matchResult(resultTypeA *ResultType, resultTypeB *ResultType) (isMatch bool) {
	if resultTypeA == nil && resultTypeB == nil {
		return true
	}
	if resultTypeA == nil || resultTypeB == nil {
		return false
	}
	if resultTypeA.typeName == resultTypeB.typeName {
		return true
	}
	return false
}

func (c *Compiler) compileFunctionBody(function *ast.Function) *FunctionBody {
	c.functionBody = &FunctionBody{}

	funcType, found := c.getFunctionType(function.Signature.Name)
	if !found {
		c.errors = append(c.errors, fmt.Errorf("function type for %s not found", function.Signature.Name))
		return nil
	}

	c.functionBody.funcName = funcType.name

	c.enterScope()

	for _, param := range function.Signature.InputParams {
		c.symbolTable.Define(param.Ident.Value, param.Type)
	}

	operations := c.compileBody(function.Body)
	c.functionBody.code = append(c.functionBody.code, operations...)

	c.leaveScope()
	return c.functionBody
}

func (c *Compiler) compileBody(body *ast.BlockStatement) []Operation {
	var operations []Operation

	for _, stmt := range body.Statements {
		ops := c.compileExpression(stmt)
		operations = append(operations, ops...)
	}
	return operations
}

func (c *Compiler) compileExpression(node ast.Node) []Operation {
	switch node := node.(type) {
	case *ast.InfixExpression:
		return c.compileInfixExpression(node)
	case *ast.ReturnStatement:
		return c.compileExpression(node.ReturnValue)
	case *ast.InitAssignExpression:
		return c.compileInitAssignExpression(node)
	case *ast.ExpressionStatement:
		return c.compileExpression(node.Expression)
	case *ast.CallExpression:
		return c.compileCallExpression(node)
	case *ast.IfExpression:
		return c.compileIfExpression(node)
	case *ast.AssignmentExpression:
		return c.compileAssignmentExpression(node)
	case *ast.Identifier:
		return c.compileIdentifier(node)
	case *ast.IntegerLiteral:
		constInt := &ConstInt{value: node.Value, typeName: "i32"}
		return []Operation{constInt}
	case *ast.String:
		c.addData([]byte(node.Value))

		offset := &ConstInt{value: 0, typeName: "i32"}
		strLength := &ConstInt{value: int64(len(node.Value)), typeName: "i32"}
		return []Operation{offset, strLength}
	}
	c.handleError(fmt.Errorf("unknown type %s", reflect.TypeOf(node)))
	return []Operation{}
}

func (c *Compiler) compileInitAssignExpression(exp *ast.InitAssignExpression) []Operation {
	var operations []Operation

	symbol := c.symbolTable.Define(exp.LeftExp.String(), c.inferType(exp.Value))

	expressionOps := c.compileExpression(exp.Value)
	operations = append(operations, expressionOps...)

	if symbol.Scope == GlobalScope {
		setGlobal := &SetGlobal{name: symbol.Name, globalIndex: symbol.Index}
		operations = append(operations, setGlobal)
	} else {
		c.functionBody.localCount++
		localEntry := &LocalEntry{count: 1, valueType: &ValueType{name: symbol.Name, typeName: symbol.Type}}
		c.functionBody.locals = append(c.functionBody.locals, localEntry)

		setLocal := &SetLocal{name: symbol.Name, localIndex: symbol.Index}
		operations = append(operations, setLocal)
	}
	return operations
}

func (c *Compiler) compileCallExpression(callExpression *ast.CallExpression) []Operation {
	var operations []Operation

	funcName := callExpression.Function.String()

	funcType, found := c.getFunctionType(funcName)
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

func (c *Compiler) compileIfExpression(ifExpression *ast.IfExpression) []Operation {
	var operations []Operation

	ifOp := &If{
		conditionOps: c.compileExpression(ifExpression.Condition),
		thenOps:      c.compileBody(ifExpression.Body),
	}

	operations = append(operations, ifOp)
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
	case "!=":
		operation, err := notEqual(infixExpression.Left, infixExpression.Right)
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

func (c *Compiler) getFunctionType(funcName string) (funcType *FuncType, found bool) {
	for _, typeEntry := range c.module.functionSection.entries {
		switch node := typeEntry.(type) {
		case *FuncType:
			if node.name == funcName {
				return node, true
			}
		}
	}
	for _, importEntry := range c.module.importSection.entries {
		switch node := importEntry.kind.(type) {
		case *FuncType:
			if node.name == funcName {
				return node, true
			}
		}
	}
	return nil, false
}

func (c *Compiler) appendExportEntry(funcType *FuncType) {
	exportEntry := &ExportEntry{index: funcType.typeIndex, field: funcType.name}

	c.module.exportSection.entries = append(c.module.exportSection.entries, exportEntry)
	c.module.exportSection.count++
}

func (c *Compiler) appendImport(externalKind Type) {
	importEntry := &ImportEntry{moduleName: "env", kind: externalKind}

	switch node := externalKind.(type) {
	case *FuncType:
		importEntry.fieldName = node.name
	}

	c.module.importSection.entries = append(c.module.importSection.entries, importEntry)
	c.module.importSection.count++
	c.functionIndex++
}

func (c *Compiler) appendType(funcType *FuncType) {
	c.module.typeSection.entries = append(c.module.typeSection.entries, funcType)
	c.module.typeSection.count++
	c.typeIndex++
}

func (c *Compiler) appendFunction(funcType *FuncType) {
	c.module.functionSection.count++
	c.module.functionSection.entries = append(c.module.functionSection.entries, funcType)
	c.functionIndex++
}

func (c *Compiler) appendCodeSection(funcBody *FunctionBody) {
	c.module.codeSection.bodies = append(c.module.codeSection.bodies, funcBody)
	c.module.codeSection.count++
}

func (c *Compiler) addData(data []byte) {
	dataSegment := DataSegment{
		index:  c.dataIndex,
		offset: c.dataOffset,
		size:   uint32(len(data)),
		data:   data,
	}
	c.module.dataSection.entries = append(c.module.dataSection.entries, &dataSegment)
	c.module.dataSection.count++
	c.dataIndex++
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
}

func (c *Compiler) inferType(expression ast.Expression) string {
	switch node := expression.(type) {
	case *ast.CallExpression:
		funcType, found := c.getFunctionType(node.Function.String())
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

func notEqual(left ast.Node, right ast.Node) (Operation, error) {
	return &NotEqual{}, nil
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
