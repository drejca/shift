package parser

import (
	"github.com/drejca/shiftlang/assert"
	"strings"
	"testing"
)

func TestParseFunc(t *testing.T) {
	input := `
fn Add(a i32, b i32) : i32 {
	return a + b;
}`
	p := New(strings.NewReader(input))
	program := p.Parse()

	for _, err := range p.errors {
		t.Error(err)
	}

	err := assert.EqualString(input, "\n" + program.String())
	if err != nil {
		t.Error(err)
	}
}