package wasm

type SymbolScope string

const (
	LocalScope SymbolScope = "LOCAL"
	GlobalScope = "GLOBAL"
)

type Symbol struct {
	Name string
	Type string
	Scope SymbolScope
	Index uint32
}

type SymbolTable struct {
	Outer *SymbolTable

	store map[string]Symbol
	numDefinitions uint32
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	return s
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{store: s}
}

func (s *SymbolTable) Define(name string, varType string) Symbol {
	symbol := Symbol{Name: name, Index: s.numDefinitions, Type: varType}
	if s.Outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}

	s.store[name] = symbol
	s.numDefinitions++
	return symbol
}

func (s *SymbolTable) Resolve(name string) (symbol Symbol, found bool) {
	obj, ok := s.store[name]
	if !ok && s.Outer != nil {
		obj, ok = s.Outer.Resolve(name)
		if !ok {
			return obj, ok
		}

		if obj.Scope == GlobalScope {
			return obj, ok
		}
	}
	return obj, ok
}