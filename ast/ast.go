package ast

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/benbanerjeerichards/lisp-calculator/parser"
	"github.com/benbanerjeerichards/lisp-calculator/types"
)

// AST node structure based on that used in the go compiler
// https://github.com/golang/go/blob/master/src/go/ast/ast.go

type Expr interface {
	exprType()
	GetRange() types.FileRange
}

type Stmt interface {
	stmtType()
	GetRange() types.FileRange
}

const (
	ExprType = "ExprType"
	StmtType = "StmtType"
)

type AstConstructor struct {
	AllowFunctionRedeclaration bool
	Functions                  map[string]*FuncDefStmt
	GlobalVariables            map[string]*VarDefStmt
	Imports                    []Import
}

func (constructor *AstConstructor) New() {
	constructor.Functions = make(map[string]*FuncDefStmt)
	constructor.GlobalVariables = make(map[string]*VarDefStmt)
	constructor.Imports = make([]Import, 0)
	constructor.AllowFunctionRedeclaration = false
}

type Import struct {
	// Path to file to import into current file
	Path string
	// Optional qualifier (may be empty). If set, will qualify all imports by this (e.g. qualifier.symbol)
	Qualifier string
	Range     types.FileRange
}

type AstResult struct {
	Asts    []Ast
	Imports []Import
}

func (constructor *AstConstructor) CreateAst(rootExpression parser.Node) (AstResult, error) {
	asts, err := constructor.createAst(rootExpression, true)
	if err != nil {
		return AstResult{}, err
	}
	return AstResult{Asts: asts, Imports: constructor.Imports}, nil
}

func (constructor *AstConstructor) createAst(expr parser.Node, isRoot bool) ([]Ast, error) {
	asts := make([]Ast, 0)
	for _, expression := range expr.Children {
		ast, err := constructor.createAstItem(expression, isRoot)
		if err != nil {
			return []Ast{}, err
		}
		asts = append(asts, ast)
	}
	return asts, nil
}

func singleNestedExpr(node parser.Node) (parser.Node, error) {
	if node.Kind != parser.ExpressionNode {
		return parser.Node{}, errors.New("single nested expression must be expression")
	}
	if len(node.Children) != 1 {
		return parser.Node{}, errors.New("must only have single nested node")
	}
	return node.Children[0], nil
}

// Helper for traversing the common Expr -> Expr -> LiteralNode
func nestedLiteralValue(node parser.Node) (bool, string) {
	if node.Kind != parser.ExpressionNode || len(node.Children) == 0 {
		return false, ""
	}
	if len(node.Children) > 0 && node.Children[0].Kind == parser.ExpressionNode && len(node.Children[0].Children) == 1 && node.Children[0].Children[0].Kind == parser.LiteralNode {
		return true, node.Children[0].Children[0].Data
	}

	return false, ""
}

func safeTraverse(node parser.Node, childIndexes []int) (parser.Node, bool) {
	for _, idx := range childIndexes {
		if idx >= len(node.Children) {
			return parser.Node{}, false
		}
		node = node.Children[idx]
	}
	return node, true
}

func (constructor *AstConstructor) createAstItem(node parser.Node, isRoot bool) (Ast, error) {
	if ok, val := nestedLiteralValue(node); ok && (val == "def" || val == "defun" || val == "while" || val == "import" || val == "defstruct") {
		varDefStmt, err := constructor.createAstStatement(node, isRoot)
		if err != nil {
			return Ast{}, err
		}
		ast := Ast{}
		ast.newStatement(varDefStmt)
		return ast, nil
	}

	// Everything else must be an expression
	expr, err := constructor.createAstExpression(node)
	if err != nil {
		return Ast{}, err
	}
	ast := Ast{}
	ast.newExpression(expr)
	return ast, nil
}

func (constructor *AstConstructor) CreateAstItem(node parser.Node) (Ast, error) {
	return constructor.createAstItem(node, false)
}

