package wasm

import (
	"github.com/drejca/shiftlang/assert"
	"github.com/drejca/shiftlang/parser"
	"strings"
	"testing"
)

func TestCompileToString(t *testing.T) {
	input := `
fn Calc(a i32, b i32) : i32 {
	let c = 2;
	return add(a, b) + c;
}

fn add(a i32, b i32) : i32 {
	return a + b;
}
`
	p := parser.New(strings.NewReader(input))
	program := p.Parse()

	compiler := New()
	module := compiler.CompileProgram(program)

	for _, err := range compiler.Errors() {
		t.Error(err)
	}

	expected := `
(module 
	(func $Calc (param $a i32) (param $b i32) (result i32) (local $c i32)
		i32.const 2
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

	err := assert.EqualString(expected, "\n"+module.String())
	if err != nil {
		t.Error(err)
	}
}
