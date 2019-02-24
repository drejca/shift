package wasm

import (
	"github.com/drejca/shiftlang/ast"
	"strconv"
)

type Compiler struct {
	buf []byte
	errors []error
	sectionSizePos int
	sectionSize int
}

func New() *Compiler {
	return &Compiler{}
}

func (c *Compiler) Bytes() []byte {
	return c.buf
}

func (c *Compiler) Compile(node ast.Node) {
	switch node.(type) {
	case *ast.Program:
		c.emit(WASM_MAGIC_NUM...)
		c.emit(WASM_VERSION_1...)

		program, _ := node.(*ast.Program)
		for _, stmt := range program.Statements {
			c.Compile(stmt)
		}
	case *ast.Function:
		fn, _ := node.(*ast.Function)

		c.typeSectionBytes(fn)
		c.functionSectionBytes(fn)
		if fn.Name[0] >= 'A' && fn.Name[0] <= 'Z' {
			c.exportSectionBytes(fn)
		}
		c.emit(SECTION_CODE)
		sectionSize := c.emit(ZERO)
		c.emit(byte(1))
		bodySize := c.emit(ZERO)
		c.emit(ZERO) // number of local declarations

		c.Compile(fn.Body)

		c.emit(BODY_END)

		c.fixup(sectionSize, byte(6))
		c.fixup(bodySize, byte(4))
	case *ast.BlockStatement:
		block, _ := node.(*ast.BlockStatement)

		for _, stmt := range block.Statements {
			c.Compile(stmt)
		}
	case *ast.ReturnStatement:
		ret, _ := node.(*ast.ReturnStatement)

		val, err :=  strconv.ParseInt(ret.ReturnValue.String(), 10, 32)
		if err != nil {
			c.errors = append(c.errors, err)
		}

		c.emit(CONST_I32)
		c.emit(byte(int32(val)))
	}
}

func (c *Compiler) typeSectionBytes(f *ast.Function) {
	c.emit(SECTION_TYPE)

	c.startSection()

	c.emit(byte(1))
	c.emit(FUNC)
	c.emit(ZERO)

	if len(f.ReturnParams.Params) > 0 {
		c.emit(byte(len(f.ReturnParams.Params)))
		c.paramsTypes(f.ReturnParams.Params)
	}

	c.endSection()
}

func (c *Compiler) paramsTypes(params []*ast.Parameter) {
	for _, param := range params {
		switch param.Type {
		case "i32":
			c.emit(TYPE_I32)
		}
	}
}

func (c *Compiler) functionSectionBytes(f *ast.Function) {
	c.emit(SECTION_FUNC)

	c.startSection()

	c.emit(byte(1))
	c.emit(byte(0))

	c.endSection()
}

func (c *Compiler) exportSectionBytes(f *ast.Function) {
	c.emit(SECTION_EXPORT)

	c.startSection()

	c.emit(byte(uint32(1)))
	c.emit(byte(len(f.Name)))
	c.emit([]byte(f.Name)...)
	c.emit(EXT_KIND_FUNC)
	c.emit(ZERO)

	c.endSection()
}

func (c *Compiler) emit(bytes ...byte) (pos int) {
	pos = len(c.buf)
	c.buf = append(c.buf, bytes...)
	c.sectionSize += len(bytes)
	return pos
}

func (c *Compiler) startSection() {
	c.sectionSizePos = c.emit(ZERO)
	c.sectionSize = 0
}

func (c *Compiler) endSection() {
	c.fixup(c.sectionSizePos, byte(c.sectionSize))
}

func (c *Compiler) fixup(pos int, bytes ...byte) {
	for i, byte := range bytes {
		c.buf[pos + i] = byte
	}
}
