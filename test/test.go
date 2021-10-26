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

func ExpectNumber(code string, expected float64) bool {
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
	ExpectNumber("(5)", 5)
	ExpectNumber("(5.5)", 5.5)
	ExpectNumber("(-5.5)", -5.5)
	ExpectNumber("(sub 3.3 -2.2)", 5.5)
	ExpectNumber("(-5)", -5)
	ExpectNumber("(add 5 10)", 15)
	ExpectNumber("(add 5 -10)", -5)
	ExpectNumber("(add 5 (add 3 6))", 14)
	ExpectNumber("(add (add 10 20) (add 3 6))", 39)
	ExpectNumber("(add (add 10 20) 100)", 130)
	ExpectNumber("(def x 10)(x)", 10)
	ExpectNumber("(def x 10)(add x 5)", 15)
	ExpectNumber("(def x 10)(def y 20)(add x y)", 30)
	ExpectNumber(
		`(def x 10)
		 (def x 20)
		 (x)`, 20)
	ExpectNumber("(div 100 2)", 50)
	ExpectNumber("(mul 100 2)", 200)
	ExpectNumber("(sub 100 2)", 98)
	ExpectNumber("(defun f (a b c) (mul a (add b c)))(f 2 7 2)", 18)
	ExpectNumber("(def x 10)(defun f (x) x)(f 20)", 20)
	ExpectNumber(
		`(defun f (x) 
			(def y 10)
			(add x y))
		(f 100)`, 110)
	ExpectNumber("(defun f () (10))(f)", 10)
	ExpectNumber("(defun f (x) (add x 1)) (f 2)", 3)
	ExpectNumber(
		`(def x 500)
		 (defun f (x) (add x 1)) 
		 (f 10)
		 (x)
		`, 500)
}
