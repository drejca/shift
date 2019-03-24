package wasm

import "testing"

func TestDefine(t *testing.T) {
	expected := map[string]Symbol {
		"a": Symbol{Name: "a", Type: "i32", Scope: GlobalScope, Index: 0},
		"b": Symbol{Name: "b", Type: "i32", Scope: LocalScope, Index: 0},
		"c": Symbol{Name: "c", Type: "i32", Scope: LocalScope, Index: 1},
	}

	global := NewSymbolTable()

	a := global.Define("a", "i32")
	if a != expected["a"] {
		t.Errorf("expected a=%+v, got=%+v", expected["a"], a)
	}

	local := NewEnclosedSymbolTable(global)

	b := local.Define("b", "i32")
	if b != expected["b"] {
		t.Errorf("expected b=%+v, got=%+v", expected["b"], b)
	}

	c := local.Define("c", "i32")
	if c != expected["c"] {
		t.Errorf("expected c=%+v, got=%+v", expected["c"], c)
	}
}