package wasm

import "fmt"

type Emmiter struct {
	buf       []byte
	sectionId int
	sections  []section
	errors []error
}

type section struct {
	id   int
	pos  int
	size int
}

func NewEmitter() *Emmiter {
	return &Emmiter{}
}

func (e *Emmiter) Emit(node Node) error {
	switch node := node.(type) {
	case *Module:
		e.emit(WASM_MAGIC_NUM...)
		e.emit(WASM_VERSION_1...)
		for _, section := range node.sections {
			e.Emit(section)
		}
	case *TypeSection:
		e.emit(SECTION_TYPE)
		sectionId := e.startSection()

		e.emit(byte(node.count))
		for _, funcType := range node.entries {
			e.Emit(funcType)
		}
		e.endSection(sectionId)
	case *ExportSection:
		e.emit(SECTION_EXPORT)
		sectionId := e.startSection()

		e.emit(byte(node.count))
		for _, exportEntry := range node.entries {
			e.Emit(exportEntry)
		}
		e.endSection(sectionId)
	case *FunctionSection:
		e.emit(SECTION_FUNC)
		sectionId := e.startSection()

		e.emit(byte(1))
		e.emit(byte(0))
		/*
		e.emit(byte(node.count))
		for _, typeIdx := range node.typesIdx {
			e.emit(byte(typeIdx))
		}
		*/
		e.endSection(sectionId)
	case *CodeSection:
		e.emit(SECTION_CODE)
		sectionId := e.startSection()

		e.emit(byte(node.count))
		for _, functionBody := range node.bodies {
			e.Emit(functionBody)
		}
		e.endSection(sectionId)
	case *FuncType:
		e.emit(FUNC)
		e.emit(byte(node.paramCount))
		for _, valueType := range node.paramTypes {
			e.Emit(valueType)
		}
		e.emit(byte(node.resultCount))
		e.Emit(node.resultType)
	case *ConstInt:
		e.emit(CONST_I32)
		e.emit(byte(node.value))
	case *ValueType:
		e.emit(e.typeOpCode(node.typeName))
	case *ResultType:
		e.emit(e.typeOpCode(node.typeName))
	case *ExportEntry:
		e.emit(byte(len(node.field)))
		e.emit([]byte(node.field)...)
		e.emit(EXT_KIND_FUNC)
		e.emit(byte(node.index))
	case *FunctionBody:
		sectionId := e.startSection()

		e.emit(byte(node.localCount))
		for _, localEntry := range node.locals {
			e.Emit(localEntry)
		}
		for _, node := range node.code {
			e.Emit(node)
		}
		e.emit(BODY_END)

		e.endSection(sectionId)
	case *LocalEntry:
		e.emit(byte(node.count))
		e.Emit(node.valueType)
	case *SetGlobal:
		e.emit(SET_GLOBAL)
		e.emit(byte(node.globalIndex))
	case *SetLocal:
		e.emit(SET_LOCAL)
		e.emit(byte(node.localIndex))
	case *GetLocal:
		e.emit(GET_LOCAL)
		e.emit(byte(node.localIndex))
	case *Add:
		e.emit(I32_ADD)
	case *Sub:
		e.emit(I32_SUB)
	}
	return nil
}

func (e *Emmiter) typeOpCode(typeName string) byte {
	switch typeName {
	case "i32":
		return TYPE_I32
	}
	e.errors = append(e.errors, fmt.Errorf("unknown type %q", typeName))
	return ZERO
}

func (e *Emmiter) startSection() (sectionId int) {
	e.sectionId++
	e.sections = append(e.sections, section{
		id:   e.sectionId,
		pos:  e.emit(ZERO),
		size: 0,
	})
	return e.sectionId
}

func (e *Emmiter) endSection(sectionId int) {
	if section, found := e.findSection(sectionId); found {
		e.fixup(section.pos, byte(section.size))
		e.removeSection(sectionId)
	}
}

func (e *Emmiter) findSection(sectionId int) (sec section, found bool) {
	for _, sec := range e.sections {
		if sec.id == sectionId {
			return sec, true
		}
	}
	return section{}, false
}

func (e *Emmiter) removeSection(sectionId int) {
	for i, sec := range e.sections {
		if sec.id == sectionId {
			e.sections = append(e.sections[:i], e.sections[i+1:]...)
			return
		}
	}
}

func (e *Emmiter) fixup(pos int, bytes ...byte) {
	for i, byte := range bytes {
		e.buf[pos+i] = byte
	}
}

func (e *Emmiter) Bytes() []byte {
	return e.buf
}

func (e *Emmiter) emit(bytes ...byte) (pos int) {
	pos = len(e.buf)
	e.buf = append(e.buf, bytes...)
	for i := range e.sections {
		e.sections[i].size += len(bytes)
	}
	return pos
}
