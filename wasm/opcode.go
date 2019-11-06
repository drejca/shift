package wasm

import (
	"bytes"
	"strconv"
)

var (
	WASM_MAGIC_NUM = []byte{0x00, 0x61, 0x73, 0x6d}
	WASM_VERSION_1 = []byte{0x01, 0x00, 0x00, 0x00}
)

const (
	ZERO byte = 0x00

	// Function Bodies
	BODY_END = 0x0b

	// Module sections
	SECTION_TYPE   = 0x01
	SECTION_IMPORT = 0x02
	SECTION_FUNC   = 0x03
	SECTION_MEMORY = 0x05
	SECTION_EXPORT = 0x07
	SECTION_CODE   = 0x0a
	SECTION_DATA   = 0x0b

	// Language Types
	FUNC = 0x60

	// Value Types
	TYPE_I32   = 0x7f
	TYPE_I64   = 0x7e
	TYPE_EMPTY = 0x40

	// Variable access
	GET_LOCAL  = 0x20
	SET_LOCAL  = 0x21
	GET_GLOBAL = 0x23
	SET_GLOBAL = 0x24

	// Control flow operators
	NOP       = 0x01
	IF        = 0x04
	ELSE      = 0x05
	END_BLOCK = 0x0b

	// Call operators
	CALL = 0x10

	// Numeric operators
	I32_ADD       = 0x6a
	I32_SUB       = 0x6b
	I32_MUL       = 0x6c
	I32_NOT_EQUAL = 0x47

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

type Type interface {
	Node
	typeNode()
	TypeIndex() uint32
}

type Operation interface {
	Node
	operationNode()
}

type Module struct {
	typeSection     *TypeSection
	importSection   *ImportSection
	functionSection *FunctionSection
	memorySection   *MemorySection
	exportSection   *ExportSection
	codeSection     *CodeSection
	dataSection     *DataSection
}

func (m *Module) String() string {
	var out bytes.Buffer
	out.WriteString("(module ")

	if m.typeSection != nil {
		out.WriteString(m.typeSection.String())
	}
	if m.importSection != nil {
		out.WriteString(m.importSection.String())
	}

	for _, typeEntry := range m.functionSection.entries {
		out.WriteString("\n	")
		switch node := typeEntry.(type) {
		case *FuncType:
			out.WriteString(printFunctionSignature(node))
			out.WriteString(printFunctionCode(m.codeSection, node.name))
		}
	}

	if m.memorySection != nil {
		out.WriteString(m.memorySection.String())
	}
	if m.dataSection != nil {
		out.WriteString(m.dataSection.String())
	}

	out.WriteString("\n)")
	return out.String()
}

type TypeSection struct {
	count   uint32
	entries []Type
}

func (t *TypeSection) sectionNode() {}
func (t *TypeSection) String() string {
	var out bytes.Buffer
	for i, typeEntry := range t.entries {
		out.WriteString("\n")
		out.WriteString("	(type $t")
		out.WriteString(strconv.Itoa(i))
		out.WriteString(" ")
		out.WriteString(typeEntry.String())
		out.WriteString(")")
	}
	return out.String()
}
func (t *TypeSection) findByIdx(typeIdx uint32) (typeEntry Type, found bool) {
	for _, typeEntry := range t.entries {
		if typeEntry.TypeIndex() == typeIdx {
			return typeEntry, true
		}
	}
	return nil, false
}

type FuncType struct {
	functionIndex uint32
	typeIndex     uint32
	name          string
	exported      bool
	paramCount    uint32
	paramTypes    []*ValueType
	resultCount   uint32
	resultType    *ResultType
}

func (f *FuncType) typeNode() {}
func (f *FuncType) String() string {
	var out bytes.Buffer
	out.WriteString("(func")

	for _, paramType := range f.paramTypes {
		out.WriteString(" ")
		out.WriteString(paramType.String())
	}

	if f.resultCount > 0 {
		out.WriteString(" ")
		out.WriteString(f.resultType.String())
	}
	out.WriteString(")")
	return out.String()
}
func (f *FuncType) TypeIndex() uint32 {
	return f.typeIndex
}

type ValueType struct {
	name     string
	typeName string
}

func (v *ValueType) String() string {
	var out bytes.Buffer
	out.WriteString("(param ")
	out.WriteString(v.typeName)
	out.WriteString(")")
	return out.String()
}

type ResultType struct {
	typeName string
}

func (r *ResultType) String() string {
	var out bytes.Buffer
	out.WriteString("(result ")
	out.WriteString(r.typeName)
	out.WriteString(")")
	return out.String()
}

type ImportSection struct {
	count   uint
	entries []*ImportEntry
}

func (is *ImportSection) String() string {
	var out bytes.Buffer
	for _, entry := range is.entries {
		out.WriteString("\n	")
		out.WriteString(entry.String())
	}
	return out.String()
}

type ImportEntry struct {
	moduleName string
	fieldName  string
	kind       Type
}

func (ie *ImportEntry) String() string {
	var out bytes.Buffer
	out.WriteString("(import ")
	out.WriteString(`"`)
	out.WriteString(ie.moduleName)
	out.WriteString(`" "`)
	out.WriteString(ie.fieldName)
	out.WriteString(`" `)
	out.WriteString(printImportKind(ie.fieldName, ie.kind))
	out.WriteString(")")
	return out.String()
}

type FunctionSection struct {
	count   uint
	entries []Type
}

func (f *FunctionSection) sectionNode()   {}
func (f *FunctionSection) String() string { return "" }

type CodeSection struct {
	count  uint32
	bodies []*FunctionBody
}

func (c *CodeSection) sectionNode()   {}
func (c *CodeSection) String() string { return "" }

type FunctionBody struct {
	funcName   string
	bodySize   uint32
	localCount uint32
	locals     []*LocalEntry
	code       []Operation
}

func (f *FunctionBody) String() string {
	var out bytes.Buffer
	for _, local := range f.locals {
		out.WriteString(" ")
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
	name          string
	functionIndex uint32
	arguments     []Operation
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

type If struct {
	conditionOps []Operation
	thenOps      []Operation
}

func (i *If) operationNode() {}
func (i *If) String() string {
	var out bytes.Buffer
	out.WriteString("(if \n")
	for _, op := range i.conditionOps {
		out.WriteString("	")
		out.WriteString(op.String())
	}
	out.WriteString("	(then \n")
	for _, op := range i.thenOps {
		out.WriteString("		")
		out.WriteString(op.String())
	}
	out.WriteString("	\n)")
	out.WriteString("\n)")
	return out.String()
}

type LocalEntry struct {
	count     uint32
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
	name       string
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
	name        string
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
	name       string
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

type Multiply struct {
}

func (m *Multiply) operationNode() {}
func (m *Multiply) String() string {
	var out bytes.Buffer
	out.WriteString("i32.mul")
	return out.String()
}

type NotEqual struct {
}

func (n *NotEqual) operationNode() {}
func (n *NotEqual) String() string {
	var out bytes.Buffer
	out.WriteString("i32.ne")
	return out.String()
}

type MemorySection struct {
	count   uint32
	entries []*MemoryType
}

func (ms *MemorySection) sectionNode() {}
func (ms *MemorySection) String() string {
	var out bytes.Buffer
	for _, memoryType := range ms.entries {
		out.WriteString("\n	")
		out.WriteString(memoryType.String())
	}
	return out.String()
}

type MemoryType struct {
	flags         uint32
	initialLength uint32
	maximum       uint32
}

func (mt *MemoryType) String() string {
	var out bytes.Buffer
	out.WriteString(`(memory $memory (export "memory") `)
	out.WriteString(strconv.Itoa(int(mt.flags)))
	out.WriteString(`)`)
	return out.String()
}

type ExportSection struct {
	count   uint32
	entries []*ExportEntry
}

func (e *ExportSection) sectionNode() {}
func (e *ExportSection) String() string {
	var out bytes.Buffer
	for _, exportEntry := range e.entries {
		out.WriteString("\n	")
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
	out.WriteString("(export \"")
	out.WriteString(e.field)
	out.WriteString(`" (func $`)
	out.WriteString(e.field)
	out.WriteString("))")
	return out.String()
}

type ConstInt struct {
	value    int64
	typeName string
}

func (c *ConstInt) operationNode() {}
func (c *ConstInt) String() string {
	var out bytes.Buffer
	out.WriteString("i32.const ")
	out.WriteString(strconv.FormatInt(c.value, 10))
	return out.String()
}

type DataSection struct {
	count   uint32
	entries []*DataSegment
}

func (ds *DataSection) sectionNode() {}
func (ds *DataSection) String() string {
	var out bytes.Buffer
	for _, dataSegment := range ds.entries {
		out.WriteString("\n	")
		out.WriteString(dataSegment.String())
	}
	return out.String()
}

type DataSegment struct {
	index  uint32
	offset int32
	size   uint32
	data   []byte
}

func (ds *DataSegment) String() string {
	var out bytes.Buffer
	out.WriteString("(data (i32.const ")
	out.WriteString(strconv.FormatInt(int64(ds.offset), 10))
	out.WriteString(`) "`)
	out.WriteString(string(ds.data))
	out.WriteString(`")`)
	return out.String()
}

func printFunctionSignature(typeKind Type) string {
	var out bytes.Buffer

	switch node := typeKind.(type) {
	case *FuncType:
		out.WriteString("(func $")
		out.WriteString(node.name)
		if node.exported {
			out.WriteString(` (export "`)
			out.WriteString(node.name)
			out.WriteString(`")`)
		}
		out.WriteString(" (type $t")
		out.WriteString(strconv.Itoa(int(node.typeIndex)))
		out.WriteString(")")

		for _, paramType := range node.paramTypes {
			out.WriteString(" (param $")
			out.WriteString(paramType.name)
			out.WriteString(" ")
			out.WriteString(paramType.typeName)
			out.WriteString(")")
		}

		if node.resultCount > 0 {
			out.WriteString(" ")
			out.WriteString(node.resultType.String())
		}
		return out.String()
	}
	return out.String()
}

func printFunctionCode(codeSection *CodeSection, funcName string) string {
	for _, body := range codeSection.bodies {
		if body.funcName == funcName {
			return body.String()
		}
	}
	return ""
}

func printImportKind(fieldName string, typeKind Type) string {
	var out bytes.Buffer

	switch node := typeKind.(type) {
	case *FuncType:
		out.WriteString("(func $")
		out.WriteString(fieldName)
		out.WriteString(" (type $t")
		out.WriteString(strconv.Itoa(int(node.functionIndex)))
		out.WriteString("))")
		return out.String()
	}
	return out.String()
}

type NoOp struct {
}

func (n *NoOp) operationNode() {}
func (n *NoOp) String() string {
	var out bytes.Buffer
	out.WriteString("nop")
	return out.String()
}
