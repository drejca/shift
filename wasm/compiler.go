package wasm

import (
	"github.com/drejca/shiftlang/ast"
)

type Compiler struct {
	program *ast.Program
	buf []byte
}

func New(program *ast.Program) *Compiler {
	return &Compiler{program: program}
}

func (c *Compiler) Bytes() []byte {
	return c.buf
}

func (c *Compiler) Compile() {
	c.emit(WASM_MAGIC_NUM...)
	c.emit(WASM_VERSION_1...)

	for _, stmt := range c.program.Statements {
		switch stmt.(type) {
		case *ast.Function:
			fn, _ := stmt.(*ast.Function)

			c.typeSectionBytes(fn)
			c.functionSectionBytes(fn)
			if fn.Name[0] >= 'A' && fn.Name[0] <= 'Z' {
				c.exportSectionBytes(fn)
			}
			c.codeSectionBytes(fn)
		}
	}
}

func (c *Compiler) typeSectionBytes(f *ast.Function) {
	c.emit(SECTION_TYPE)
	sectionSize := c.emit(ZERO)
	c.emit(byte(1))
	c.emit(FUNC)
	c.emit(ZERO)

	if len(f.ReturnParams.Params) > 0 {
		c.emit(byte(len(f.ReturnParams.Params)))
		c.paramsTypes(f.ReturnParams.Params)
	}

	c.fixup(sectionSize, byte(5))
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
	sectionSize := c.emit(ZERO)
	c.emit(byte(1))
	c.emit(byte(0))

	c.fixup(sectionSize, byte(2))
}

func (c *Compiler) exportSectionBytes(f *ast.Function) {
	c.emit(SECTION_EXPORT)
	sectionSize := c.emit(ZERO)
	c.emit(byte(uint32(1)))
	c.emit(byte(len(f.Name)))
	c.emit([]byte(f.Name)...)
	c.emit(EXT_KIND_FUNC)
	c.emit(ZERO)

	c.fixup(sectionSize, byte(7))
}

func (c *Compiler) codeSectionBytes(f *ast.Function) {
	c.emit(SECTION_CODE)
	sectionSize := c.emit(ZERO)
	c.emit(byte(1))
	bodySize := c.emit(ZERO)
	c.emit(ZERO) // number of local declarations
	c.emit(CONST_I32)
	c.emit(byte(int32(5)))
	c.emit(BODY_END)

	c.fixup(sectionSize, byte(6))
	c.fixup(bodySize, byte(4))
}

func (c *Compiler) emit(bytes ...byte) (pos int) {
	pos = len(c.buf)
	c.buf = append(c.buf, bytes...)
	return pos
}

func (c *Compiler) fixup(pos int, bytes ...byte) {
	for i, byte := range bytes {
		c.buf[pos + i] = byte
	}
}
