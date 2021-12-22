package calc

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/benbanerjeerichards/lisp-calculator/ast"
	"github.com/benbanerjeerichards/lisp-calculator/parser"
	"github.com/benbanerjeerichards/lisp-calculator/types"
	"github.com/benbanerjeerichards/lisp-calculator/vm"
	"github.com/c-bata/go-prompt"
)

//go:embed stdlib.lisp
var stdlibCode string

func AnnotateError(code string, error types.Error) string {
	// reset := "\033[0m"
	// bold := "\033[1m"
	// red := "\031[1m"
	output := fmt.Sprintf("%s - %s(%s)\n", error.Range, error.Simple, error.Detail)
	codeLines := strings.Split(code, "\n")
	start := error.Range.Start.Line - 2
	end := error.Range.End.Line + 1
	if end > len(codeLines)-1 {
		end = len(codeLines) - 1
	}
	if start < 0 {
		start = 0
	}

	for i, errorLine := range codeLines[start:end] {
		column := " "
		if i+1+start == error.Range.Start.Line {
			column = "*"
		}
		output = output + fmt.Sprint(column, errorLine, "\n")
	}
	return output
}

func ParseAndEval(code string, programArgs []string) (vm.Value, error) {
	ast, err := Ast(code)
	if err != nil {
		return vm.Value{}, err
	}
	compiler := vm.Compiler{}
	compiler.New()

	// loadStdLib(&evalulator)
	compileRes, err := compiler.CompileProgram(ast)
	if err != nil {
		return vm.Value{}, err
	}
	evalResult, err := vm.Eval(compileRes, programArgs)
	if err != nil {
		return vm.Value{}, err
	}
	return evalResult, nil
}

func Ast(code string) (ast.AstResult, error) {
	tokens := parser.Tokenise(code)
	calcParser := parser.Parser{}
	calcParser.New(tokens)
	syntaxTree, err := calcParser.ParseProgram()
	if err != nil {
		return ast.AstResult{}, err
	}
	astConstruct := ast.AstConstructor{}
	astConstruct.New()
	astTree, err := astConstruct.CreateAst(syntaxTree)
	if err != nil {
		return ast.AstResult{}, err
	}
	return astTree, nil
}
func completer(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}
