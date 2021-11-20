package calc

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/benbanerjeerichards/lisp-calculator/ast"
	"github.com/benbanerjeerichards/lisp-calculator/eval"
	"github.com/benbanerjeerichards/lisp-calculator/parser"
	"github.com/benbanerjeerichards/lisp-calculator/types"
	"github.com/benbanerjeerichards/lisp-calculator/util"
)

func AnnotateError(code string, error types.Error) string {
	// reset := "\033[0m"
	// bold := "\033[1m"
	// red := "\031[1m"
	output := fmt.Sprintf("%s - %s(%s)\n", error.Range, error.Simple, error.Detail)
	codeLines := strings.Split(code, "\n")
	start := error.Range.Start.Line - 1
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

func ParseAndEval(code string, programArgs []string) (eval.Value, error) {
	ast, err := Ast(code)
	if err != nil {
		return eval.Value{}, err
	}
	evalulator := eval.Evalulator{}
	evalResult, err := evalulator.EvalProgram(ast, programArgs)
	if err != nil {
		return eval.Value{}, err
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

func RunRepl() {
	reader := bufio.NewReader(os.Stdin)
	evalutor := eval.Evalulator{}
	astConstruct := ast.AstConstructor{}
	astConstruct.AllowFunctionRedeclaration = true
	astConstruct.New()

	for {
		fmt.Print("calc> ")
		text, _ := reader.ReadString('\n')
		text = text[:len(text)-1]
		if strings.HasPrefix(text, ":l") {
			// Load file into REPL environment
			parts := strings.Split(text, ":l ")
			if len(parts) <= 1 {
				fmt.Println("Incorrect format for load command. Expected :l <path>")
				continue
			}
			fileContents, err := util.ReadFile(parts[1])
			fmt.Println(parts[1])
			if err != nil {
				fmt.Println("Failed to load file", parts[1])
				continue
			}

			fileAst, err := Ast(fileContents)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = evalutor.UpdateGlobalState(fileAst)
			if err != nil {
				fmt.Println("Failed to initialize global state", err)
			}
			continue
		}
		val, err := runInRepl(text, &evalutor, &astConstruct)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(val.ToString())
	}
}

func runInRepl(code string, evalulator *eval.Evalulator, astConstruct *ast.AstConstructor) (eval.Value, error) {
	tokens := parser.Tokenise(code)
	parser := parser.Parser{}
	parser.New(tokens)
	expr, err := parser.ParseProgram()
	if err != nil {
		return eval.Value{}, err
	}

	ast, err := astConstruct.CreateAst(expr)
	if err != nil {
		return eval.Value{}, err
	}

	val, err := evalulator.EvalProgram(ast, []string{})
	if err != nil {
		return eval.Value{}, err
	}
	return val, nil

}
