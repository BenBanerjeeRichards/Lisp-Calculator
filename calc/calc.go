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
	PrintFunctions bool
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
	asts, err := AstWithDebugOptions(path, code, options.PrintTokens, options.PrintParseTree, options.PrintAst, options.PrintFunctions)
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

type file struct {
	filePath      string
	functionNames map[string]struct{}
	imports       []ast.Import
	asts          []ast.Ast
	// For debugging
	functionApplications []*ast.FunctionApplicationExpr
}

type AstBuilder struct {
	fileAsts       map[string]file
	functionNames  map[string]string
	printTokens    bool
	printParseTree bool
	printAst       bool
	printFunctions bool
}

func (a *AstBuilder) New() {
	a.fileAsts = make(map[string]file)
	a.functionNames = make(map[string]string)
	a.printTokens = false
	a.printParseTree = false
	a.printAst = false
}

// buildFile builds a file's ast and recursivly builds all its imports
// Only provide code if no file exists (e.g. from unit tests, command line, etc)
func (a *AstBuilder) buildFile(path string, code string) error {
	// Don't rebuild files we have already seen
	if _, ok := a.fileAsts[path]; ok {
		return nil
	}
	if len(code) == 0 {
		var err error
		code, err = util.ReadFile(path)
		if err != nil {
			return err
		}
	}
	astResult, err := createAstForFile(path, code, a.printTokens, a.printParseTree)
	if err != nil {
		return err
	}

	if a.printAst {
		spew.Dump(astResult)
	}

	functionNames := make(map[string]struct{})
	for _, anAst := range astResult.Asts {
		if anAst.Kind == ast.StmtType {
			if funDef, ok := anAst.Statement.(ast.FuncDefStmt); ok {
				functionNames[funDef.Identifier] = struct{}{}
			}
		}
	}

	a.fileAsts[path] = file{filePath: path, asts: astResult.Asts, imports: astResult.Imports, functionNames: functionNames}

	for idx, fileImport := range astResult.Imports {
		if fullPath, ok := resolveImportPath(path, fileImport.Path); ok {
			a.fileAsts[path].imports[idx].Path = fullPath
			err := a.buildFile(fullPath, "")
			if err != nil {
				return err
			}
		} else {
			return types.Error{Range: fileImport.Range, Simple: fmt.Sprintf("Failed to find file to import - %s", fileImport.Path)}
		}
	}
	return nil
}

