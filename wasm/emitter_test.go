package wasm

import (
	"github.com/drejca/shift/parser"
	"github.com/perlin-network/life/exec"
	"strings"
	"testing"
)

func TestEmitter(t *testing.T) {
	input := `
fn Calc(a i32, b i32) : i32 {
	let c = 2
	return add(a, b) + c
}

fn add(a i32, b i32) : i32 {
	return a + b
}
`
	p := parser.New(strings.NewReader(input))
	program := p.Parse()

	compiler := New()
	module := compiler.CompileProgram(program)

	for _, err := range compiler.Errors() {
		t.Error(err)
	}
	

	emitter := NewEmitter()
	err := emitter.Emit(module)
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

	expect := int64(14)
	if ret != expect {
		t.Errorf("expected %d but got %d", expect, ret)
	}
}
