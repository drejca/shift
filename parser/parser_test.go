package parser

import (
	"errors"
	"github.com/drejca/shiftlang/assert"
	"github.com/drejca/shiftlang/token"
	"strings"
	"testing"
)

func TestParseFunc(t *testing.T) {
	input := `
fn add(a i32, b i32) : i32 {
	return (a + b);
}
fn Calc(a i32, b i32) : i32 {
	let c = 2;
	return (add(a, b) + c);
}`
	p := New(strings.NewReader(input))
	program := p.Parse()

	for _, err := range p.errors {
		t.Error(err)
	}

	err := assert.EqualString(input, program.String())
	if err != nil {
		t.Error(err)
	}
}

func TestReturnStatement(t *testing.T) {
	tests := []struct{
		input string
		err error
	} {
		{input: `return (2 - 1);`},
		{input: `return (5 + (2 - 1));`},
		{input: `let a = (5 - 2);`},
	}

	for _, test := range tests {
		p := New(strings.NewReader(test.input))
		program := p.Parse()

		for _, err := range p.errors {
			t.Error(err)
		}

		err := assert.EqualString(test.input, program.String())
		if err != nil {
			t.Error(err)
		}
	}
}

func TestReturnStatementErrors(t *testing.T) {
	tests := []struct{
		input string
		err error
		pos token.Position
	} {
		{input: `return ~2;`, err: errors.New("illegal symbol ~"), pos: token.Position{}},
		{input: `return 5 + (2 - 1;`, err: errors.New("missing )"), pos: token.Position{}},
		{input: `fn Add {}`, err: errors.New("missing ("), pos: token.Position{}},
		{input: `fn Add( {}`, err: errors.New("missing )"), pos: token.Position{}},
		{input: `fn Add()`, err: errors.New("missing {"), pos: token.Position{}},
		{input: `fn Add() {`, err: errors.New("missing }"), pos: token.Position{}},
	}

	for _, test := range tests {
		p := New(strings.NewReader(test.input))
		p.Parse()

		if len(p.errors) < 1 {
			t.Fatal("expected error")
		}

		if p.errors[0].Error() != test.err.Error() {
			t.Errorf("\nexpected:\n %s\ngot:\n %s", test.err.Error(), p.errors[0].Error())
		}
	}
}