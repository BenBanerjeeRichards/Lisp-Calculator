package test

import (
	"fmt"

	"github.com/benbanerjeerichards/lisp-calculator/calc"
	"github.com/benbanerjeerichards/lisp-calculator/eval"
	"github.com/benbanerjeerichards/lisp-calculator/parser"
	"github.com/google/go-cmp/cmp"
)

func printTestFailedErr(code string, err error) {
	fmt.Printf("Failed: %s\nReason: Error - %v\n", code, err)
}

func printTestFailedNum(code string, expected float64, actual float64) {
	fmt.Printf("Failed: %s\nReason: Expected %f but got %f \n", code, expected, actual)
}

func printTestFailedBool(code string, expected bool, actual bool) {
	fmt.Printf("Failed: %s\nReason: Expected %v but got %v \n", code, expected, actual)
}
func printTestFailedString(code string, expected string, actual string) {
	fmt.Printf("Failed: %s\nReason: Expected %v but got %v \n", code, expected, actual)
}
func printTestFailedList(code string, expected []eval.Value, actual []eval.Value) {
	fmt.Printf("Failed: %s\nReason: Expected %v but got %v \n", code, expected, actual)
}

func printTokensFailed(code string, message string, expected []parser.Token, actual []parser.Token) {
	fmt.Printf("Failed: %s\nExpected %s\nRecieved %s\n%s\n", code, expected, actual, message)
}

func evalProgram(code string) (eval.Value, bool) {
	asts, err := calc.Ast(code)
	if err != nil {
		printTestFailedErr(code, err)
		return eval.Value{}, false
	}
	evalResult, err := eval.EvalProgram(asts)
	if err != nil {
		printTestFailedErr(code, err)
		return eval.Value{}, false
	}
	return evalResult, true
}

func ExpectNumber(code string, expected float64) bool {
	if evalResult, ok := evalProgram(code); ok {
		if evalResult.Kind != eval.NumType {
			fmt.Printf("Failed: %s\nReason: Expected %f but got type %s\n", code, expected, evalResult.Kind)
			return false
		}
		if evalResult.Num != expected {
			printTestFailedNum(code, expected, evalResult.Num)
			return false
		}
	}
	return true
}

func ExpectBool(code string, expected bool) bool {
	if evalResult, ok := evalProgram(code); ok {
		if evalResult.Kind != eval.BoolType {
			fmt.Printf("Failed: %s\nReason: Expected %v but got type %s\n", code, expected, evalResult.Kind)
			return false
		}
		if evalResult.Bool != expected {
			printTestFailedBool(code, expected, evalResult.Bool)
			return false
		}
	}
	return true
}

func ExpectList(code string, expected []eval.Value) bool {
	if evalResult, ok := evalProgram(code); ok {
		if evalResult.Kind != eval.ListType {
			fmt.Printf("Failed: %s\nReason: Expected %v but got type %s\n", code, expected, evalResult.Kind)
			return false
		}
		if len(evalResult.List) != len(expected) {
			printTestFailedList(code, expected, evalResult.List)
			return false
		}
		for i, actual := range evalResult.List {
			if !cmp.Equal(actual, expected[i]) {
				printTestFailedList(code, expected, evalResult.List)
				return false
			}
		}

	}
	return true
}

