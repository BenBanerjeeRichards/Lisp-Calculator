package calc

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/benbanerjeerichards/lisp-calculator/ast"
	"github.com/benbanerjeerichards/lisp-calculator/parser"
	"github.com/benbanerjeerichards/lisp-calculator/types"
	"github.com/benbanerjeerichards/lisp-calculator/util"
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
	if start > end || end > len(code) {
		return output
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
	asts, err := Ast(code)
	if err != nil {
		return vm.Value{}, err
	}

	compiler := vm.Compiler{}
	compiler.New()

	// loadStdLib(&evalulator)
	compileRes, err := compiler.CompileProgram(asts)
	if err != nil {
		return vm.Value{}, err
	}
	evalResult, err := vm.Eval(compileRes, programArgs)
	if err != nil {
		return vm.Value{}, err
	}
	return evalResult, nil
}

func Ast(code string) ([]ast.Ast, error) {
	fileAstResut, err := createAstForFile(code)
	if err != nil {
		return []ast.Ast{}, err
	}
	asts := fileAstResut.Asts

	for _, fileImport := range fileAstResut.Imports {
		importCodePath, ok := resolveImport(fileImport.Path)
		if !ok {
			return []ast.Ast{}, types.Error{Range: fileImport.Range, Simple: fmt.Sprintf("Failed to resolve import `%s` - does not exist", fileImport.Path)}
		}
		codeContents, err := util.ReadFile(importCodePath)
		if err != nil {
			return []ast.Ast{}, types.Error{Range: fileImport.Range, Simple: fmt.Sprintf("Failed to resolve import `%s` - file read failed", fileImport.Path)}
		}
		importAsts, err := Ast(codeContents)
		if err != nil {
			// TODO Need some sort of import stack trace
			return []ast.Ast{}, err
		}
		// TODO qualify import
		asts = append(asts, importAsts...)
	}
	return asts, nil
}

func resolveImport(importPath string) (string, bool) {
	// TODO this needs to be more sophisticated to resolve in a few different locations
	fullPath, err := filepath.Abs(importPath)
	if err != nil {
		return "", false
	}
	return fullPath, util.FileExists(fullPath)
}

func createAstForFile(code string) (ast.AstResult, error) {
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
