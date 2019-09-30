package wasm_test

import (
	"strings"
	"testing"

	"github.com/drejca/shift/assert"
	"github.com/drejca/shift/parser"
	"github.com/drejca/shift/wasm"
)

func TestCompileMainToString(t *testing.T) {
	input := `
fn main() {
}
`
	p := parser.New(strings.NewReader(input))
	program, parseErr := p.ParseProgram()
	if parseErr != nil {
		t.Fatal(parseErr.Error())
	}

	compiler := wasm.NewCompiler()
	wasmModule := compiler.CompileProgram(program)

	for _, err := range compiler.Errors() {
		t.Error(err)
	}

	expected := `
(module 
	(type $t0 (func))
	(func $main (export "main") (type $t0))
)`

	err := assert.EqualString(expected, "\n"+wasmModule.String())
	if err != nil {
		t.Error(err)
	}
}

func TestCompileToString(t *testing.T) {
	input := `
import fn assert(expected i32, actual i32)

fn main() {
	res := Calc(6, 7)
	assert(21, res)
}

fn Calc(a i32, b i32) : i32 {
	c := 2
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

	compiler := wasm.NewCompiler()
	wasmModule := compiler.CompileProgram(program)

	for _, err := range compiler.Errors() {
		t.Error(err)
	}

	expected := `
(module 
	(type $t0 (func (param i32) (param i32)))
	(type $t1 (func))
	(type $t2 (func (param i32) (param i32) (result i32)))
	(import "env" "assert" (func $assert (type $t0)))
	(func $main (export "main") (type $t1) (local $res i32)
		(call $Calc (i32.const 6) (i32.const 7))
		set_local $res
		(call $assert (i32.const 21) (get_local $res)))
	(func $Calc (export "Calc") (type $t2) (param $a i32) (param $b i32) (result i32) (local $c i32)
		i32.const 2
		set_local $c
		get_local $c
		get_local $a
		i32.add
		set_local $c
		(call $add (get_local $a) (get_local $b))
		get_local $c
		i32.add)
	(func $add (type $t2) (param $a i32) (param $b i32) (result i32)
		get_local $a
		get_local $b
		i32.add)
)`

	err := assert.EqualString(expected, "\n"+wasmModule.String())
	if err != nil {
		t.Error(err)
	}
}
