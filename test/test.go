package test

import (
	"fmt"

	"github.com/benbanerjeerichards/lisp-calculator/calc"
	"github.com/benbanerjeerichards/lisp-calculator/parser"
	"github.com/benbanerjeerichards/lisp-calculator/vm"
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
func printTestFailedList(code string, expected []vm.Value, actual []vm.Value) {
	fmt.Printf("Failed: %s\nReason: Expected %v but got %v \n", code, expected, actual)
}

func printTokensFailed(code string, message string, expected []parser.Token, actual []parser.Token) {
	fmt.Printf("Failed: %s\nExpected %s\nRecieved %s\n%s\n", code, expected, actual, message)
}

func evalProgram(code string) (vm.Value, bool) {
	asts, err := calc.Ast(code)
	if err != nil {
		printTestFailedErr(code, err)
		return vm.Value{}, false
	}
	compiler := vm.Compiler{}
	compiler.New()
	frame, err := compiler.CompileProgram(asts)
	if err != nil {
		printTestFailedErr(code, err)
		return vm.Value{}, false
	}
	evalResult, err := vm.Eval(frame, []string{"arg1", "arg2", "arg3"}, false)
	if err != nil {
		printTestFailedErr(code, err)
		return vm.Value{}, false
	}
	return evalResult, true
}

func (r *Runner) ExpectNumber(code string, expected float64) bool {
	if evalResult, ok := evalProgram(code); ok {
		if evalResult.Kind != vm.NumType {
			r.numFailed += 1
			fmt.Printf("Failed: %s\nReason: Expected %f but got type %s\n", code, expected, evalResult.Kind)
			return false
		}
		if evalResult.Num != expected {
			r.numFailed += 1
			printTestFailedNum(code, expected, evalResult.Num)
			return false
		}
		r.numPassed += 1
		return true
	}
	r.numFailed += 1
	return false
}

func (r *Runner) ExpectBool(code string, expected bool) bool {
	if evalResult, ok := evalProgram(code); ok {
		if evalResult.Kind != vm.BoolType {
			r.numFailed += 1
			fmt.Printf("Failed: %s\nReason: Expected %v but got type %s\n", code, expected, evalResult.Kind)
			return false
		}
		if evalResult.Bool != expected {
			r.numFailed += 1
			printTestFailedBool(code, expected, evalResult.Bool)
			return false
		}
		r.numPassed += 1
		return true
	}
	r.numFailed += 1
	return false
}

func (r *Runner) ExpectList(code string, expected []vm.Value) bool {
	if evalResult, ok := evalProgram(code); ok {
		if evalResult.Kind != vm.ListType {
			r.numFailed += 1
			fmt.Printf("Failed: %s\nReason: Expected %v but got type %s\n", code, expected, evalResult.Kind)
			return false
		}
		if len(evalResult.List) != len(expected) {
			r.numFailed += 1
			printTestFailedList(code, expected, evalResult.List)
			return false
		}
		for i, actual := range evalResult.List {
			if !cmp.Equal(actual, expected[i]) {
				r.numFailed += 1
				printTestFailedList(code, expected, evalResult.List)
				return false
			}
		}
		r.numPassed += 1
		return true
	}
	r.numFailed += 1
	return false
}

func (r *Runner) ExpectString(code string, expected string) bool {
	if evalResult, ok := evalProgram(code); ok {
		if evalResult.Kind != vm.StringType {
			r.numFailed += 1
			fmt.Printf("Failed: %s\nReason: Expected %v but got type %s\n", code, expected, evalResult.Kind)
			return false
		}
		if evalResult.String != expected {
			r.numFailed += 1
			printTestFailedString(code, expected, evalResult.String)
			return false
		}
		r.numPassed += 1
		return true
	}
	r.numFailed += 1
	return false
}
func (r *Runner) ExpectNull(code string) bool {
	if evalResult, ok := evalProgram(code); ok {
		if evalResult.Kind != vm.NullType {
			r.numFailed += 1
			fmt.Printf("Failed: %s\nReason: Expected null but got type %s (%s)\n", code, evalResult.Kind, evalResult.ToString())
			return false
		}
		r.numPassed += 1
		return true
	}
	r.numFailed += 1
	return false
}

func (r *Runner) ExpectError(code string) bool {
	ast, err := calc.Ast(code)
	if err != nil {
		r.numPassed += 1
		return true
	}
	c := vm.Compiler{}
	c.New()
	compileRes, err := c.CompileProgram(ast)
	if err != nil {
		r.numPassed += 1
		return true
	}
	res, err := vm.Eval(compileRes, []string{}, false)
	if err != nil {
		r.numPassed += 1
		return true
	}
	r.numFailed += 1
	fmt.Printf("Failed: %s\nReason: Expected an error but code compiled and eval'd to %v \n", code, res)
	return false
}

func (r *Runner) ExpectParseError(code string) bool {
	tokens := parser.Tokenise(code)
	p := parser.Parser{}
	p.New(tokens)
	_, err := p.ParseProgram()
	if err == nil {
		fmt.Printf("Failed: %s\nReason: Expected Parse error but code parsed successfully\n", code)
		r.numFailed += 1
		return false
	}
	r.numPassed += 1
	return true
}

func (r *Runner) ExpectTokens(code string, expected []parser.Token) bool {
	actual := parser.Tokenise(code)
	if len(actual) != len(expected) {
		r.numFailed += 1
		printTokensFailed(code, fmt.Sprintf("Expected %d tokens but got %d\n", len(expected), len(actual)), expected, actual)
		return false
	}

	for i, eTok := range expected {
		if eTok.Kind != actual[i].Kind || eTok.Data != actual[i].Data {
			r.numFailed += 1
			printTokensFailed(code, fmt.Sprintf("Token %d error", i), expected, actual)
			return false
		}
	}

	r.numPassed += 1
	return true
}

func mkToken(kind string, data string) parser.Token {
	return parser.Token{Kind: kind, Data: data}
}

type Runner struct {
	numPassed int
	numFailed int
}

func (r Runner) printSummary() {
	if r.numFailed == 0 {
		fmt.Printf("All %d tests passed\n", r.numPassed)
	} else {
		s := "s"
		if r.numFailed == 1 {
			s = ""
		}
		fmt.Printf("Ran %d tests with %d failure%s\n", r.numPassed+r.numFailed, r.numFailed, s)
	}
}

func Run() {
	r := Runner{numPassed: 0, numFailed: 0}
	r.ExpectTokens("(x)", []parser.Token{mkToken(parser.TokLBracket, ""), mkToken(parser.TokIdent, "x"),
		mkToken(parser.TokRBracket, "")})
	r.ExpectTokens("(5)", []parser.Token{mkToken(parser.TokLBracket, ""), mkToken(parser.TokNumber, "5"),
		mkToken(parser.TokRBracket, "")})
	r.ExpectTokens("(hello", []parser.Token{mkToken(parser.TokLBracket, ""), mkToken(parser.TokIdent, "hello")})
	r.ExpectTokens("+", []parser.Token{mkToken(parser.TokIdent, "+")})
	r.ExpectTokens(`"te"`, []parser.Token{mkToken(parser.TokString, "te")})
	r.ExpectTokens(`"hello \" world"`, []parser.Token{mkToken(parser.TokString, "hello \" world")})
	r.ExpectTokens("\"\n\"", []parser.Token{mkToken(parser.TokString, "\n")})

	r.ExpectNumber("(5)", 5)
	r.ExpectNumber("(5.5)", 5.5)
	r.ExpectNumber("(-5.5)", -5.5)
	r.ExpectNumber("(- 3.3 -2.2)", 5.5)
	r.ExpectNumber("(-5)", -5)
	r.ExpectNumber("(+ 5 10)", 15)
	r.ExpectNumber("(+ 5 -10)", -5)
	r.ExpectNumber("(+ 5 (+ 3 6))", 14)
	r.ExpectNumber("(+ (+ 10 20) (+ 3 6))", 39)
	r.ExpectNumber("(+ (+ 10 20) 100)", 130)

	r.ExpectParseError("(34")

	r.ExpectString(`("Hello World")`, "Hello World")

	r.ExpectBool("(true)", true)
	r.ExpectBool("(false)", false)

	// Comparison operations
	r.ExpectBool("(< 10 5)", false)
	r.ExpectBool("(< 5 10)", true)
	r.ExpectBool("(> 5 10)", false)
	r.ExpectBool("(> 10 5)", true)
	r.ExpectBool("(> 10 10)", false)
	r.ExpectBool("(< 10 10)", false)
	r.ExpectBool("(>= 10 5)", true)
	r.ExpectBool("(>= 5 10)", false)
	r.ExpectBool("(>= 10 10)", true)
	r.ExpectBool("(<= 10 10)", true)
	r.ExpectBool("(<= 5 10)", true)
	r.ExpectBool("(<= 10 5)", false)

	// Boolean logic
	r.ExpectBool("(not false)", true)
	r.ExpectBool("(not true)", false)
	r.ExpectBool("(and true false)", false)
	r.ExpectBool("(and false true)", false)
	r.ExpectBool("(and true true)", true)
	r.ExpectBool("(and false false)", false)
	r.ExpectBool("(or true false)", true)
	r.ExpectBool("(or false true)", true)
	r.ExpectBool("(or true true)", true)
	r.ExpectBool("(or false false)", false)

	// Equality
	r.ExpectBool("(= 10 10)", true)
	r.ExpectBool("(= 10 7)", false)
	r.ExpectBool("(= true false)", false)
	r.ExpectBool("(= false false)", true)
	r.ExpectBool("(= true true)", true)
	r.ExpectBool("(= null null)", true)
	r.ExpectBool(`(= "hello" "world")`, false)
	r.ExpectBool(`(= "hello" "hello")`, true)
	r.ExpectBool(`(= "" "")`, true)
	r.ExpectBool("(= (list) (list))", true)
	r.ExpectBool("(= (list 1 2) (list 1 2))", true)
	r.ExpectBool("(= (list 1 2) (list 1 3))", false)
	r.ExpectBool("(= (list) (list 1 3))", false)
	r.ExpectBool("(= (list false true) (list false true))", true)
	r.ExpectBool("(= (list false true) (list false false))", false)
	r.ExpectBool("(= (list null) (list null))", true)
	r.ExpectBool(`(= (list 1 false true 23 null "hello") (list 1 false true 23 null "hello"))`, true)
	r.ExpectBool("(= (list 1 false true 23 null) (list 1 false true 23 null))", true)
	r.ExpectBool(`(= (list "hello" "world") (list "hello" "world"))`, true)
	r.ExpectBool(`(= (list "hello" "world2") (list "hello" "world"))`, false)
	r.ExpectBool(`(= (list 1 2 (list true false)) (list 1 2 (list true false)))`, true)
	// TODO fix these tests
	r.ExpectBool(`(= (list 1 2 (list true false)) (list 1 2 (list null false)))`, false)
	r.ExpectBool(`(= (list 1 2 (list true false)) (list 1 2 (list false false)))`, false)

	r.ExpectList("(list 1 2 3)", []vm.Value{{Kind: vm.NumType, Num: 1},
		{Kind: vm.NumType, Num: 2}, {Kind: vm.NumType, Num: 3}})
	r.ExpectList("(list)", []vm.Value{})
	r.ExpectList(`(list 1 false null "s")`, []vm.Value{{Kind: vm.NumType, Num: 1},
		{Kind: vm.BoolType, Bool: false}, {Kind: vm.NullType}, {Kind: vm.StringType, String: "s"}})
	r.ExpectList("(list 1 (list 2 3) null)", []vm.Value{{Kind: vm.NumType, Num: 1},
		{Kind: vm.ListType, List: []vm.Value{{Kind: vm.NumType, Num: 2}, {Kind: vm.NumType, Num: 3}}}, {Kind: vm.NullType}})

	// List length
	r.ExpectNumber("(length (list))", 0)
	r.ExpectNumber("(length (list 1 2 3))", 3)
	r.ExpectNumber("(length (list 1 null false))", 3)
	r.ExpectNumber(`(length (list 1 (list 4 3 2 1 false "hello") 3))`, 3)
	r.ExpectNumber(`(length "test")`, 4)
	r.ExpectNumber(`(length "")`, 0)

	// Insert into list
	r.ExpectBool("(= (insert 0 10 (list 1 2 3)) (list 10 1 2 3))", true)
	r.ExpectBool("(= (insert -50 10 (list 1 2 3)) (list 10 1 2 3))", true)
	r.ExpectBool("(= (insert 1 10 (list 1 2 3)) (list 1 10 2 3))", true)
	r.ExpectBool("(= (insert 3 10 (list 1 2 3)) (list 1 2 3 10))", true)
	r.ExpectBool("(= (insert 30 10 (list 1 2 3)) (list 1 2 3 10))", true)

	// Index into list
	r.ExpectNumber("(nth 0 (list 1 2 3))", 1)
	r.ExpectNumber("(nth 1 (list 1 2 3))", 2)
	r.ExpectNumber("(nth 2 (list 1 2 3))", 3)
	r.ExpectNull("(nth -1 (list 1 2 3))")
	r.ExpectNull("(nth 4 (list 1 2 3))")
	r.ExpectNull("(nth 40 (list 1 2 3))")

	r.ExpectString(`(nth 0 "abcd")`, "a")
	r.ExpectString(`(nth 2 "abcd")`, "c")
	r.ExpectString(`(nth 3 "abcd")`, "d")
	r.ExpectNull(`(nth 4 "abcd")`)
	r.ExpectNull(`(nth 5 "abcd")`)
	r.ExpectNull(`(nth -1 "abcd")`)

	r.ExpectString("(chr 65)", "A")
	r.ExpectNumber(`(ord "A")`, 65)

	r.ExpectNull("(null)")

	r.ExpectNumber("(def x 10)(x)", 10)
	r.ExpectNumber("(def x (+ 3 7))(x)", 10)
	r.ExpectNumber("(def x1 10)(x1)", 10)
	r.ExpectNumber("(def var10able 10)(var10able)", 10)
	r.ExpectNumber("(def x 10)(+ x 5)", 15)
	r.ExpectNumber("(def x 10)(def y 20)(+ x y)", 30)
	r.ExpectNumber(
		`(def x 10)
		 (def x 20)
		 (x)`, 20)
	r.ExpectNumber("(/ 100 2)", 50)
	r.ExpectNumber("(* 100 2)", 200)
	r.ExpectNumber("(- 100 2)", 98)
	r.ExpectNumber("(^ 2 10)", 1024)
	r.ExpectNumber("(log 2 1024)", 10)
	r.ExpectNumber("(sqrt 9)", 3)
	r.ExpectNumber("(defun f (a b c) (* a (+ b c)))(f 2 7 2)", 18)
	r.ExpectNumber("(def x 10)(defun f (x) x)(f 20)", 20)
	r.ExpectNumber(
		`(defun f (x) 
			(def y 10)
			(+ x y))
		(f 100)`, 110)
	r.ExpectNumber("(defun f () (10))(f)", 10)
	r.ExpectNumber("(defun f (x) (+ x 1)) (f 2)", 3)
	r.ExpectNumber(
		`(def x 500)
		 (defun f (x) (+ x 1)) 
		 (f 10)
		 (x)
		`, 500)
	r.ExpectNumber(`
	(defun quadraticFirst (a b c)
		(def disc (- (^ b 2) (* (* 4 a) c)))
		(def first (/ 
			(+ (* -1 b) (sqrt disc)) 
			(* 2 a)))     
		(first))
	(quadraticFirst 2 5 3)`, -1)
	r.ExpectNumber(`
	(defun f (a b) (/ a b))
	(f 10 2)`, 5)
	r.ExpectNumber(`
	  (def x 10)
	  (defun f () x)
	  (f)
	`, 10)
	r.ExpectNumber(`
	  (def x 10)
	  (defun f (x) x)
	  (f 20)
	`, 20)

	r.ExpectNumber("(if true 4 2)", 4)
	r.ExpectNumber("(if true 4)", 4)
	r.ExpectNull("(if false 4)")
	r.ExpectNumber("(if false 4 2)", 2)
	r.ExpectNumber("(if (< 10 5) (+ 4 10) (- 10 4))", 6)
	r.ExpectNumber(`(if (< 10 5) 
		(+ 4 10) 
		((def x 10)
		 (def y 20)
		 (- x y)
	))`, -10)

	r.ExpectNumber(`(if true 1 0) (20)`, 20)

	r.ExpectNumber(`
	(def sum 0)
	(def x 5)
	(while (> x 0)
		(def sum (+ sum x))
		(def x (- x 1))
	)
	(sum)`,
		15)

	r.ExpectNumber(`
	(def f (lambda (x) (+ x 1)))
	(funcall f 20)
	`, 21)
	r.ExpectNumber(`
	(def f (lambda () 10))
	(funcall f)
	`, 10)
	r.ExpectNumber(`
	(def f (lambda (x) 
		(def t 20)
		(def y (+ t x))
		(* y 2)
		))
	(funcall f 4)
	`, 48)
	r.ExpectNumber(`
	(defun f () 
		(def x 10)
		(def l (lambda (y) (+ x y)))
		(funcall l 30)
	)
	(f)
	`, 40)
	r.ExpectNumber(`
	(defun f () 
		(def x 10)
		(def l (lambda (y) (+ x y)))
		(def x 100)
		(funcall l 30)
	)
	(f)
	`, 40)
	r.ExpectNumber(`
	(def x 200)
	(def f (lambda (l) (+ x l)))
	(def x 1000)
	(funcall f 5)
	`, 205)

	r.ExpectNumber(`
		((lambda (x y) (+ x y)) 10 20)
	`, 30)
	r.ExpectNumber(`
	(def x 10)
	(defun f (x y) (+ x y))
	(f 100 200)
	`, 300)
	r.ExpectNumber(`
	(defun plusOne (g a b)
       (def result (funcall g a b))
       (+ 1 result)
    )
    (plusOne (lambda (p q) (+ p q)) 10 20)`, 31)

	r.ExpectNumber(`
     (def x 10)
	 (defun f () 
		(def x 20)
		(+ x 1)
	 )
	 (x)
	`, 10)
	r.ExpectNumber(`
     (def x 10)
	 (defun f () 
		(def x 20)
		(+ x 1)
	 )
	 (f)
	 (x)
	`, 10)
	r.ExpectNumber(`
	  (def f (lambda (a b) (- a b)))
	  (funcall f 20 10)
	`, 10)

	// Functions can used before declaration
	r.ExpectNumber(`
	(f 20 )
    (defun f (x) (+ x 10))
    `, 30)

	// Strange but due to function hoisting
	r.ExpectNumber(`
	(f 20 )
    (defun f (x) (+ x 10))
    `, 30)

	r.ExpectNumber(`
	(def y (f 20))
    (defun f (x) (+ x 10))
    (y)`, 30)

	r.ExpectNumber(`
	(defun g (x) (f x))
    (defun f (x) (+ x 10))
     (g 20)`, 30)

	// Embedding function calls in arguments
	r.ExpectNumber(`
	(defun g (x) (200))
	(defun sum (a b) (+ a b))
    (sum 1 (g 3))`, 201)

	// Main function
	r.ExpectNumber(`
	(100)
	(defun main () 10)
	`, 10)
	r.ExpectString(`
	(100)
	(defun main (args) (nth 0 args))
	`, "arg1")
	r.ExpectString(`
	(100)
	(defun main (args) (nth 2 args))
	`, "arg3")
	r.ExpectNumber(`
	(defun f (a b c) (+ a (- b c)))
	(defun main (args) (f 10 20 30))
	`, 0)
	r.ExpectNumber(`
	(def aGlobalVariable 100)
	(defun main () (aGlobalVariable))
	`, 100)

	r.ExpectError(`(defun main (a b) a)`)

	// Concat
	r.ExpectString(`(concat "Hello " "World")`, "Hello World")

	// Comments
	r.ExpectNumber(`
	; Test comment string
	(def x 20);
	(def y 10); Inline comment example
	 (x)
	`, 20)

	r.ExpectNumber(`
	; Test comment string
	; This is a comment
	 (20)
	`, 20)
	r.ExpectNumber(`
	;; Test comment; string;
	 (20)
	`, 20)

	fmt.Print("\033[1m")
	r.printSummary()
	fmt.Print("\033[0m")
}