func (constructor *AstConstructor) createAstExpression(node parser.Node) (Expr, error) {
	switch node.Kind {
	case parser.NumberNode:
		f, err := strconv.ParseFloat(node.Data, 64)
		if err != nil {
			return nil, types.Error{Range: node.Range,
				Simple: fmt.Sprintf("Failed to parse `%s` as float", node.Data)}
		}
		return NumberExpr{Value: f, Range: node.Range}, nil
	case parser.BoolNode:
		if node.Data == "true" {
			return BoolExpr{Value: true, Range: node.Range}, nil
		} else if node.Data == "false" {
			return BoolExpr{Value: false, Range: node.Range}, nil
		} else {
			return nil, types.Error{Range: node.Range,
				Simple: fmt.Sprintf("Failed to parse `%s` as bool", node.Data)}
		}
	case parser.NullNode:
		return NullExpr{Range: node.Range}, nil
	case parser.LiteralNode:
		return VarUseExpr{Identifier: node.Data, Range: node.Range}, nil
	case parser.StringNode:
		return StringExpr{Value: node.Data, Range: node.Range}, nil
	case parser.AccessorOperationNode:
		return constructor.createStructAccessorOperation(node)
	case parser.AccessorNode:
		return constructor.createStructAccessorFromShortenedNotation(node)
	case parser.ExpressionNode:
		if len(node.Children) == 0 {
			return nil, types.Error{Range: node.Range,
				Simple: "Parse error",
				Detail: "Expression must have non-zero children"}
		}
		litNode, litNodeOk := safeTraverse(node, []int{0, 0})
		isFirstNodeLiteral := litNodeOk && len(node.Children[0].Children) == 1
		if isFirstNodeLiteral {
			if litNode.Kind == parser.LiteralNode {
				if litNode.Data == "if" {
					return constructor.createIfExpr(node)
				} else if litNode.Data == "list" {
					return constructor.createList(node)
				} else if litNode.Data == "lambda" {
					return constructor.createClosure(node)
				} else if litNode.Data == "struct" {
					return constructor.createStruct(node)
				} else if litNode.Data == "funcall" {
					// Force application of closure value
					return constructor.createAppExpr(node.Children[1:], node.Range)
				} else {
					return constructor.createFuncAppExpr(node)
				}
			}
		}
		if len(node.Children) == 1 {
			return constructor.createAstExpression(node.Children[0])
		}
		if !isFirstNodeLiteral {
			return constructor.createAppExpr(node.Children, node.Range)
		}
	}
	return nil, types.Error{Simple: "Parse Error",
		Detail: fmt.Sprintf("unknown syntax node kind %s", node.Kind),
		Range:  node.Range}
}

func (constructor *AstConstructor) createStruct(node parser.Node) (StructExpr, error) {
	// Create a struct, with optional initialization
	if len(node.Children) < 2 {
		return StructExpr{}, types.Error{Range: node.Range, Simple: "Invalid struct initialization - struct name required"}
	}
	structName, err := singleNestedExpr(node.Children[1])
	if err != nil {
		return StructExpr{}, types.Error{Range: node.Range, Simple: "Invalid struct initialization - struct name required"}
	}

	// Now consider optional initialization
	initialValues := make(map[string]Expr)
	for i, initializationNode := range node.Children[2:] {
		if len(initializationNode.Children) != 2 {
			return StructExpr{}, types.Error{Range: node.Range,
				Simple: fmt.Sprintf("Invalid struct initialization - initalizer %d has invalid syntax: format is (fieldName value)", i+1)}
		}
		fieldNameNode, err := singleNestedExpr(initializationNode.Children[0])
		if err != nil {
			return StructExpr{}, types.Error{Range: node.Range,
				Simple: fmt.Sprintf("Invalid struct initialization - initalizer %d has invalid syntax: format is (fieldName value)", i+1)}
		}
		fieldValue, err := constructor.createAstExpression(node.Children[1])
		if err != nil {
			return StructExpr{}, err
		}
		initialValues[fieldNameNode.Data] = fieldValue
	}
	return StructExpr{StructIdentifier: structName.Data, Values: initialValues}, nil
}

