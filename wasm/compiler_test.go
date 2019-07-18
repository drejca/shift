package wasm

import (
	"github.com/drejca/shift/assert"
	"github.com/drejca/shift/parser"
	"strings"
	"testing"
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

	compiler := NewCompiler()
	wasmModule := compiler.CompileProgram(program)

	for _, err := range compiler.Errors() {
		t.Error(err)
	}

	expected := `
(module 
	(func $main )
	(export "main" (func $main))
)`

	err := assert.EqualString(expected, "\n"+wasmModule.String())
	if err != nil {
		t.Error(err)
	}
}

func TestCompileToString(t *testing.T) {
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

	expected := `
(module 
	(func $Calc (param $a i32) (param $b i32) (result i32) (local $c i32)
		i32.const 2
		set_local $c
		get_local $c
		get_local $a
		i32.add
		set_local $c
		(call $add (get_local $a) (get_local $b))
		get_local $c
		i32.add)
	(func $add (param $a i32) (param $b i32) (result i32) 
		get_local $a
		get_local $b
		i32.add)
	(export "Calc" (func $Calc))
)`

	err := assert.EqualString(expected, "\n"+wasmModule.String())
	if err != nil {
		t.Error(err)
	}
}
