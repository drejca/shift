package wasm

import "bytes"

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
	FUNC = 0x60

	// Variable access
	GET_LOCAL = 0x20
	SET_LOCAL = 0x21
	GET_GLOBAL = 0x23
	SET_GLOBAL = 0x24

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

type Module struct {
	sections []Section
}
func (m *Module) String() string {
	var out bytes.Buffer

	out.WriteString("(module \n")
	for _, section := range m.sections {
		out.WriteString(section.String())
	}
	out.WriteString("\n)")
	return out.String()
}

type TypeSection struct {
	count uint32
	entries []*FuncType
}
func (t *TypeSection) String() string {
	var out bytes.Buffer

	for _, funcType := range t.entries {
		out.WriteString(funcType.String())
	}
	return out.String()
}
func (t *TypeSection) sectionNode() {}

type FuncType struct {
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
func (f *FunctionSection) String() string {
	return ""
}
func (f *FunctionSection) sectionNode() {}

type CodeSection struct {
	count uint32
	bodies []*FunctionBody
}
func (c *CodeSection) String() string {
	var out bytes.Buffer

	for _, body := range c.bodies {
		for _, instruction := range body.code {
			out.WriteString(instruction.String())
		}
		out.WriteString(")")
	}

	return out.String()
}
func (c *CodeSection) sectionNode() {}

type FunctionBody struct {
	bodySize uint32
	localCount uint32
	locals []*LocalEntry
	code []Node
}
func (f *FunctionBody) String() string {
	return ""
}

type LocalEntry struct {
	count uint32
	valueType *ValueType
}
func (l *LocalEntry) String() string {
	return ""
}

type GetLocal struct {
	name string
	localIndex uint32
}
func (g *GetLocal) String() string {
	var out bytes.Buffer

	out.WriteString("\n		get_local $")
	out.WriteString(g.name)
	return out.String()
}

type SetGlobal struct {
	name string
	globalIndex uint32
}
func (s *SetGlobal) String() string {
	var out bytes.Buffer

	out.WriteString("\n 		set_global $")
	out.WriteString(s.name)
	return out.String()
}

type SetLocal struct {
	name string
	localIndex uint32
}
func (s *SetLocal) String() string {
	var out bytes.Buffer

	out.WriteString("\n 		set_local $")
	out.WriteString(s.name)
	return out.String()
}

type Add struct {
}
func (a *Add) String() string {
	var out bytes.Buffer

	out.WriteString("\n		i32.add")
	return out.String()
}

type Sub struct {
}
func (s *Sub) String() string {
	var out bytes.Buffer

	out.WriteString("\n		i32.sub")
	return out.String()
}

type ExportSection struct {
	count uint32
	entries []*ExportEntry
}
func (e *ExportSection) String() string {
	var out bytes.Buffer

	for _, exportEntry := range e.entries {
		out.WriteString(exportEntry.String())
	}

	return out.String()
}
func (e *ExportSection) sectionNode() {}

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
func (c *ConstInt) String() string {
	var out bytes.Buffer

	out.WriteString("	i32.const ")
	out.WriteString(string(c.value))
	return out.String()
}
