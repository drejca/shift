package parser

import (
	"errors"
	"fmt"
	"github.com/drejca/shift/assert"
	"github.com/drejca/shift/print"
	"github.com/drejca/shift/token"
	"os"
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
	program, compilerError := p.ParseProgram()

	if compilerError != nil {
		t.Fatal(compilerError.Error())
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
		{input: `
fn calc() {
	return (2 - 1)
}
`},
		{input: `
fn calc() {
	return (5 + (2 - 1))
}
`},
		{input: `
fn calc() {
	let a = (5 - 2)
}
`},
		{input: `
fn calc() {
	let a i32 = (5 - 2)
}
`},
		{input: `let a = 0.6`},
	}

	for _, test := range tests {
		p := New(strings.NewReader(test.input))
		program, compilerError := p.ParseProgram()

		if compilerError != nil {
			t.Fatal(compilerError.Error())
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
		error ParseError
	} {
		{input: `fn () {}`, error: ParseError{
			error: errors.New("missing function name"),
			position: token.Position{Line: 1, Column: 3},
		}},
		{input: `fn A() {return ~2}`, error: ParseError{
				error: errors.New("illegal symbol ~"),
			 	position: token.Position{Line: 1, Column: 16},
		}},
		{input: `fn A() {return 5 + (2 - 1}`, error: ParseError{
			error: errors.New("missing )"),
			position: token.Position{Line: 1, Column: 26},
		}},
		{input: `fn Add {}`, error: ParseError{
			error: errors.New("missing ("),
			position: token.Position{Line: 1, Column: 8},
		}},
		{input: `fn Add( {}`,error: ParseError{
			error: errors.New("missing )"),
			position: token.Position{Line: 1, Column: 9},
		}},
		{input: `fn Add()`, error: ParseError{
			error: errors.New("missing {"),
			position: token.Position{Line: 1, Column: 9},
		}},
		{input: `fn Add() {`, error: ParseError{
			error: errors.New("missing }"),
			position: token.Position{Line: 1, Column: 12},
		}},
		{input: `fn Add(a i32, b) {}`, error: ParseError{
			error: errors.New("missing function parameter type"),
			position: token.Position{Line: 1, Column: 15},
		}},
		{input: `fn Add(a i32, b i32,) {}`, error: ParseError{
			error: errors.New("trailing comma in parameters"),
			position: token.Position{Line: 1, Column: 19},
		}},
		{input: `fn Add(a i32, b i32, {}`, error: ParseError{
			error:    errors.New("trailing comma in parameters"),
			position: token.Position{Line: 1, Column: 19},
		}},
	}

	for i, test := range tests {
		p := New(strings.NewReader(test.input))
		_, err := p.ParseProgram()

		if err == nil {
			t.Fatalf("%d) expected error for test", i+1)
		}

		if err.Error().Error() != test.error.Error().Error() {
			t.Errorf("%d) \nexpected:\n %s\ngot:\n %s", i+1, test.error.Error(), err.Error())
		}

		if err.Position().Line != test.error.position.Line {
			t.Errorf("%d) expected line %d but got %d", i+1, test.error.position.Line, err.Position().Line)
		}

		if err.Position().Column != test.error.position.Column {
			t.Errorf("%d) expected column %d but got %d", i+1, test.error.position.Column, err.Position().Column)
		}
	}
}

func TestReadFromFile(t *testing.T) {
	filename := "../testprogram/main.sf"

	file, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}

	p := New(file)
	_, parseErr := p.ParseProgram()

	file.Close()

	if parseErr != nil {
		refile, err := os.Open(filename)
		if err != nil {
			t.Fatal(err)
		}

		printer := print.New(refile)
		fmt.Print(printer.PrintError(parseErr))
	}
	file.Close()
}