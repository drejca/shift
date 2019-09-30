package wasm_test

import (
	"testing"

	"github.com/drejca/shift/wasm"
)

func TestDefine(t *testing.T) {
	expected := map[string]wasm.Symbol{
		"a": wasm.Symbol{Name: "a", Type: "i32", Scope: wasm.GlobalScope, Index: 0},
		"b": wasm.Symbol{Name: "b", Type: "i32", Scope: wasm.LocalScope, Index: 0},
		"c": wasm.Symbol{Name: "c", Type: "i32", Scope: wasm.LocalScope, Index: 1},
	}

	global := wasm.NewSymbolTable()

	a := global.Define("a", "i32")
	if a != expected["a"] {
		t.Errorf("expected a=%+v, got=%+v", expected["a"], a)
	}

	local := wasm.NewEnclosedSymbolTable(global)

	b := local.Define("b", "i32")
	if b != expected["b"] {
		t.Errorf("expected b=%+v, got=%+v", expected["b"], b)
	}

	c := local.Define("c", "i32")
	if c != expected["c"] {
		t.Errorf("expected c=%+v, got=%+v", expected["c"], c)
	}
}
