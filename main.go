package main

import (
	"fmt"
	"os"

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
	if len(args) == 1 {
		fileContents, err := util.ReadFile(args[0])
		if err != nil {
			fmt.Printf("Failed to open file %s\n", args[0])
			return
		}
		evalResult, err := calc.ParseAndEval(fileContents)
		if err != nil {
			fmt.Printf("Erorr: %s\n", err.Error())
			return
		}
		if evalResult.HasValue {
			fmt.Println(evalResult.Value)
		}
	}
}
