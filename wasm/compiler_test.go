package wasm

import (
	"github.com/drejca/shiftlang/assert"
	"github.com/drejca/shiftlang/life/exec"
	"github.com/drejca/shiftlang/parser"
	"strings"
	"testing"
)

func TestCompileFunction(t *testing.T) {
	input := `
fn Get() : i32 {
	return 5;
}
`
	expected := `
(module
	(type $t0 (func (result i32)))
	(func $Get (export "Get") (type $t0) (result i32)
		i32.const 5))`

	p := parser.New(strings.NewReader(input))
	program := p.Parse()

	printer := NewPrinter(program)

	err := assert.EqualString(expected, "\n" + printer.Print())
	if err != nil {
		t.Error(err)
	}
}

func TestFunctionWithVm(t *testing.T) {
	input := `
fn Get() : i32 {
	return 5;
}
`
	p := parser.New(strings.NewReader(input))
	program := p.Parse()

	compiler := New(program)
	compiler.Compile()

	vm, err := exec.NewVirtualMachine(compiler.Bytes(), exec.VMConfig{}, &exec.NopResolver{}, nil)
	if err != nil {
		panic(err)
	}

	entryID, ok := vm.GetFunctionExport("Get")
	if !ok {
		panic("entry function not found")
	}

	ret, err := vm.Run(entryID)
	if err != nil {
		panic(err)
	}
	if ret != 5 {
		t.Errorf("expected %d but got %d", 5, ret)
	}
}