func (a *AstBuilder) resolveFunctions() error {
	// Find all functionApplications and set the FilePath on them to resolve them to the correct file
	// So need to iterate over entire tree
	for _, theFile := range a.fileAsts {
		for _, anAst := range theFile.asts {
			err := a.resolveFunctionAst(theFile, anAst)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *AstBuilder) resolveFunctionAst(theFile file, node ast.Ast) error {
	if node.Kind == ast.StmtType {
		return a.resolveFunctionStatement(theFile, node.Statement)
	} else {
		return a.resolveFunctionExpression(theFile, node.Expression)
	}
}

func (a *AstBuilder) resolveFunctionFilePath(theFile file, expr *ast.FunctionApplicationExpr) error {
	if len(expr.Qualifier) != 0 {
		// If the function is qualified, then need to look up correctly
		for _, fileImport := range theFile.imports {
			if fileImport.Qualifier == expr.Qualifier {
				if importedFile, ok := a.fileAsts[fileImport.Path]; ok {
					if _, functionIsInFile := a.fileAsts[fileImport.Path].functionNames[expr.Identifier]; functionIsInFile {
						expr.FilePath = importedFile.filePath
						return nil
					} else {
						return types.Error{Range: expr.Range,
							Simple: fmt.Sprintf("Failed to find function %s in file %s (qualified by %s)", expr.Identifier, importedFile.filePath, expr.Qualifier)}
					}
				} else {
					spew.Dump(a.fileAsts)
					return types.Error{Range: expr.Range, Simple: fmt.Sprintf("Function application uses unknown qualifier %s", expr.Qualifier)}
				}
			}
		}
	} else {
		// Non qualified, could either 1) buildin 2) this file 3) unqualified import from other file
		// TODO we should have a buildin map for quick checking
		for _, builtin := range vm.Builtins {
			if builtin.Identifier == expr.Identifier {
				expr.IsBuiltin = true
				return nil
			}
		}
		if _, inThisFile := theFile.functionNames[expr.Identifier]; inThisFile {
			expr.FilePath = theFile.filePath
		} else {
			for _, fileImport := range theFile.imports {
				if len(fileImport.Qualifier) == 0 {
					if importedFile, ok := a.fileAsts[fileImport.Path]; ok {
						if _, ok := importedFile.functionNames[expr.Identifier]; ok {
							expr.FilePath = importedFile.filePath
							return nil
						}
					}
				}
			}
			// Not an issue if we can't resove a non-qualifid function application, as it could also be a local/global variable
			// Should probably rename the FunctionApplicationNode to prevent confusion
			return nil
		}
	}
	return nil
}

func (a *AstBuilder) resolveFunctionExpression(theFile file, node ast.Expr) error {
	switch expr := node.(type) {
	case ast.FunctionApplicationExpr:
		err := a.resolveFunctionFilePath(theFile, &expr)
		if err != nil {
			return err
		}
		if a.printFunctions {
			fmt.Println(expr.Range, expr.Qualifier, expr.Identifier, expr.FilePath)
		}
	case ast.BoolExpr, ast.StringExpr, ast.NullExpr, ast.NumberExpr, ast.VarUseExpr:
		// Do nothing for literals
	case ast.ClosureApplicationExpr:
		for _, arg := range expr.Args {
			err := a.resolveFunctionExpression(theFile, arg)
			if err != nil {
				return err
			}
		}
		return a.resolveFunctionExpression(theFile, expr.Closure)
	case ast.ClosureDefExpr:
		for _, bodyAst := range expr.Body {
			err := a.resolveFunctionAst(theFile, bodyAst)
			if err != nil {
				return err
			}
		}
	case ast.IfElseExpr:
		err := a.resolveFunctionExpression(theFile, expr.Condition)
		if err != nil {
			return err
		}
		for _, bodyAst := range expr.IfBranch {
			err := a.resolveFunctionAst(theFile, bodyAst)
			if err != nil {
				return err
			}
		}
		for _, bodyAst := range expr.ElseBranch {
			err := a.resolveFunctionAst(theFile, bodyAst)
			if err != nil {
				return err
			}
		}
	case ast.IfOnlyExpr:
		err := a.resolveFunctionExpression(theFile, expr.Condition)
		if err != nil {
			return err
		}
		for _, bodyAst := range expr.IfBranch {
			err := a.resolveFunctionAst(theFile, bodyAst)
			if err != nil {
				return err
			}
		}
	case ast.ListExpr:
		for _, valExpr := range expr.Value {
			err := a.resolveFunctionExpression(theFile, valExpr)
			if err != nil {
				return err
			}
		}
	case ast.StructAccessorExpr:
		return a.resolveFunctionExpression(theFile, expr.Struct)
	case ast.StructExpr:
		for _, valExpr := range expr.Values {
			err := a.resolveFunctionExpression(theFile, valExpr)
			if err != nil {
				return err
			}
		}
	default:
		spew.Dump(node)
		panic("unknown type")
	}

	return nil
}

func (a *AstBuilder) resolveFunctionStatement(theFile file, node ast.Stmt) error {
	switch stmt := node.(type) {
	case ast.FuncDefStmt:
		for _, bodyAst := range stmt.Body {
			err := a.resolveFunctionAst(theFile, bodyAst)
			if err != nil {
				return err
			}
		}
	case ast.ImportStmt, ast.ReturnStmt, ast.StructDefStmt:
		// Leaf, do nothing
	case ast.ReturnValueStmt:
		return a.resolveFunctionExpression(theFile, stmt.Value)
	case ast.StructFieldDeclarationStmt:
		return a.resolveFunctionExpression(theFile, stmt.Value)
	case ast.VarDefStmt:
		return a.resolveFunctionExpression(theFile, stmt.Value)
	case ast.WhileStmt:
		for _, bodyAst := range stmt.Body {
			err := a.resolveFunctionAst(theFile, bodyAst)
			if err != nil {
				return err
			}
		}
		return a.resolveFunctionExpression(theFile, stmt.Condition)
	default:
		spew.Dump(node)
		panic("unknown type")
	}
	return nil
}

func AstWithDebugOptions(path string, code string, printTokens bool, printParseTree bool, printAst bool, printFunctions bool) ([]ast.Ast, error) {
	builder := AstBuilder{}
	builder.New()
	builder.printTokens = printTokens
	builder.printParseTree = printParseTree
	builder.printAst = printAst
	builder.printFunctions = printFunctions

	err := builder.buildFile(path, code)
	if err != nil {
		return []ast.Ast{}, err
	}

	err = builder.resolveFunctions()
	if err != nil {
		return []ast.Ast{}, err
	}

	allAsts := []ast.Ast{}
	for _, fileAst := range builder.fileAsts {
		allAsts = append(allAsts, fileAst.asts...)
	}

	return allAsts, nil
}

func Ast(path string, code string) ([]ast.Ast, error) {
	return AstWithDebugOptions(path, code, false, false, false, false)
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
	return fullPath, util.FileExists(fullPath)
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
