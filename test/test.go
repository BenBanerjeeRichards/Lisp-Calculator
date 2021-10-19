package test

import (
	"fmt"

	"github.com/benbanerjeerichards/lisp-calculator/calc"
	"github.com/benbanerjeerichards/lisp-calculator/eval"
)

func printTestFailedErr(code string, err error) {
	fmt.Printf("Failed: %s\nReason: Error - %v\n", code, err)
}

func printTestFailed(code string, expected float64, actual float64) {
	fmt.Printf("Failed: %s\nReason: Expected %f but got %f \n", code, expected, actual)
}

func expectNumber(code string, expected float64) bool {
	asts, err := calc.Ast(code)
	if err != nil {
		printTestFailedErr(code, err)
		return false
	}
	evalResult, err := eval.EvalProgram(asts)
	if err != nil {
		printTestFailedErr(code, err)
		return false
	}
	if !evalResult.HasValue {
		fmt.Printf("Failed: %s\nReason: Expected %f but got <NIL>", code, expected)
		return false
	}
	if evalResult.Value != expected {
		printTestFailed(code, expected, evalResult.Value)
		return false
	}
	return true
}

func Run() {
	expectNumber("(5)", 5)
	expectNumber("(add 5 10)", 15)
	expectNumber("(add 5 (add 3 6))", 14)
	expectNumber("(add (add 10 20) (add 3 6))", 39)
	expectNumber("(add (add 10 20) 100)", 130)
	expectNumber("(def x 10)(x)", 10)
	expectNumber("(def x 10)(add x 5)", 15)
	expectNumber("(def x 10)(def y 20)(add x y)", 30)
	expectNumber("(def x 10)(def x 20)(x)", 20)
	expectNumber("(div 100 2)", 50)
	expectNumber("(mul 100 2)", 200)
	expectNumber("(sub 100 2)", 98)
}
