package ast

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/benbanerjeerichards/lisp-calculator/parser"
)

// AST node structure based on that used in the go compiler
// https://github.com/golang/go/blob/master/src/go/ast/ast.go

type Expr interface {
	exprType()
}

type Stmt interface {
	stmtType()
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
}

type FuncDefStmt struct {
	Identifier string
	Args       []string
	Body       []Ast
}

type VarUseExpr struct {
	Identifier string
}

type FuncAppExpr struct {
	Identifier string
	Args       []Expr
}

type NumberExpr struct {
	Value float64
}

// These are just to prevent assigning a statement to an expression
// Same as what go compiler does
func (FuncAppExpr) exprType() {}
func (NumberExpr) exprType()  {}
func (VarUseExpr) exprType()  {}

func (VarDefStmt) stmtType()  {}
func (FuncDefStmt) stmtType() {}

func (ast *Ast) newStatement(stmt Stmt) {
	ast.Kind = StmtType
	ast.Statement = stmt
}

func (ast *Ast) newExpression(expr Expr) {
	ast.Kind = ExprType
	ast.Expression = expr
}

type AstConstructor struct {
	FunctionNames map[string]bool
}

func (constructor *AstConstructor) New() {
	constructor.FunctionNames = make(map[string]bool)
	constructor.FunctionNames["add"] = true
	constructor.FunctionNames["mul"] = true
	constructor.FunctionNames["div"] = true
	constructor.FunctionNames["sub"] = true
	constructor.FunctionNames["log"] = true
	constructor.FunctionNames["pow"] = true
	constructor.FunctionNames["sqrt"] = true
}

func (constructor *AstConstructor) CreateAst(expr parser.Node) ([]Ast, error) {
	asts := make([]Ast, 0)
	// Keep track of declared functions. This is needed to differentiate between function application
	// and variable usage
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
	if ok, val := nestedLiteralValue(node); ok && (val == "def" || val == "defun") {
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
			return nil, errors.New("failed to parse as float")
		}
		return NumberExpr{Value: f}, nil
	case parser.LiteralNode:
		return VarUseExpr{Identifier: node.Data}, nil
	case parser.ExpressionNode:
		if len(node.Children) == 0 {
			return nil, errors.New("expression must have non-zero children")
		}
		// Determine if we have a function or variable declaration
		if litNode, ok := safeTraverse(node, []int{0, 0}); ok {
			if _, isFunc := constructor.FunctionNames[litNode.Data]; isFunc {
				return constructor.createFuncAppExpr(node)
			}
		}
		if len(node.Children) == 1 {
			return constructor.createAstExpression(node.Children[0])
		} else {
			return nil, fmt.Errorf("unknown expression type with node kind %s", node.Children[0].Kind)
		}
	}
	return nil, fmt.Errorf("unknown node type %s", node.Kind)
}

func (constructor *AstConstructor) createFuncAppExpr(node parser.Node) (Expr, error) {
	funcNameExpr, err := singleNestedExpr(node.Children[0])
	if err == nil && funcNameExpr.Kind == parser.LiteralNode {
		appExpr := FuncAppExpr{Identifier: funcNameExpr.Data, Args: make([]Expr, len(node.Children)-1)}
		for i, argNode := range node.Children[1:] {
			argExpr, err := constructor.createAstExpression(argNode)
			if err != nil {
				return nil, errors.New("function application argument must be an expression")
			}
			appExpr.Args[i] = argExpr
		}
		return appExpr, nil
	}
	return nil, errors.New("invalid function application")
}

func (constructor *AstConstructor) createAstStatement(node parser.Node) (Stmt, error) {
	ok, literal := nestedLiteralValue(node)
	if !ok {
		return nil, errors.New("could not create statement")
	}
	if literal == "def" {
		if len(node.Children) != 3 {
			return nil, fmt.Errorf("invalid variable declaration syntax - expected 3 expression children, got %d", len(node.Children))
		}
		if len(node.Children[1].Children) != 1 || node.Children[1].Children[0].Kind != parser.LiteralNode {
			return nil, errors.New("invalid variable name")
		}
		varValue, err := constructor.createAstExpression(node.Children[2])
		if err != nil {
			return nil, fmt.Errorf("variable value must be an expression - %w", err)
		}
		return VarDefStmt{Identifier: node.Children[1].Children[0].Data, Value: varValue}, nil
	} else if literal == "defun" {
		// (defun identifier (args) definition)
		if len(node.Children) < 4 {
			return nil, errors.New("invalid function declaration syntax")
		}
		if len(node.Children[1].Children) != 1 || node.Children[1].Children[0].Kind != parser.LiteralNode {
			return nil, errors.New("invalid function name name")
		}

		funcDefExpr := FuncDefStmt{Identifier: node.Children[1].Children[0].Data, Args: make([]string, 0), Body: make([]Ast, 0)}
		if _, ok = constructor.FunctionNames[funcDefExpr.Identifier]; ok {
			return nil, fmt.Errorf("duplicate declaration of function %s", funcDefExpr.Identifier)
		}
		constructor.FunctionNames[funcDefExpr.Identifier] = true

		argNode := node.Children[2]
		for _, argExpr := range argNode.Children {
			if len(argExpr.Children) != 1 || argExpr.Children[0].Kind != parser.LiteralNode {
				return nil, errors.New("bad argument for function definition")
			}
			funcDefExpr.Args = append(funcDefExpr.Args, argExpr.Children[0].Data)
		}

		for _, expr := range node.Children[3:] {
			exprAst, err := constructor.CreateExpressionAst(expr)
			if err != nil {
				return nil, err
			}
			funcDefExpr.Body = append(funcDefExpr.Body, exprAst)
		}

		return funcDefExpr, nil
	}
	return nil, errors.New("could not create statement")
}
