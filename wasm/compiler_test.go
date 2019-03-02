package wasm

import (
	"github.com/perlin-network/life/exec"
	"github.com/drejca/shiftlang/parser"
	"strings"
	"testing"
)

func TestFunctionWithVm(t *testing.T) {
	input := `
fn Add(a i32, b i32) : i32 {
	return a + b;
}
`
	p := parser.New(strings.NewReader(input))
	program := p.Parse()

	compiler := New()
	compiler.Compile(program)

	vm, err := exec.NewVirtualMachine(compiler.Bytes(), exec.VMConfig{}, &exec.NopResolver{}, nil)
	if err != nil {
		panic(err)
	}

	entryID, ok := vm.GetFunctionExport("Add")
	if !ok {
		panic("entry function not found")
	}

	ret, err := vm.Run(entryID, int64(5), int64(7))
	if err != nil {
		vm.PrintStackTrace()
		panic(err)
	}

	expect := int64(12)
	if ret != expect {
		t.Errorf("expected %d but got %d", expect, ret)
	}
}
