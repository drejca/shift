package wasm

import (
	"github.com/drejca/shiftlang/assert"
	"github.com/drejca/shiftlang/parser"
	"strings"
	"testing"
)

func TestCompileToString(t *testing.T) {
	input := `
fn Add(a i32, b i32) : i32 {
	return a + b;
}
`
	p := parser.New(strings.NewReader(input))
	program := p.Parse()

	compiler := New()
	compiler.Compile(program)

	for _, err := range compiler.Errors() {
		t.Error(err)
	}

	expected := `
(module 
	(func $Add (param $a i32) (param $b i32) (result i32) 
		get_local $a
		get_local $b
		i32.add)
	(export "Add" (func $Add))
)`

	output := compiler.module.String()
	err := assert.EqualString(expected, "\n"+output)
	if err != nil {
		t.Error(err)
	}
}