func (constructor *AstConstructor) createStructAccessorFromShortenedNotation(node parser.Node) (StructAccessorExpr, error) {
	if len(node.Children) != 2 {
		return StructAccessorExpr{}, types.Error{Range: node.Range, Simple: "Invalid accessor operation syntax - format is (<structName><fieldName>)"}
	}

	// This will always by a literal expression
	structName, err := constructor.createAstExpression(node.Children[0])
	if err != nil {
		return StructAccessorExpr{}, err
	}
	return StructAccessorExpr{Range: node.Range, Struct: structName, FieldIdentifier: node.Children[1].Data}, nil
}

func (constructor *AstConstructor) createStructAccessorOperation(node parser.Node) (StructAccessorExpr, error) {
	if len(node.Children) != 2 {
		return StructAccessorExpr{}, types.Error{Range: node.Range, Simple: "Invalid accessor operation syntax - format is (:<field name> <struct expression>)"}
	}
	structExpr, err := constructor.createAstExpression(node.Children[1])
	if err != nil {
		return StructAccessorExpr{}, nil
	}
	if node.Children[0].Kind != parser.LiteralNode {
		return StructAccessorExpr{}, types.Error{Range: node.Children[0].Range, Simple: fmt.Sprintf("Struct field name must be a literal (got %s)", node.Children[0].Kind)}
	}
	return StructAccessorExpr{Range: node.Range, FieldIdentifier: node.Children[0].Data, Struct: structExpr}, nil
}

func (constructor *AstConstructor) createList(node parser.Node) (Expr, error) {
	list := ListExpr{Value: make([]Expr, len(node.Children)-1), Range: node.Range}
	for i, valueNode := range node.Children[1:] {
		valueExpr, err := constructor.createAstExpression(valueNode)
		if err != nil {
			return nil, err
		}
		list.Value[i] = valueExpr
	}
	return list, nil
}

func (constructor *AstConstructor) createWhileLoop(node parser.Node) (Stmt, error) {
	if len(node.Children) < 3 {
		return nil, types.Error{Range: node.Range,
			Simple: "Syntax error for while",
			Detail: fmt.Sprintf("Expected >= 3 children for while, got %d", len(node.Children)),
		}
	}
	cond, err := constructor.createAstExpression(node.Children[1])
	if err != nil {
		return nil, err
	}
	whileStmt := WhileStmt{Condition: cond, Body: make([]Ast, 0), Range: node.Range}
	for _, expr := range node.Children[2:] {
		exprAst, err := constructor.CreateAstItem(expr)
		if len(node.Children)-3 > 1 && len(expr.Children) == 1 && expr.Children[0].Kind != parser.ExpressionNode {
			return nil, types.Error{Range: expr.Range, Simple: "Syntax error", Detail: "While requires body to be contained in expression"}
		}
		if err != nil {
			return nil, err
		}
		whileStmt.Body = append(whileStmt.Body, exprAst)
	}

	return whileStmt, nil
}

func (constructor *AstConstructor) createIfExpr(node parser.Node) (Expr, error) {
	if len(node.Children) != 4 && len(node.Children) != 3 {
		return nil, types.Error{Range: node.Range,
			Simple: "Syntax error for if-else",
			Detail: fmt.Sprintf("Expected 3 or 5 children for if, got %d", len(node.Children)),
		}
	}
	cond, err := constructor.createAstExpression(node.Children[1])
	if err != nil {
		return nil, err
	}
	ifBranch, err := constructor.createBody(node.Children[2])
	if err != nil {
		return nil, err
	}

	if len(node.Children) == 4 {
		elseBranch, err := constructor.createBody(node.Children[3])
		if err != nil {
			return nil, err
		}
		return IfElseExpr{Condition: cond, IfBranch: ifBranch, ElseBranch: elseBranch, Range: node.Range}, nil
	} else {
		return IfOnlyExpr{Condition: cond, IfBranch: ifBranch, Range: node.Range}, nil
	}
}

