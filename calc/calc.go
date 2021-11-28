package calc

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/benbanerjeerichards/lisp-calculator/ast"
	"github.com/benbanerjeerichards/lisp-calculator/eval"
	"github.com/benbanerjeerichards/lisp-calculator/parser"
	"github.com/benbanerjeerichards/lisp-calculator/types"
	"github.com/benbanerjeerichards/lisp-calculator/util"
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

func ParseAndEval(code string, programArgs []string) (eval.Value, error) {
	ast, err := Ast(code)
	if err != nil {
		return eval.Value{}, err
	}
	evalulator := eval.Evalulator{}
	loadStdLib(&evalulator)
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
func completer(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func RunRepl() {
	evalutor := eval.Evalulator{}
	astConstruct := ast.AstConstructor{}
	astConstruct.AllowFunctionRedeclaration = true
	astConstruct.New()
	loadedReplFilePath := ""
	loadedReplFileName := ""

	err := loadStdLib(&evalutor)
	if err != nil {
		fmt.Printf("Failed to load standard library - %v\n", err)
	}

	history := make([]string, 0)
	for {
		promptText := "calc> "
		if len(loadedReplFilePath) > 0 {
			promptText = fmt.Sprintf("%s> ", loadedReplFileName)
		}
		text := prompt.Input(promptText, completer, prompt.OptionHistory(history), prompt.OptionPrefixTextColor(prompt.DarkRed))
		history = append(history, text)
		if strings.HasPrefix(text, ":q") {
			break
		}
		if strings.HasPrefix(text, ":l") {
			// Load file into REPL environment
			parts := strings.Split(text, ":l ")
			if len(parts) <= 1 {
				fmt.Println("Incorrect format for load command. Expected :l <path>")
				continue
			}
			err := loadFileIntoRepl(parts[1], &evalutor)
			if err != nil {
				fmt.Println(err)
				continue
			}

			loadedReplFilePath = parts[1]
			loadedReplFileName = filepath.Base(parts[1])
			continue
		}
		if strings.HasPrefix(text, ":r") {
			if len(loadedReplFilePath) == 0 {
				fmt.Println("No file loaded into REPL")
				continue
			}
			// Clear evalulator context
			evalutor = eval.Evalulator{}
			err := loadFileIntoRepl(loadedReplFilePath, &evalutor)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = loadStdLib(&evalutor)
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println("Reloaded")
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

func loadFileIntoRepl(path string, evalulator *eval.Evalulator) error {
	fileContents, err := util.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to load file - %s", path)
	}

	fileAst, err := Ast(fileContents)
	if err != nil {
		return err
	}
	err = evalulator.UpdateGlobalState(fileAst)
	if err != nil {
		return err
	}
	return nil
}

func loadStdLib(evalulator *eval.Evalulator) error {
	fileAst, err := Ast(stdlibCode)
	if err != nil {
		return err
	}
	err = evalulator.UpdateGlobalState(fileAst)
	if err != nil {
		return err
	}
	return nil
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
