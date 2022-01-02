package main

import (
	"fmt"
	"path/filepath"

	"github.com/benbanerjeerichards/lisp-calculator/calc"
	"github.com/benbanerjeerichards/lisp-calculator/test"
	"github.com/benbanerjeerichards/lisp-calculator/types"
	"github.com/benbanerjeerichards/lisp-calculator/util"
	"github.com/jessevdk/go-flags"
)

var opts struct {
	Test           bool `short:"t" long:"test" description:"Run tests"`
	Debug          bool `short:"D" long:"debug" description:"Print out a debug trace of instructions executed by the VM"`
	PrintTokens    bool `short:"T" long:"tokens" description:"Print out the tokens"`
	PrintParseTree bool `short:"P" long:"parse-tree" description:"Print out the parse"`
	PrintAst       bool `short:"A" long:"ast" description:"Print out the AST"`
}

func main() {
	args, _ := flags.Parse(&opts)

	if opts.Test {
		test.Run()
		return
	}
	if len(args) != 1 {
		fmt.Println("Provide file path to execute")
		return
	}
	file := args[0]

	filePath, _ := filepath.Abs(file)
	fileContents, err := util.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Failed to open file %s\n", file)
		return
	}
	opts := calc.RunOptions{Debug: opts.Debug, PrintParseTree: opts.PrintParseTree, PrintTokens: opts.PrintTokens, PrintAst: opts.PrintAst}
	evalResult, err := calc.ParseAndEval(filePath, fileContents, args, opts)
	if err != nil {
		if astError, ok := err.(types.Error); ok {
			fmt.Println(calc.AnnotateError(fileContents, astError))
		} else {
			fmt.Println(err)
		}
		return
	}
	fmt.Println(evalResult.ToString())
}
