package calc

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/benbanerjeerichards/lisp-calculator/ast"
	"github.com/benbanerjeerichards/lisp-calculator/parser"
	"github.com/benbanerjeerichards/lisp-calculator/types"
	"github.com/benbanerjeerichards/lisp-calculator/util"
	"github.com/benbanerjeerichards/lisp-calculator/vm"
	"github.com/davecgh/go-spew/spew"
)

type RunOptions struct {
	Debug          bool
	PrintParseTree bool
	PrintTokens    bool
	PrintAst       bool
}

//go:embed stdlib.lisp
var stdlibCode string

func AnnotateError(code string, error types.Error) string {
	// reset := "\033[0m"
	// bold := "\033[1m"
	// red := "\031[1m"
	output := fmt.Sprintf("%s\n", error)
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

func ParseAndEval(path string, code string, programArgs []string, options RunOptions) (vm.Value, error) {
	asts, err := AstWithDebugOptions(path, code, options.PrintTokens, options.PrintParseTree, options.PrintAst)
	if err != nil {
		return vm.Value{}, err
	}

	compiler := vm.Compiler{}
	compiler.New()

	// loadStdLib(&evalulator)
	compileRes, err := compiler.CompileProgram(path, asts)
	if err != nil {
		return vm.Value{}, err
	}
	evalResult, err := vm.Eval(compileRes, programArgs, options.Debug, os.Stdout)
	if err != nil {
		return vm.Value{}, err
	}
	return evalResult, nil
}
func AstWithDebugOptions(path string, code string, printTokens bool, printParseTree bool, printAst bool) ([]ast.Ast, error) {
	fileAstResult, err := createAstForFile(path, code, printTokens, printParseTree)
	if printAst {
		spew.Dump(fileAstResult)
	}

	if err != nil {
		return []ast.Ast{}, err
	}
	asts := fileAstResult.Asts
	for _, fileImport := range fileAstResult.Imports {
		importCodePath, ok := resolveImportPath(path, fileImport.Path)
		if !ok {
			return []ast.Ast{}, types.Error{Range: fileImport.Range, Simple: fmt.Sprintf("Failed to resolve import `%s` - does not exist", fileImport.Path), File: path}
		}
		codeContents, err := util.ReadFile(importCodePath)
		if err != nil {
			return []ast.Ast{}, types.Error{Range: fileImport.Range, Simple: fmt.Sprintf("Failed to resolve import `%s` - file read failed", fileImport.Path), File: path}
		}
		importAsts, err := AstWithDebugOptions(importCodePath, codeContents, printTokens, printParseTree, printAst)
		if err != nil {
			return []ast.Ast{}, err
		}
		// TODO qualify import
		asts = append(asts, importAsts...)
	}
	return asts, nil

}

func Ast(path string, code string) ([]ast.Ast, error) {
	return AstWithDebugOptions(path, code, false, false, false)
}

func resolveImportPath(importFilePath, importPath string) (string, bool) {
	searchPaths := make([]string, 0)
	if len(importFilePath) > 0 {
		searchPaths = append(searchPaths, filepath.Dir(importFilePath))
	}
	stdlibPath, _ := filepath.Abs("calc")
	searchPaths = append(searchPaths, stdlibPath)
	fullPath := ""
	for _, searchPath := range searchPaths {
		fullPath = filepath.Join(searchPath, importPath)
		if util.FileExists(fullPath) {
			break
		}
	}
	if len(fullPath) == 0 {
		return "", false
	}
	return fullPath, true
}

func createAstForFile(path string, code string, printTokens bool, printParseTree bool) (ast.AstResult, error) {
	tokens := parser.Tokenise(code)
	if printTokens {
		spew.Dump(tokens)
	}
	calcParser := parser.Parser{}
	calcParser.New(tokens)
	syntaxTree, err := calcParser.ParseProgram()
	if err != nil {
		if ourErr, ok := err.(types.Error); ok {
			ourErr.File = path
			return ast.AstResult{}, err
		}
		return ast.AstResult{}, err
	}
	if printParseTree {
		fmt.Println(util.ParseTreeToString(syntaxTree))
	}
	astConstruct := ast.AstConstructor{}
	astConstruct.New()
	astTree, err := astConstruct.CreateAst(syntaxTree)
	if err != nil {
		if ourErr, ok := err.(types.Error); ok {
			ourErr.File = path
			return ast.AstResult{}, err
		}
		return ast.AstResult{}, err
	}
	for i := range astTree.Asts {
		astTree.Asts[i].FilePath = path
	}
	return astTree, nil
}
