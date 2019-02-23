package wasm

import (
	"bytes"
	"github.com/drejca/shiftlang/ast"
	"strconv"
)

type Printer struct {
	program *ast.Program
	functionIdx int
	out bytes.Buffer
}

func NewPrinter(program *ast.Program) *Printer {
	return &Printer{program: program, functionIdx: 0}
}

func (p * Printer) Print() string {
	p.write("(module\n")
	for _, stmt := range p.program.Statements {
		switch stmt.(type) {
		case *ast.Function:
			fn, _ := stmt.(*ast.Function)
			p.printFunction(fn)
		}
	}
	p.write(")")

	return p.out.String()
}
func (p * Printer) printFunction(function *ast.Function) {
	p.write("\t(type $t")
	p.write(strconv.Itoa(p.functionIdx))
	p.write(" (func")
	if function.ReturnParams != nil {
		p.printReturnParamGroup(function.ReturnParams)
	}
	p.write("))\n")
	p.write("\t(func $")
	p.write(function.Name)
	p.write(` (export "`)
	p.write(function.Name)
	p.write(`")`)

	p.write(" (type $t")
	p.write(strconv.Itoa(p.functionIdx))
	p.write(")")

	p.printReturnParamGroup(function.ReturnParams)
	p.write("\n")

	p.write("\t\ti32.const 5)")
}
func (p * Printer) printReturnParamGroup(paramGroup *ast.ReturnParamGroup) {
	p.write(" (result")
	for _, param := range paramGroup.Params {
		p.write(" ")
		p.write(param.Type)
	}
	p.write(")")
}
func (p * Printer) printParameter(param *ast.Parameter) {
	if param.Ident != nil {
		p.write(param.Ident.String())
		p.write(" ")
	}
	p.write(param.Type)
}

func (p * Printer) write(s string) {
	p.out.WriteString(s)
}
