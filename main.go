package main

import (
	"fmt"
	"os"

	"github.com/benbanerjeerichards/lisp-calculator/ast"
	"github.com/benbanerjeerichards/lisp-calculator/calc"
	"github.com/benbanerjeerichards/lisp-calculator/eval"
	"github.com/benbanerjeerichards/lisp-calculator/test"
	"github.com/benbanerjeerichards/lisp-calculator/util"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		test.Run()
		eval.RunRepl()
	}
	if len(args) == 1 && args[0] == "-i" {
		eval.RunRepl()
		return
	}
	if len(args) == 1 && args[0] == "-t" {
		test.Run()
		return
	}
	if len(args) == 1 {
		fileContents, err := util.ReadFile(args[0])
		if err != nil {
			fmt.Printf("Failed to open file %s\n", args[0])
			return
		}
		evalResult, err := calc.ParseAndEval(fileContents)
		if err != nil {
			if astError, ok := err.(ast.AstError); ok {
				fmt.Println(calc.AnnotateError(fileContents, astError))
			} else {
				fmt.Println(err)
			}
			return
		}
		fmt.Println(evalResult.ToString())
	}
}
