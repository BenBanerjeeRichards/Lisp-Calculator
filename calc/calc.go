package calc

import (
	"fmt"

	"github.com/benbanerjeerichards/lisp-calculator/ast"
	"github.com/benbanerjeerichards/lisp-calculator/eval"
	"github.com/benbanerjeerichards/lisp-calculator/parser"
)

func ParseAndEval(code string) (eval.EvalResult, error) {
	ast, err := Ast(code)
	if err != nil {
		return eval.EvalResult{}, err
	}
	evalResult, err := eval.EvalProgram(ast)
	if err != nil {
		return eval.EvalResult{}, err
	}
	return evalResult, nil
}

func Ast(code string) ([]ast.Ast, error) {
	tokens := parser.Tokenise(code)
	calcParser := parser.Parser{}
	calcParser.New(tokens)
	syntaxTree, err := calcParser.ParseProgram()
	if err != nil {
		return []ast.Ast{}, fmt.Errorf("parser Error - %v", err)
	}
	astConstruct := ast.AstConstructor{}
	astConstruct.New()
	astTree, err := astConstruct.CreateAst(syntaxTree)
	if err != nil {
		return []ast.Ast{}, fmt.Errorf("parse Ast Error - %v", err)
	}
	return astTree, nil
}
