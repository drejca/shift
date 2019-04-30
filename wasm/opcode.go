package wasm

import (
	"bytes"
	"strconv"
)

var (
	WASM_MAGIC_NUM = []byte{0x00, 0x61, 0x73, 0x6d}
	WASM_VERSION_1 = []byte{0x01, 0x00,0x00, 0x00}
)

const (
	ZERO byte = 0x00

	// Function Bodies
	BODY_END = 0x0b

	// Module sections
	SECTION_TYPE = 0x01
	SECTION_FUNC = 0x03
	SECTION_EXPORT = 0x07
	SECTION_CODE = 0x0a

	// Language Types
	TYPE_I32 = 0x7f
	TYPE_I64 = 0x7e
	FUNC = 0x60

	// Variable access
	GET_LOCAL = 0x20
	SET_LOCAL = 0x21
	GET_GLOBAL = 0x23
	SET_GLOBAL = 0x24

	// Control flow operators
	NOP = 0x01

	// Call operators
	CALL = 0x10

	// Numeric operators
	I32_ADD = 0x6a
	I32_SUB = 0x6b

	// external_kind kind for import/export
	EXT_KIND_FUNC = 0x00

	// Constants
	CONST_I32 = 0x41
)

type Node interface {
	String() string
}

type Section interface {
	Node
	sectionNode()
}

type Operation interface {
	Node
	operationNode()
}

type Module struct {
	typeSection *TypeSection
	functionSection *FunctionSection
	exportSection *ExportSection
	codeSection *CodeSection
}
func (m *Module) String() string {
	var out bytes.Buffer
	out.WriteString("(module \n")

	for i, funcType := range m.typeSection.entries {
		if i > 0 {
			out.WriteString("\n")
		}
		out.WriteString(funcType.String())
		out.WriteString(printFunctionCode(m.codeSection, funcType.functionIndex))
	}
	out.WriteString(m.exportSection.String())

	out.WriteString("\n)")
	return out.String()
}

type TypeSection struct {
	count uint32
	entries []*FuncType
}
func (t *TypeSection) sectionNode() {}
func (t *TypeSection) String() string { return ""}

type FuncType struct {
	functionIndex uint32
	name string
	paramCount uint32
	paramTypes []*ValueType
	resultCount uint32
	resultType *ResultType
}
func (f *FuncType) String() string {
	var out bytes.Buffer
	out.WriteString("	(func $")
	out.WriteString(f.name)
	out.WriteString(" ")

	for _, paramType := range f.paramTypes {
		out.WriteString(paramType.String())
	}

	if f.resultCount > 0 {
		out.WriteString(f.resultType.String())
	}
	return out.String()
}

type ValueType struct {
	name string
	typeName string
}
func (v *ValueType) String() string {
	var out bytes.Buffer
	out.WriteString("(param $")
	out.WriteString(v.name)
	out.WriteString(" ")
	out.WriteString(v.typeName)
	out.WriteString(") ")
	return out.String()
}

type ResultType struct {
	typeName string
}
func (r *ResultType) String() string {
	var out bytes.Buffer
	out.WriteString("(result ")
	out.WriteString(r.typeName)
	out.WriteString(") ")
	return out.String()
}

type FunctionSection struct {
	count uint
	typesIdx []uint32
}
func (f *FunctionSection) sectionNode() {}
func (f *FunctionSection) String() string { return "" }

type CodeSection struct {
	count uint32
	bodies []*FunctionBody
}
func (c *CodeSection) sectionNode() {}
func (c *CodeSection) String() string { return ""}

type FunctionBody struct {
	functionIndex uint32
	bodySize uint32
	localCount uint32
	locals []*LocalEntry
	code []Operation
}
func (f *FunctionBody) String() string {
	var out bytes.Buffer
	for _, local := range f.locals {
		out.WriteString(local.String())
	}

	for _, instruction := range f.code {
		out.WriteString("\n		")
		out.WriteString(instruction.String())
	}
	out.WriteString(")")
	return out.String()
}

type Call struct {
	name string
	functionIndex uint32
	arguments []Operation
}
func (c *Call) operationNode() {}
func (c *Call) String() string {
	var out bytes.Buffer
	out.WriteString("(call $")
	out.WriteString(c.name)

	for _, arg := range c.arguments {
		out.WriteString(" (")
		out.WriteString(arg.String())
		out.WriteString(")")
	}
	out.WriteString(")")
	return out.String()
}

type LocalEntry struct {
	count uint32
	valueType *ValueType
}
func (l *LocalEntry) operationNode() {}
func (l *LocalEntry) String() string {
	var out bytes.Buffer
	out.WriteString("(local $")
	out.WriteString(l.valueType.name)
	out.WriteString(" ")
	out.WriteString(l.valueType.typeName)
	out.WriteString(")")
	return out.String()
}

type GetLocal struct {
	name string
	localIndex uint32
}
func (g *GetLocal) operationNode() {}
func (g *GetLocal) String() string {
	var out bytes.Buffer
	out.WriteString("get_local $")
	out.WriteString(g.name)
	return out.String()
}

type SetGlobal struct {
	name string
	globalIndex uint32
}
func (s *SetGlobal) operationNode() {}
func (s *SetGlobal) String() string {
	var out bytes.Buffer
	out.WriteString("set_global $")
	out.WriteString(s.name)
	return out.String()
}

type SetLocal struct {
	name string
	localIndex uint32
}
func (s *SetLocal) operationNode() {}
func (s *SetLocal) String() string {
	var out bytes.Buffer
	out.WriteString("set_local $")
	out.WriteString(s.name)
	return out.String()
}

type Add struct {
}
func (a *Add) operationNode() {}
func (a *Add) String() string {
	var out bytes.Buffer
	out.WriteString("i32.add")
	return out.String()
}

type Sub struct {
}
func (s *Sub) operationNode() {}
func (s *Sub) String() string {
	var out bytes.Buffer
	out.WriteString("i32.sub")
	return out.String()
}

type ExportSection struct {
	count uint32
	entries []*ExportEntry
}
func (e *ExportSection) sectionNode() {}
func (e *ExportSection) String() string {
	var out bytes.Buffer
	for _, exportEntry := range e.entries {
		out.WriteString(exportEntry.String())
	}
	return out.String()
}

type ExportEntry struct {
	field string
	index uint32
}
func (e *ExportEntry) String() string {
	var out bytes.Buffer
	out.WriteString("\n	(export \"")
	out.WriteString(e.field)
	out.WriteString(`" (func $`)
	out.WriteString(e.field)
	out.WriteString("))")
	return out.String()
}

type ConstInt struct {
	value int64
	typeName string
}
func (c *ConstInt) operationNode() {}
func (c *ConstInt) String() string {
	var out bytes.Buffer
	out.WriteString("i32.const ")
	out.WriteString(strconv.FormatInt(c.value, 10))
	return out.String()
}

func printFunctionCode(codeSection *CodeSection, functionIndex uint32) string {
	for _, body := range codeSection.bodies {
		if body.functionIndex == functionIndex {
			return body.String()
		}
	}
	return ""
}

type NoOp struct {
}
func (n *NoOp) operationNode() {}
func (n *NoOp) String() string {
	var out bytes.Buffer
	out.WriteString("nop")
	return out.String()
}
