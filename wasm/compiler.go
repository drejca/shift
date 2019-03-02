package wasm

import (
	"fmt"
	"github.com/drejca/shiftlang/ast"
)

type Compiler struct {
	buf []byte
	errors []error

	symbolTable *SymbolTable
	scopeIndex uint32

	sectionId int
	sections []section
}

type section struct {
	id int
	pos int
	size int
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
		c.emit(WASM_MAGIC_NUM...)
		c.emit(WASM_VERSION_1...)

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
			c.emit(I32_ADD)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}
		c.loadSymbol(symbol)
	case *ast.Function:
		c.enterScope()

		for _, param := range node.InputParams {
			c.symbolTable.Define(param.Ident.Value)
		}

		c.typeSectionBytes(node)
		c.functionSectionBytes(node)
		if node.Name[0] >= 'A' && node.Name[0] <= 'Z' {
			c.exportSectionBytes(node)
		}
		c.emit(SECTION_CODE)
		sectionId := c.startSection()
		c.emit(byte(1)) // count(varuint32): count of function bodies to follow
		// bodies(function_body...): sequence of Function Bodies
		bodySectionId := c.startSection() // body_size(varuint32): size of function body to follow, in bytes
		c.emit(ZERO) // local_count(varuint32): number of local entries
		// none         locals(local_entry...): local variables

		c.Compile(node.Body)

		c.emit(BODY_END) // end(byte): 0x0b, indicating the end of the body

		c.endSection(bodySectionId)
		c.endSection(sectionId)

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

func (c *Compiler) typeSectionBytes(f *ast.Function) {
	c.emit(SECTION_TYPE)
	sectionId := c.startSection()

	c.emit(byte(1)) // count(varuint32): count of type entries to follow
	// entries(func_type...): repeated type entries as described above
	c.emit(FUNC) // form(varint7): the value for the 'func' type constructor as defined above

	c.emit(byte(len(f.InputParams))) // param_count(varuint32): the numbers of parameters to function
	//param_types(value_type...): the parameter types of the function
	c.paramsTypes(f.InputParams)

	if len(f.ReturnParams.Params) > 0 {
		c.emit(byte(len(f.ReturnParams.Params))) // return_count(varuint1): the number of results from the function
		c.paramsTypes(f.ReturnParams.Params) // return_type(value_type...?): the result type of the function (if return_count is 1)
	}

	c.endSection(sectionId)
}

func (c *Compiler) paramsTypes(params []*ast.Parameter) {
	for _, param := range params {
		c.paramsType(param)
	}
}

func (c *Compiler) paramsType(param *ast.Parameter) {
	switch param.Type {
	case "i32":
		c.emit(TYPE_I32)
	}
}

func (c *Compiler) functionSectionBytes(f *ast.Function) {
	c.emit(SECTION_FUNC)
	sectionId := c.startSection()

	c.emit(byte(1)) // count(varuint32): count of signature indices to follow
	c.emit(byte(0)) // types(varuint32...): sequence of indices into the type section 0,1,2,3...

	c.endSection(sectionId)
}

func (c *Compiler) exportSectionBytes(f *ast.Function) {
	c.emit(SECTION_EXPORT)
	sectionId := c.startSection()

	c.emit(byte(uint32(1))) // count(varuint32): count of export entries to follow
	c.emit(byte(len(f.Name))) // field_len(varuint32): length of field_str in bytes
	c.emit([]byte(f.Name)...) // field_str(bytes): valid UTF-8 byte sequence
	c.emit(EXT_KIND_FUNC) // kind(external_kind): the kind of definition being exported
	c.emit(ZERO) // index(varuint32): the index into the corresponding index space

	c.endSection(sectionId)
}

func (c *Compiler) emit(bytes ...byte) (pos int) {
	pos = len(c.buf)
	c.buf = append(c.buf, bytes...)
	for i := range c.sections {
		c.sections[i].size += len(bytes)
	}
	return pos
}

func (c *Compiler) startSection() (sectionId int) {
	c.sectionId++
	c.sections = append(c.sections, section{
		id: c.sectionId,
		pos: c.emit(ZERO),
		size: 0,
	})
	return c.sectionId
}

func (c *Compiler) endSection(sectionId int) {
	if section, found := c.findSection(sectionId); found {
		c.fixup(section.pos, byte(section.size))
		c.removeSection(sectionId)
	}
}

func (c *Compiler) findSection(sectionId int) (sec section, found bool) {
	for _, sec := range c.sections {
		if sec.id == sectionId {
			return sec, true
		}
	}
	return section{}, false
}

func (c *Compiler) removeSection(sectionId int) {
	for i, sec := range c.sections {
		if sec.id == sectionId {
			c.sections = append(c.sections[:i], c.sections[i+1:]...)
			return
		}
	}
}

func (c *Compiler) fixup(pos int, bytes ...byte) {
	for i, byte := range bytes {
		c.buf[pos + i] = byte
	}
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
		c.emit(GET_GLOBAL)
	case LocalScope:
		c.emit(GET_LOCAL)
	}
	c.emit(byte(s.Index))
}