func (constructor *AstConstructor) createBody(node parser.Node) ([]Ast, error) {
	if len(node.Children) > 0 && len(node.Children[0].Children) == 1 && node.Children[0].Children[0].Kind != parser.ExpressionNode {
		// e.g. (+ 10 4)
		ast, err := constructor.CreateAstItem(node)
		if err != nil {
			return []Ast{}, err
		}
		return []Ast{ast}, nil
	} else {
		return constructor.createAst(node, false)
	}
}

// Create an (closure) application expression
// The first node must be a value (e.g. in-place closure declaration or variable containing a closure)
// The remaining nodes are expressions that resolve to arguments
func (constructor AstConstructor) createAppExpr(exprParts []parser.Node, appRange types.FileRange) (ClosureApplicationExpr, error) {
	if len(exprParts) == 0 {
		return ClosureApplicationExpr{}, types.Error{Range: appRange, Simple: "Syntax error",
			Detail: "Closure application must have a value, got 0-length s-expr"}
	}
	val, err := constructor.createAstExpression(exprParts[0])
	if err != nil {
		return ClosureApplicationExpr{}, err
	}
	closureApp := ClosureApplicationExpr{Range: appRange, Closure: val, Args: make([]Expr, len(exprParts)-1)}
	for i, exprNode := range exprParts[1:] {
		arg, err := constructor.createAstExpression(exprNode)
		if err != nil {
			return ClosureApplicationExpr{}, err
		}
		closureApp.Args[i] = arg
	}
	return closureApp, nil
}

func (constructor *AstConstructor) createFuncAppExpr(node parser.Node) (Expr, error) {
	funcNameExpr, err := singleNestedExpr(node.Children[0])
	if err == nil && funcNameExpr.Kind == parser.LiteralNode {
		appExpr := FunctionApplicationExpr{Identifier: funcNameExpr.Data, Args: make([]Expr, len(node.Children)-1), Range: node.Range}
		for i, argNode := range node.Children[1:] {
			argExpr, err := constructor.createAstExpression(argNode)
			if err != nil {
				return nil, types.Error{
					Simple: "Function application expression must an expression",
					Range:  argNode.Range}
			}
			appExpr.Args[i] = argExpr
		}
		return appExpr, nil
	}
	return nil, types.Error{Simple: "Parse error", Detail: "bad function application", Range: node.Range}
}

func (constructor *AstConstructor) createClosure(node parser.Node) (ClosureDefExpr, error) {
	if len(node.Children) < 3 {
		return ClosureDefExpr{}, types.Error{Range: node.Range,
			Simple: "Syntax error whilst declaring closure",
			Detail: fmt.Sprintf("Expected at least 3 child nodes for closure, got %d", len(node.Children))}
	}
	closure := ClosureDefExpr{Args: make([]string, len(node.Children[1].Children))}
	for i, argExpr := range node.Children[1].Children {
		if len(argExpr.Children) != 1 || argExpr.Children[0].Kind != parser.LiteralNode {
			return ClosureDefExpr{}, types.Error{Range: node.Range,
				Simple: "Syntax error - argument name must be a string",
				Detail: fmt.Sprintf("Closure argument %d type is %s", i+1, argExpr.Kind)}
		}
		closure.Args[i] = argExpr.Children[0].Data
	}
	body, err := constructor.createFunctionBody(node.Children[2:])
	if err != nil {
		return ClosureDefExpr{}, err
	}
	closure.Body = body
	return closure, nil
}

func (constructor *AstConstructor) createAstStatement(node parser.Node, isRoot bool) (Stmt, error) {
	ok, literal := nestedLiteralValue(node)
	if !ok {
		return nil, types.Error{
			Simple: "Parse error",
			Detail: "could not find nested literal",
			Range:  node.Range}
	}
	if literal == "def" {
		if len(node.Children) > 2 && node.Children[1].Kind == parser.ExpressionNode && len(node.Children[1].Children) > 0 &&
			node.Children[1].Children[0].Kind == parser.AccessorNode {
			return constructor.createStructFieldDeclaration(node)
		} else {
			return constructor.createVariableDeclaration(node, isRoot)
		}
	} else if literal == "defun" {
		return constructor.createFunctionDeclaration(node, isRoot)
	} else if literal == "while" {
		return constructor.createWhileLoop(node)
	} else if literal == "import" {
		return constructor.handleImport(node)
	} else if literal == "defstruct" {
		return constructor.createStructDeclaration(node)
	}
	return nil, types.Error{
		Simple: "Parse error",
		Detail: fmt.Sprintf("Failed to process node - %s", node.Kind),
		Range:  node.Range,
	}
}

