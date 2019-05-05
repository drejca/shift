package print

import (
	"bufio"
	"bytes"
	"github.com/drejca/shift/token"
	"io"
	"strconv"
)

type Printer struct {
	buf *bufio.Reader
	lines [][]byte
}

func New(input io.Reader) *Printer {
	printer := &Printer{}

	buf := bufio.NewReader(input)
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			break
		}
		printer.lines = append(printer.lines, line)
	}
	return printer
}

func (p *Printer) PrintError(err token.CompileError) string {
	var out bytes.Buffer

	lineNumberText := strconv.Itoa(err.Position().Line)
	out.WriteString("\n[")
	out.WriteString(lineNumberText)
	out.WriteString("]  ")
	out.WriteString(p.PrintLine(err.Position().Line))
	out.WriteString("\n")
	for i := 0; i < (err.Position().Column + len(lineNumberText) + 3); i++ {
		out.WriteString(" ")
	}
	out.WriteString("^\n")
	out.WriteString(err.Error().Error())

	return out.String()
}

func (p *Printer) PrintLine(line int) string {
	return string(p.lines[line-1])
}

