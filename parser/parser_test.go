package parser

import (
	"errors"
	"github.com/drejca/shift/assert"
	"github.com/drejca/shift/token"
	"strings"
	"testing"
)

func TestParseFunc(t *testing.T) {
	input := `
fn Calc(a i32, b i32) : i32 {
	let c = 2
	c = (c + a)
	return (add(a, b) + c)
}

fn add(a i32, b i32) : i32 {
	return (a + b)
}
`
	p := New(strings.NewReader(input))
	program := p.ParseProgram()

	for _, err := range p.errors {
		t.Error(err.Error())
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
		{input: `return (2 - 1)`},
		{input: `return (5 + (2 - 1))`},
		{input: `let a = (5 - 2)`},
	}

	for _, test := range tests {
		p := New(strings.NewReader(test.input))
		program := p.ParseProgram()

		for _, err := range p.errors {
			t.Error(err)
		}

		err := assert.EqualString(test.input, program.String())
		if err != nil {
			t.Error(err)
		}
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct{
		input string
		errors []ParseError
	} {
		{input: `return ~2`, errors: []ParseError{{
				error: errors.New("illegal symbol ~"),
			 	position: token.Position{Line: 1, Column: 8 },
		}}},
		{input: `return 5 + (2 - 1`, errors: []ParseError{{
			error: errors.New("missing )"),
			position: token.Position{Line: 1, Column: 18},
		}}},
		{input: `fn Add {}`, errors: []ParseError{{
			error: errors.New("missing ("),
			position: token.Position{Line: 1, Column: 8},
		}}},
		{input: `fn Add( {}`,errors: []ParseError{{
			error: errors.New("missing )"),
			position: token.Position{Line: 1, Column: 9},
		}}},
		{input: `fn Add()`, errors: []ParseError{{
			error: errors.New("missing {"),
			position: token.Position{Line: 1, Column: 9},
		}}},
		{input: `fn Add() {`, errors: []ParseError{{
			error: errors.New("missing }"),
			position: token.Position{Line: 1, Column: 12},
		}}},
		{input: `fn Add(a i32, b) {}`, errors: []ParseError{{
			error: errors.New("missing function parameter type"),
			position: token.Position{Line: 1, Column: 16},
		}}},
		{input: `fn Add(a i32, b i32,) {}`, errors: []ParseError{{
			error: errors.New("trailing comma in parameters"),
			position: token.Position{Line: 1, Column: 21},
		}}},
		{input: `fn Add(a i32, b i32, {}`, errors: []ParseError{
			{
				error:    errors.New("trailing comma in parameters"),
				position: token.Position{Line: 1, Column: 22},
			},
			{
				error: errors.New("missing )"),
				position: token.Position{Line: 1, Column: 22},
			},
		}},
	}

	for i, test := range tests {
		p := New(strings.NewReader(test.input))
		p.ParseProgram()

		if len(p.errors) == 0 {
			t.Fatalf("%d) expected error for test", i+1)
		}

		for ei, err := range p.errors {
			if len(p.errors) != len(test.errors) {
				t.Fatalf("%d) e%d expected %d errors but got %d ", i+1, ei+1, len(test.errors), len(p.errors))
			}
			if err.Error().Error() != test.errors[ei].Error().Error() {
				t.Errorf("%d) e%d\nexpected:\n %s\ngot:\n %s", i+1, ei+1, test.errors[ei].Error(), err.Error())
			}

			if err.Position().Line != test.errors[ei].position.Line {
				t.Errorf("%d) e%d expected line %d but got %d", i+1, ei+1, test.errors[ei].position.Line, err.Position().Line)
			}

			if err.Position().Column != test.errors[ei].position.Column {
				t.Errorf("%d) e%d expected column %d but got %d", i+1, ei+1, test.errors[ei].position.Column, err.Position().Column)
			}
		}
	}
}