func (constructor *AstConstructor) createStructDeclaration(node parser.Node) (StructDefStmt, error) {
	// (defstruct <name> <fields>...)
	if len(node.Children) < 2 {
		return StructDefStmt{}, types.Error{Range: node.Range, Simple: "Invalid defstruct - expected (defstruct <name> <fields>...)"}
	}
	if len(node.Children[1].Children) != 1 || node.Children[1].Children[0].Kind != parser.LiteralNode {
		return StructDefStmt{}, types.Error{
			Simple: "Invalid struct declaration - name must be a literal",
			Range:  node.Children[1].Range,
		}
	}
	ident := node.Children[1].Children[0].Data
	fieldNames := []string{}
	for _, fieldNode := range node.Children[2:] {
		if len(fieldNode.Children) != 1 || fieldNode.Children[0].Kind != parser.LiteralNode {
			return StructDefStmt{}, types.Error{
				Simple: "Bad struct field name - expected identifier",
				Range:  fieldNode.Range,
			}
		}
		fieldNames = append(fieldNames, fieldNode.Children[0].Data)
	}

	return StructDefStmt{Identifier: ident, FieldNames: fieldNames, Range: node.Range}, nil
}

func (constructor *AstConstructor) createFunctionDeclaration(node parser.Node, isRoot bool) (FuncDefStmt, error) {
	// (defun identifier (args) definition)
	if len(node.Children) < 4 {
		return FuncDefStmt{}, types.Error{
			Simple: "Syntax error - function declaration should take form (defun <name> <args> <body>)",
			Detail: fmt.Sprintf("invalid function declaration syntax - expected 4  children, got %d", len(node.Children)),
			Range:  node.Range,
		}
	}
	if len(node.Children[1].Children) != 1 || node.Children[1].Children[0].Kind != parser.LiteralNode {
		return FuncDefStmt{}, types.Error{
			Simple: "Invalid function declaration - name must be a literal",
			Range:  node.Children[1].Range,
		}
	}
	if !isRoot {
		return FuncDefStmt{}, types.Error{Range: node.Range,
			Simple: fmt.Sprintf("Invalid function declaration %s - functions can only be declared at top level. Use a closure instead", node.Children[1].Children[0].Data),
		}
	}

	funcDefStmt := FuncDefStmt{Identifier: node.Children[1].Children[0].Data, Args: make([]string, 0), Body: make([]Ast, 0), Range: node.Range}
	if _, ok := constructor.Functions[funcDefStmt.Identifier]; ok && !constructor.AllowFunctionRedeclaration {
		return FuncDefStmt{}, types.Error{
			Simple: fmt.Sprintf("Duplicate declaration of function %s", funcDefStmt.Identifier),
			Range:  node.Range,
		}
	}

	argNode := node.Children[2]
	for _, argExpr := range argNode.Children {
		if len(argExpr.Children) != 1 || argExpr.Children[0].Kind != parser.LiteralNode {
			return FuncDefStmt{}, types.Error{
				Simple: "Bad function argument - expected identifier",
				Range:  argExpr.Range,
			}
		}
		funcDefStmt.Args = append(funcDefStmt.Args, argExpr.Children[0].Data)
	}

	body, err := constructor.createFunctionBody(node.Children[3:])
	if err != nil {
		return FuncDefStmt{}, err
	}
	funcDefStmt.Body = body

	constructor.Functions[funcDefStmt.Identifier] = &funcDefStmt
	return funcDefStmt, nil

}

