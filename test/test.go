package test

import (
	"fmt"

	"github.com/benbanerjeerichards/lisp-calculator/calc"
	"github.com/benbanerjeerichards/lisp-calculator/eval"
	"github.com/benbanerjeerichards/lisp-calculator/parser"
)

func printTestFailedErr(code string, err error) {
	fmt.Printf("Failed: %s\nReason: Error - %v\n", code, err)
}

func printTestFailed(code string, expected float64, actual float64) {
	fmt.Printf("Failed: %s\nReason: Expected %f but got %f \n", code, expected, actual)
}

func printTokensFailed(code string, message string, expected []parser.Token, actual []parser.Token) {
	fmt.Printf("Failed: %s\nExpected %s\nRecieved %s\n%s\n", code, expected, actual, message)
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
		fmt.Printf("Failed: %s\nReason: Expected %f but got <NIL>\n", code, expected)
		return false
	}
	if evalResult.Value != expected {
		printTestFailed(code, expected, evalResult.Value)
		return false
	}
	return true
}

func ExpectTokens(code string, expected []parser.Token) bool {
	actual := parser.Tokenise(code)
	if len(actual) != len(expected) {
		printTokensFailed(code, fmt.Sprintf("Expected %d tokens but got %d\n", len(expected), len(actual)), expected, actual)
		return false
	}

	for i, eTok := range expected {
		if eTok.Kind != actual[i].Kind || eTok.Data != actual[i].Data {
			printTokensFailed(code, fmt.Sprintf("Token %d error", i), expected, actual)
			return false
		}
	}

	return true
}

func mkToken(kind string, data string) parser.Token {
	return parser.Token{Kind: kind, Data: data}
}

func Run() {
	ExpectTokens("(5)", []parser.Token{mkToken(parser.TokLBracket, ""), mkToken(parser.TokNumber, "5"),
		mkToken(parser.TokRBracket, "")})
	ExpectTokens("(hello", []parser.Token{mkToken(parser.TokLBracket, ""), mkToken(parser.TokString, "hello")})

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
	ExpectNumber("(def x1 10)(x1)", 10)
	ExpectNumber("(def var10able 10)(var10able)", 10)
	ExpectNumber("(def x 10)(add x 5)", 15)
	ExpectNumber("(def x 10)(def y 20)(add x y)", 30)
	ExpectNumber(
		`(def x 10)
		 (def x 20)
		 (x)`, 20)
	ExpectNumber("(div 100 2)", 50)
	ExpectNumber("(mul 100 2)", 200)
	ExpectNumber("(sub 100 2)", 98)
	ExpectNumber("(pow 2 10)", 1024)
	ExpectNumber("(log 2 1024)", 10)
	ExpectNumber("(sqrt 9)", 3)
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
	ExpectNumber(`
	(defun quadraticFirst (a b c)
		(def disc (sub (pow b 2) (mul (mul 4 a) c)))
		(def first (div 
			(add (mul -1 b) (sqrt disc)) 
			(mul 2 a)))     
		(first))
	(quadraticFirst 2 5 3)`, -1)
}
