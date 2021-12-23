package main

import (
	"fmt"
	"os"

	"github.com/benbanerjeerichards/lisp-calculator/calc"
	"github.com/benbanerjeerichards/lisp-calculator/test"
	"github.com/benbanerjeerichards/lisp-calculator/types"
	"github.com/benbanerjeerichards/lisp-calculator/util"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		return
	}
	if len(args) == 1 && args[0] == "-t" {
		test.Run()
		return
	}
	debug := false
	if args[0] == "-D" {
		debug = true
	}

	file := args[0]
	if debug {
		file = args[1]
	}
	fileContents, err := util.ReadFile(file)
	if err != nil {
		fmt.Printf("Failed to open file %s\n", file)
		return
	}
	evalResult, err := calc.ParseAndEval(fileContents, args, debug)
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