func (constructor *AstConstructor) createVariableDeclaration(node parser.Node, isRoot bool) (VarDefStmt, error) {
	if len(node.Children) != 3 {
		return VarDefStmt{}, types.Error{
			Simple: "Syntax error - variable declaration should take form (def <name> <value>)",
			Detail: fmt.Sprintf("invalid variable declaration syntax - expected 3 expression children, got %d", len(node.Children)),
			Range:  node.Range,
		}
	}
	if len(node.Children[1].Children) != 1 || node.Children[1].Children[0].Kind != parser.LiteralNode {
		return VarDefStmt{}, types.Error{Simple: "Parse error - variable name must be literal", Range: node.Children[1].Range}
	}
	varValue, err := constructor.createAstExpression(node.Children[2])
	if err != nil {
		return VarDefStmt{}, types.Error{Simple: "Invalid variable assignment - variable assigned to statement",
			Detail: err.Error(),
			Range:  node.Children[2].Range}
	}
	varAst, err := VarDefStmt{Identifier: node.Children[1].Children[0].Data, Value: varValue, Range: node.Range}, nil
	if isRoot && err == nil {
		constructor.GlobalVariables[varAst.Identifier] = &varAst
	}
	return varAst, err
}

func (constructor *AstConstructor) createStructFieldDeclaration(node parser.Node) (StructFieldDeclarationStmt, error) {
	if len(node.Children) != 3 {
		return StructFieldDeclarationStmt{}, types.Error{
			Simple: "Syntax error - struct field declaration should take form (def <structLiteral>:<fieldName> <value>)",
			Range:  node.Range,
		}
	}
	if !(len(node.Children) > 2 && node.Children[1].Kind == parser.ExpressionNode && len(node.Children[1].Children) > 0 &&
		node.Children[1].Children[0].Kind == parser.AccessorNode) {
		return StructFieldDeclarationStmt{}, types.Error{Range: node.Range, Simple: "Syntax error - struct field declaration should take form (def <structLiteral>:<fieldName> <value>)"}
	}
	accessorNode := node.Children[1].Children[0]
	if len(accessorNode.Children) != 2 || accessorNode.Children[0].Kind != parser.LiteralNode || accessorNode.Children[1].Kind != parser.LiteralNode {
		return StructFieldDeclarationStmt{}, types.Error{Range: node.Range, Simple: "Syntax error - struct field declaration should take form (def <structLiteral>:<fieldName> <value>)"}
	}

	varValue, err := constructor.createAstExpression(node.Children[2])
	if err != nil {
		return StructFieldDeclarationStmt{}, err
	}
	return StructFieldDeclarationStmt{Range: node.Range, Value: varValue,
		StructIdentifier: accessorNode.Children[0].Data,
		FieldIdentifier:  accessorNode.Children[1].Data}, nil
}

func (constructor *AstConstructor) handleImport(node parser.Node) (Stmt, error) {
	importAst := Import{Range: node.Range}

	path, ok := safeTraverse(node, []int{1, 0})
	if !ok {
		return nil, types.Error{Range: node.Range, Simple: "Invalid import - must include path as second element in statement"}
	}
	importAst.Path = path.Data

	qualifier, ok := safeTraverse(node, []int{2, 0})
	qualifierString := ""
	if ok {
		qualifierString = qualifier.Data
	}
	importAst.Qualifier = qualifierString
	constructor.Imports = append(constructor.Imports, importAst)
	return ImportStmt{Range: node.Range}, nil
}

// Function body
// We allow for a single direct value e.g. (defun f () 1)
// But anything else must be in its own expression
func (constructor *AstConstructor) createFunctionBody(bodyNodes []parser.Node) ([]Ast, error) {
	body := make([]Ast, 0)
	for _, expr := range bodyNodes {
		exprAst, err := constructor.CreateAstItem(expr)
		if len(bodyNodes) > 1 && len(expr.Children) == 1 && expr.Children[0].Kind != parser.ExpressionNode {
			return nil, types.Error{Range: expr.Range, Simple: "Syntax error", Detail: "Function requires body to be contained in expression"}
		}
		if err != nil {
			return nil, err
		}
		body = append(body, exprAst)
	}
	return body, nil
}
