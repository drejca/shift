package wasm

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/drejca/shift/parser"
	"github.com/drejca/shift/print"
	"github.com/perlin-network/life/exec"
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

type Resolver struct {
	t *testing.T
}

func (r *Resolver) ResolveFunc(module string, field string) exec.FunctionImport {
	switch module {
	case "env":
		switch field {
		case "assert":
			return func(vm *exec.VirtualMachine) int64 {
				expected := uint32(vm.GetCurrentFrame().Locals[0])
				actual := uint32(vm.GetCurrentFrame().Locals[1])
				if expected != actual {
					r.t.Errorf("expected %d got %d", expected, actual)
				}
				return 0
			}
		default:
			panic(fmt.Errorf("unknown import resolved: %s", field))
		}
	default:
		panic(fmt.Errorf("unknown module: %s", module))
	}
}

func (r *Resolver) ResolveGlobal(module, field string) int64 {
	panic("we're not resolving global variables for now")
}

func TestEmitter(t *testing.T) {
	input := `
import fn assert(expected i32, actual i32)

fn main() {
	let res = Calc(85, 25)
	let expected = 197
	
	if res != expected {
		assert(expected, res)
	}
}

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

	resolver := &Resolver{t: t}

	vm, err := exec.NewVirtualMachine(emitter.Bytes(), exec.VMConfig{}, resolver, nil)
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
