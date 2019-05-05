package main

import (
	"fmt"
	"github.com/drejca/shift/parser"
	"github.com/drejca/shift/print"
	"github.com/drejca/shift/wasm"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func main() {
	app := cli.NewApp()

	app.Name = "Shift language compiler"
	app.Description = "Shift language compiler - cli application"
	app.Usage = "cli application"

	app.Commands = []cli.Command{
		{
			Name:    "build",
			Aliases: []string{"b"},
			Usage:   "build [filename]",
			Action:  build,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func build(c *cli.Context) error {
	filename := c.Args().First()
	fmt.Println("build: ", filename)

	file, err := os.Open(filename)
	if err != nil {
		fmt.Print(err)
		return err
	}

	p := parser.New(file)
	program, parseErr := p.ParseProgram()

	file.Close()

	if parseErr != nil {
		file, err = os.Open(filename)
		if err != nil {
			fmt.Print(err)
			return err
		}

		printer := print.New(file)
		fmt.Print(printer.PrintError(parseErr))

		file.Close()
		return parseErr.Error()
	}

	compiler := wasm.NewCompiler()
	wasmModule := compiler.CompileProgram(program)

	for _, err := range compiler.Errors() {
		fmt.Print(err)
		return err
	}

	emitter := wasm.NewEmitter()
	err = emitter.Emit(wasmModule)
	if err != nil {
		fmt.Print(err)
		return err
	}

	fileExtPos := strings.LastIndex(filename, ".")
	if fileExtPos != -1 {
		filename = filename[:fileExtPos]
	}

	err = ioutil.WriteFile(filename + ".wasm", emitter.Bytes(), 0644)
	if err != nil {
		fmt.Print(err)
		return err
	}
	return nil
}
