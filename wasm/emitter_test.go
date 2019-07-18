package wasm

import (
	"github.com/drejca/shift/parser"
	"github.com/drejca/shift/print"
	"github.com/perlin-network/life/exec"
	"os"
	"strings"
	"testing"
)

func TestEmptyMain(t *testing.T) {
	input := `
fn main() {
}
`
	p := parser.New(strings.NewReader(input))
	program, parseErr := p.ParseProgram()
	if parseErr != nil {
		t.Fatal(parseErr.Error())
	}

	compiler := NewCompiler()
	wasmModule := compiler.CompileProgram(program)

	for _, err := range compiler.Errors() {
		t.Error(err)
	}

	emitter := NewEmitter()
	err := emitter.Emit(wasmModule)
	if err != nil {
		t.Error(err)
	}

	vm, err := exec.NewVirtualMachine(emitter.Bytes(), exec.VMConfig{}, &exec.NopResolver{}, nil)
	if err != nil {
		panic(err)
	}

	entryID, ok := vm.GetFunctionExport("main")
	if !ok {
		panic("entry function not found")
	}

	_, err = vm.Run(entryID)
	if err != nil {
		vm.PrintStackTrace()
		panic(err)
	}
}

func TestEmitter(t *testing.T) {
	input := `
fn Calc(a i32, b i32) : i32 {
	let c = 2
	c = c + a
	return add(a, b) + c
}

fn add(a i32, b i32) : i32 {
	return a + b
}
`
	p := parser.New(strings.NewReader(input))
	program, parseErr := p.ParseProgram()
	if parseErr != nil {
		t.Fatal(parseErr.Error())
	}

	compiler := NewCompiler()
	wasmModule := compiler.CompileProgram(program)

	for _, err := range compiler.Errors() {
		t.Error(err)
	}

	emitter := NewEmitter()
	err := emitter.Emit(wasmModule)
	if err != nil {
		t.Error(err)
	}

	vm, err := exec.NewVirtualMachine(emitter.Bytes(), exec.VMConfig{}, &exec.NopResolver{}, nil)
	if err != nil {
		panic(err)
	}

	entryID, ok := vm.GetFunctionExport("Calc")
	if !ok {
		panic("entry function not found")
	}

	ret, err := vm.Run(entryID, int64(5), int64(7))
	if err != nil {
		vm.PrintStackTrace()
		panic(err)
	}

	expect := int64(19)
	if ret != expect {
		t.Errorf("expected %d but got %d", expect, ret)
	}
}

func TestCompilingFromFile(t *testing.T) {
	filename := "../testprogram/main.sf"

	file, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}

	p := parser.New(file)
	program, parseErr := p.ParseProgram()

	file.Close()

	if parseErr != nil {
		refile, err := os.Open(filename)
		if err != nil {
			t.Fatal(err)
		}

		printer := print.New(refile)
		t.Fatal(printer.PrintError(parseErr))
	}
	file.Close()

	compiler := NewCompiler()
	wasmModule := compiler.CompileProgram(program)

	for _, err := range compiler.Errors() {
		t.Fatal(err)
	}

	emitter := NewEmitter()
	err = emitter.Emit(wasmModule)
	if err != nil {
		t.Fatal(err)
	}
}
