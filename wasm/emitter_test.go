package wasm_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/drejca/shift/parser"
	"github.com/drejca/shift/print"
	"github.com/drejca/shift/wasm"
	"github.com/perlin-network/life/exec"
)

type Resolver struct {
	t *testing.T
}

func (r *Resolver) ResolveFunc(module string, field string) exec.FunctionImport {
	switch module {
	case "env":
		switch field {
		case "error":
			return func(vm *exec.VirtualMachine) int64 {
				offset := uint32(vm.GetCurrentFrame().Locals[0])
				msgLength := uint32(vm.GetCurrentFrame().Locals[1])
				msg := string(vm.Memory[offset:msgLength])
				r.t.Error(msg)
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

func TestCompilingFromFile(t *testing.T) {
	testCases := []struct {
		name string
		file string
	}{
		{
			name: "empty main main.sf",
			file: "../testprogram/empty_main.sf",
		},
		{
			name: "main.sf Add function",
			file: "../testprogram/main.sf",
		},
		{
			name: "operators tests",
			file: "../testprogram/operators.sf",
		},
	}

	for _, tc := range testCases {

		file, err := os.Open(tc.file)
		if err != nil {
			t.Fatal(err)
		}

		p := parser.New(file)
		program, parseErr := p.ParseProgram()

		file.Close()

		if parseErr != nil {
			refile, err := os.Open(tc.file)
			if err != nil {
				t.Fatal(err)
			}

			printer := print.New(refile)
			t.Fatal(printer.PrintError(parseErr))
		}
		file.Close()

		compiler := wasm.NewCompiler()
		wasmModule := compiler.CompileProgram(program)

		for _, err := range compiler.Errors() {
			t.Fatal(err)
		}

		emitter := wasm.NewEmitter()
		err = emitter.Emit(wasmModule)
		if err != nil {
			t.Fatal(err)
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
}
