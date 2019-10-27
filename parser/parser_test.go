package parser_test

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/drejca/shift/assert"
	"github.com/drejca/shift/parser"
	"github.com/drejca/shift/print"
	"github.com/drejca/shift/token"
)

func TestParseFunc(t *testing.T) {
	input := `
import fn assert(expected i32, actual i32)

fn main() {
	res := Calc(6, 7)
	if (21 != res) {
		assert(21, res)
	}
}

fn Calc(a i32, b i32) : i32 {
	c := 2
	c = (c + a)
	return (add(a, b) + c)
}

fn add(a i32, b i32) : i32 {
	return (a + b)
}
`
	p := parser.New(strings.NewReader(input))
	program, compilerError := p.ParseProgram()

	if compilerError != nil {
		printer := print.New(strings.NewReader(input))
		t.Fatal(printer.PrintError(compilerError))
	}

	err := assert.EqualString(input, program.String())
	if err != nil {
		t.Error(err)
	}
}

func TestReturnStatement(t *testing.T) {
	tests := []struct {
		input string
		err   error
	}{
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
	a := (5 - 2)
}
`},
		{input: `
fn name() {
	name := "shift"
}
`},
		{input: `
import fn error(msg string)
`},
	}

	for _, test := range tests {
		p := parser.New(strings.NewReader(test.input))
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
	tests := []struct {
		input    string
		parseErr parser.ParseError
	}{
		{input: `fn () {}`, parseErr: parser.ParseError{
			Err: errors.New("missing function name"),
			Pos: token.Position{Line: 1, Column: 3},
		}},
		{input: `fn A() {return ~2}`, parseErr: parser.ParseError{
			Err: errors.New("illegal symbol ~"),
			Pos: token.Position{Line: 1, Column: 16},
		}},
		{input: `fn A() {return 5 + (2 - 1}`, parseErr: parser.ParseError{
			Err: errors.New("missing )"),
			Pos: token.Position{Line: 1, Column: 26},
		}},
		{input: `fn Add {}`, parseErr: parser.ParseError{
			Err: errors.New("missing ("),
			Pos: token.Position{Line: 1, Column: 8},
		}},
		{input: `fn Add( {}`, parseErr: parser.ParseError{
			Err: errors.New("missing )"),
			Pos: token.Position{Line: 1, Column: 9},
		}},
		{input: `fn Add()`, parseErr: parser.ParseError{
			Err: errors.New("missing {"),
			Pos: token.Position{Line: 1, Column: 9},
		}},
		{input: `fn Add() {`, parseErr: parser.ParseError{
			Err: errors.New("missing }"),
			Pos: token.Position{Line: 1, Column: 12},
		}},
		{input: `fn Add(a i32, b) {}`, parseErr: parser.ParseError{
			Err: errors.New("missing function parameter type"),
			Pos: token.Position{Line: 1, Column: 15},
		}},
		{input: `fn Add(a i32, b i32,) {}`, parseErr: parser.ParseError{
			Err: errors.New("trailing comma in parameters"),
			Pos: token.Position{Line: 1, Column: 19},
		}},
		{input: `fn Add(a i32, b i32, {}`, parseErr: parser.ParseError{
			Err: errors.New("trailing comma in parameters"),
			Pos: token.Position{Line: 1, Column: 19},
		}},
	}

	for i, test := range tests {
		p := parser.New(strings.NewReader(test.input))
		_, err := p.ParseProgram()

		if err == nil {
			t.Fatalf("%d) expected error for test", i+1)
		}

		if err.Error().Error() != test.parseErr.Err.Error() {
			t.Errorf("%d) \nexpected:\n %s\ngot:\n %s", i+1, test.parseErr.Err, err.Error())
		}

		if err.Position().Line != test.parseErr.Pos.Line {
			t.Errorf("%d) expected line %d but got %d", i+1, test.parseErr.Pos.Line, err.Position().Line)
		}

		if err.Position().Column != test.parseErr.Pos.Column {
			t.Errorf("%d) expected column %d but got %d", i+1, test.parseErr.Pos.Column, err.Position().Column)
		}
	}
}

func TestReadFromFile(t *testing.T) {
	filename := "../testprogram/main.sf"

	file, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}

	p := parser.New(file)
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
