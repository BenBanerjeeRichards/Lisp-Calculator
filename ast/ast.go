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

func (VarDefStmt) stmtType() {}

func (ast *Ast) newStatement(stmt Stmt) {
	ast.Kind = StmtType
	ast.Statement = stmt
}

func (ast *Ast) newExpression(expr Expr) {
	ast.Kind = ExprType
	ast.Expression = expr
}

func CreateAst(expr parser.Node) (Ast, error) {
	if expr.Kind != parser.ExpressionNode {
		return Ast{}, errors.New("expected ExpressionNode when creating AST")
	}
	return doCreateAst(expr)
}

func doCreateAst(node parser.Node) (Ast, error) {
	if node.Children[0].Data == "def" && node.Kind == parser.ExpressionNode {
		varDefStmt, err := createAstStatement(node)
		if err != nil {
			return Ast{}, nil
		}
		ast := Ast{}
		ast.newStatement(varDefStmt)
		return ast, nil
	}

	// Everything else must be an expression
	expr, err := createAstExpression(node)
	if err != nil {
		return Ast{}, err
	}
	ast := Ast{}
	ast.newExpression(expr)
	return ast, nil
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

func createAstExpression(node parser.Node) (Expr, error) {
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
		if len(node.Children) == 1 {
			return createAstExpression(node.Children[0])
		}
		// Function application
		funcNameExpr, err := singleNestedExpr(node.Children[0])
		if err == nil && funcNameExpr.Kind == parser.LiteralNode {
			appExpr := FuncAppExpr{Identifier: funcNameExpr.Data, Args: make([]Expr, len(node.Children)-1)}
			for i, argNode := range node.Children[1:] {
				argExpr, err := createAstExpression(argNode)
				if err != nil {
					return nil, errors.New("function application argument must be an expression")
				}
				appExpr.Args[i] = argExpr
			}
			return appExpr, nil
		} else {
			return nil, fmt.Errorf("unknown expression type with node kind %s", node.Children[0].Kind)
		}
	}
	return nil, fmt.Errorf("unknown node type %s", node.Kind)
}

func createAstStatement(node parser.Node) (Stmt, error) {
	if node.Children[0].Data == "def" {
		if len(node.Children) != 3 {
			return nil, errors.New("invalid variable declaration syntax")
		}
		if node.Children[1].Kind != parser.LiteralNode {
			return nil, errors.New("invalid variable name")
		}
		varValue, err := createAstExpression(node.Children[2])
		if err != nil {
			return nil, fmt.Errorf("variable value must be an expression - %w", err)
		}
		return VarDefStmt{Identifier: node.Children[1].Data, Value: varValue}, nil
	}
	return nil, errors.New("could not create statement")
}
