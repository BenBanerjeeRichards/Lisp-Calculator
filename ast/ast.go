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

type Ast struct {
	Expression Expr
	Statement  Stmt
	Kind       string
}

type VarDefStmt struct {
	Identifier string
	Value      Expr
	Range      types.FileRange
}

type FuncDefStmt struct {
	Identifier string
	Args       []string
	Body       []Ast
	Range      types.FileRange
}

type VarUseExpr struct {
	Identifier string
	Range      types.FileRange
}

type FunctionApplicationExpr struct {
	Identifier string
	Args       []Expr
	Range      types.FileRange
}

type ClosureApplicationExpr struct {
	Closure Expr
	Args    []Expr
	Range   types.FileRange
}

type ClosureDefExpr struct {
	Args  []string
	Body  []Ast
	Range types.FileRange
}

type NumberExpr struct {
	Value float64
	Range types.FileRange
}

type StringExpr struct {
	Value string
	Range types.FileRange
}
type BoolExpr struct {
	Value bool
	Range types.FileRange
}

type NullExpr struct {
	Range types.FileRange
}

type ListExpr struct {
	Value []Expr
	Range types.FileRange
}

type IfElseExpr struct {
	Condition  Expr
	IfBranch   []Ast
	ElseBranch []Ast
	Range      types.FileRange
}

type IfOnlyExpr struct {
	Condition Expr
	IfBranch  []Ast
	Range     types.FileRange
}

type WhileStmt struct {
	Condition Expr
	Body      []Ast
	Range     types.FileRange
}

func (v VarDefStmt) GetRange() types.FileRange {
	return v.Range
}

func (v FuncDefStmt) GetRange() types.FileRange {
	return v.Range
}

func (v VarUseExpr) GetRange() types.FileRange {
	return v.Range
}

func (v FunctionApplicationExpr) GetRange() types.FileRange {
	return v.Range
}

func (v NumberExpr) GetRange() types.FileRange {
	return v.Range
}

func (v StringExpr) GetRange() types.FileRange {
	return v.Range
}

func (v BoolExpr) GetRange() types.FileRange {
	return v.Range
}

func (v NullExpr) GetRange() types.FileRange {
	return v.Range
}

func (v ListExpr) GetRange() types.FileRange {
	return v.Range
}

func (v IfElseExpr) GetRange() types.FileRange {
	return v.Range
}

func (v IfOnlyExpr) GetRange() types.FileRange {
	return v.Range
}

func (v WhileStmt) GetRange() types.FileRange {
	return v.Range
}

func (v ClosureDefExpr) GetRange() types.FileRange {
	return v.Range
}

func (v ClosureApplicationExpr) GetRange() types.FileRange {
	return v.Range
}

// These are just to prevent assigning a statement to an expression
// Same as what go compiler does
func (FunctionApplicationExpr) exprType() {}
func (NumberExpr) exprType()              {}
func (VarUseExpr) exprType()              {}
func (BoolExpr) exprType()                {}
func (IfElseExpr) exprType()              {}
func (IfOnlyExpr) exprType()              {}
func (StringExpr) exprType()              {}
func (ListExpr) exprType()                {}
func (NullExpr) exprType()                {}
func (ClosureDefExpr) exprType()          {}
func (ClosureApplicationExpr) exprType()  {}

func (VarDefStmt) stmtType()  {}
func (FuncDefStmt) stmtType() {}
func (WhileStmt) stmtType()   {}

func (ast *Ast) newStatement(stmt Stmt) {
	ast.Kind = StmtType
	ast.Statement = stmt
}

func (ast *Ast) newExpression(expr Expr) {
	// Keep track of declared functions. This is needed to differentiate between function application
	// and variable usage
	ast.Kind = ExprType
	ast.Expression = expr
}

type AstConstructor struct {
	FunctionNames map[string]bool
}

func (constructor *AstConstructor) New() {
	constructor.FunctionNames = make(map[string]bool)
	constructor.FunctionNames["+"] = true
	constructor.FunctionNames["*"] = true
	constructor.FunctionNames["/"] = true
	constructor.FunctionNames["-"] = true
	constructor.FunctionNames["log"] = true
	constructor.FunctionNames["^"] = true
	constructor.FunctionNames["sqrt"] = true
	constructor.FunctionNames[">"] = true
	constructor.FunctionNames[">="] = true
	constructor.FunctionNames["<"] = true
	constructor.FunctionNames["<="] = true
	constructor.FunctionNames["="] = true
	constructor.FunctionNames["print"] = true
}