func ExpectString(code string, expected string) bool {
	if evalResult, ok := evalProgram(code); ok {
		if evalResult.Kind != eval.StringType {
			fmt.Printf("Failed: %s\nReason: Expected %v but got type %s\n", code, expected, evalResult.Kind)
			return false
		}
		if evalResult.String != expected {
			printTestFailedString(code, expected, evalResult.String)
			return false
		}
	}
	return true
}
func ExpectNull(code string) bool {
	if evalResult, ok := evalProgram(code); ok {
		if evalResult.Kind != eval.NullType {
			fmt.Printf("Failed: %s\nReason: Expected null but got type %s\n", code, evalResult.Kind)
			return false
		}
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
	ExpectTokens("(x)", []parser.Token{mkToken(parser.TokLBracket, ""), mkToken(parser.TokIdent, "x"),
		mkToken(parser.TokRBracket, "")})
	ExpectTokens("(5)", []parser.Token{mkToken(parser.TokLBracket, ""), mkToken(parser.TokNumber, "5"),
		mkToken(parser.TokRBracket, "")})
	ExpectTokens("(hello", []parser.Token{mkToken(parser.TokLBracket, ""), mkToken(parser.TokIdent, "hello")})
	ExpectTokens("+", []parser.Token{mkToken(parser.TokIdent, "+")})
	ExpectTokens(`"te"`, []parser.Token{mkToken(parser.TokString, "te")})
	ExpectTokens(`"hello \" world"`, []parser.Token{mkToken(parser.TokString, "hello \" world")})

	ExpectNumber("(5)", 5)
	ExpectNumber("(5.5)", 5.5)
	ExpectNumber("(-5.5)", -5.5)
	ExpectNumber("(- 3.3 -2.2)", 5.5)
	ExpectNumber("(-5)", -5)
	ExpectNumber("(+ 5 10)", 15)
	ExpectNumber("(+ 5 -10)", -5)
	ExpectNumber("(+ 5 (+ 3 6))", 14)
	ExpectNumber("(+ (+ 10 20) (+ 3 6))", 39)
	ExpectNumber("(+ (+ 10 20) 100)", 130)

	ExpectString(`("Hello World")`, "Hello World")

	ExpectBool("(true)", true)
	ExpectBool("(false)", false)
	ExpectBool("(< 10 5)", false)
	ExpectBool("(< 5 10)", true)
	ExpectBool("(> 5 10)", false)
	ExpectBool("(> 10 5)", true)
	ExpectBool("(> 10 10)", false)
	ExpectBool("(< 10 10)", false)
	ExpectBool("(>= 10 5)", true)
	ExpectBool("(>= 5 10)", false)
	ExpectBool("(>= 10 10)", true)
	ExpectBool("(<= 10 10)", true)
	ExpectBool("(<= 5 10)", true)
	ExpectBool("(<= 10 5)", false)
	ExpectBool("(= 10 10)", true)
	ExpectBool("(= 10 7)", false)

	ExpectList("(list 1 2 3)", []eval.Value{{Kind: eval.NumType, Num: 1},
		{Kind: eval.NumType, Num: 2}, {Kind: eval.NumType, Num: 3}})
	ExpectList("(list)", []eval.Value{})
	ExpectList(`(list 1 false null "s")`, []eval.Value{{Kind: eval.NumType, Num: 1},
		{Kind: eval.BoolType, Bool: false}, {Kind: eval.NullType}, {Kind: eval.StringType, String: "s"}})
	ExpectList("(list 1 (list 2 3) null)", []eval.Value{{Kind: eval.NumType, Num: 1},
		{Kind: eval.ListType, List: []eval.Value{{Kind: eval.NumType, Num: 2}, {Kind: eval.NumType, Num: 3}}}, {Kind: eval.NullType}})

	ExpectNull("(null)")

	ExpectNumber("(def x 10)(x)", 10)
	ExpectNumber("(def x (+ 3 7))(x)", 10)
	ExpectNumber("(def x1 10)(x1)", 10)
	ExpectNumber("(def var10able 10)(var10able)", 10)
	ExpectNumber("(def x 10)(+ x 5)", 15)
	ExpectNumber("(def x 10)(def y 20)(+ x y)", 30)
	ExpectNumber(
		`(def x 10)
		 (def x 20)
		 (x)`, 20)
	ExpectNumber("(/ 100 2)", 50)
	ExpectNumber("(* 100 2)", 200)
	ExpectNumber("(- 100 2)", 98)
	ExpectNumber("(^ 2 10)", 1024)
	ExpectNumber("(log 2 1024)", 10)
	ExpectNumber("(sqrt 9)", 3)
	ExpectNumber("(defun f (a b c) (* a (+ b c)))(f 2 7 2)", 18)
	ExpectNumber("(def x 10)(defun f (x) x)(f 20)", 20)
	ExpectNumber(
		`(defun f (x) 
			(def y 10)
			(+ x y))
		(f 100)`, 110)
	ExpectNumber("(defun f () (10))(f)", 10)
	ExpectNumber("(defun f (x) (+ x 1)) (f 2)", 3)
	ExpectNumber(
		`(def x 500)
		 (defun f (x) (+ x 1)) 
		 (f 10)
		 (x)
		`, 500)
	ExpectNumber(`
	(defun quadraticFirst (a b c)
		(def disc (- (^ b 2) (* (* 4 a) c)))
		(def first (/ 
			(+ (* -1 b) (sqrt disc)) 
			(* 2 a)))     
		(first))
	(quadraticFirst 2 5 3)`, -1)

	ExpectNumber("(if true 4 2)", 4)
	ExpectNumber("(if true 4)", 4)
	ExpectNull("(if false 4)")
	ExpectNumber("(if false 4 2)", 2)
	ExpectNumber("(if (< 10 5) (+ 4 10) (- 10 4))", 6)
	ExpectNumber(`(if (< 10 5) 
		(+ 4 10) 
		((def x 10)
		 (def y 20)
		 (- x y)
	))`, -10)

	ExpectNumber(`
	(def sum 0)
	(def x 5)
	(while (> x 0)
		(def sum (+ sum x))
		(def x (- x 1))
	)
	(sum)`,
		15)

}