func (constructor *AstConstructor) CreateAst(expr parser.Node) ([]Ast, error) {
	asts := make([]Ast, 0)
	for _, expression := range expr.Children {
		ast, err := constructor.CreateExpressionAst(expression)
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

func (constructor *AstConstructor) CreateExpressionAst(node parser.Node) (Ast, error) {
	if ok, val := nestedLiteralValue(node); ok && (val == "def" || val == "defun" || val == "while") {
		varDefStmt, err := constructor.createAstStatement(node)
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
		exprAst, err := constructor.CreateExpressionAst(expr)
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
		ast, err := constructor.CreateExpressionAst(node)
		if err != nil {
			return []Ast{}, err
		}
		return []Ast{ast}, nil
	} else {
		return constructor.CreateAst(node)
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

func (constructor *AstConstructor) createAstStatement(node parser.Node) (Stmt, error) {
	ok, literal := nestedLiteralValue(node)
	if !ok {
		return nil, types.Error{
			Simple: "Parse error",
			Detail: "could not find nested literal",
			Range:  node.Range}
	}
	if literal == "def" {
		if len(node.Children) != 3 {
			return nil, types.Error{
				Simple: "Syntax error - variable declaration should take form (def <name> <value>)",
				Detail: fmt.Sprintf("invalid variable declaration syntax - expected 3 expression children, got %d", len(node.Children)),
				Range:  node.Range,
			}
		}
		if len(node.Children[1].Children) != 1 || node.Children[1].Children[0].Kind != parser.LiteralNode {
			return nil, types.Error{Simple: "Parse error - variable name must be literal", Range: node.Children[1].Range}
		}
		varValue, err := constructor.createAstExpression(node.Children[2])
		if err != nil {
			return nil, types.Error{Simple: "Invalid variable assignment - variable assigned to statement",
				Detail: err.Error(),
				Range:  node.Children[2].Range}
		}
		return VarDefStmt{Identifier: node.Children[1].Children[0].Data, Value: varValue, Range: node.Range}, nil
	} else if literal == "defun" {
		// (defun identifier (args) definition)
		if len(node.Children) < 4 {
			return nil, types.Error{
				Simple: "Syntax error - function declaration should take form (defun <name> <args> <body>)",
				Detail: fmt.Sprintf("invalid function declaration syntax - expected 4  children, got %d", len(node.Children)),
				Range:  node.Range,
			}
		}
		if len(node.Children[1].Children) != 1 || node.Children[1].Children[0].Kind != parser.LiteralNode {
			return nil, types.Error{
				Simple: "Invalid function declaration - name must be a literal",
				Range:  node.Children[1].Range,
			}
		}

		funcDefExpr := FuncDefStmt{Identifier: node.Children[1].Children[0].Data, Args: make([]string, 0), Body: make([]Ast, 0), Range: node.Range}
		if _, ok = constructor.FunctionNames[funcDefExpr.Identifier]; ok {
			return nil, types.Error{
				Simple: fmt.Sprintf("Duplicate declaration of function %s", funcDefExpr.Identifier),
				Range:  node.Range,
			}
		}
		constructor.FunctionNames[funcDefExpr.Identifier] = true

		argNode := node.Children[2]
		for _, argExpr := range argNode.Children {
			if len(argExpr.Children) != 1 || argExpr.Children[0].Kind != parser.LiteralNode {
				return nil, types.Error{
					Simple: "Bad function argument - expected identifier",
					Range:  argExpr.Range,
				}
			}
			funcDefExpr.Args = append(funcDefExpr.Args, argExpr.Children[0].Data)
		}

		body, err := constructor.createFunctionBody(node.Children[3:])
		if err != nil {
			return nil, err
		}
		funcDefExpr.Body = body

		return funcDefExpr, nil
	} else if literal == "while" {
		return constructor.createWhileLoop(node)
	}
	return nil, types.Error{
		Simple: "Parse error",
		Detail: fmt.Sprintf("Failed to process node - %s", node.Kind),
		Range:  node.Range,
	}
}

// Function body
// We allow for a single direct value e.g. (defun f () 1)
// But anything else must be in its own expression
func (constructor *AstConstructor) createFunctionBody(bodyNodes []parser.Node) ([]Ast, error) {
	body := make([]Ast, 0)
	for _, expr := range bodyNodes {
		exprAst, err := constructor.CreateExpressionAst(expr)
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
