package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/benbanerjeerichards/lisp-calculator/ast"
	"github.com/benbanerjeerichards/lisp-calculator/parser"
	"github.com/benbanerjeerichards/lisp-calculator/util"
)

func evalLiteral(node parser.Node) (string, error) {
	switch node.Kind {
	case parser.LiteralNode:
		return node.Data, nil
	case parser.ExpressionNode:
		if len(node.Children) != 1 {
			return "", errors.New("can not obtain literal from n-ary expression")
		}
		return evalLiteral(node.Children[0])
	default:
		return "", fmt.Errorf("can not obtain literal from not type %s", node.Kind)
	}
}

func Eval(node parser.Node) (float64, error) {
	switch node.Kind {
	case parser.NumberNode:
		f, err := strconv.ParseFloat(node.Data, 64)
		if err != nil {
			return 0, errors.New("failed to parse as float")
		}
		return f, nil
	case parser.LiteralNode:
		// TODO implement definitions
		return 0, errors.New("variables not yet implemented")
	case parser.ExpressionNode:
		if len(node.Children) == 0 {
			return 0, errors.New("invalid expression - must have non-zero child expressions")
		}
		if len(node.Children) == 1 {
			return Eval(node.Children[0])
		}
		literal, err := evalLiteral(node.Children[0])
		if err != nil {
			return 0, errors.New("first argument to expression operation should be a literal")
		}
		switch literal {
		case "add":
			if len(node.Children) != 3 {
				return 0, errors.New("expected two operands to binary operation add")
			}
			lhs, err := Eval(node.Children[1])
			if err != nil {
				return 0, err
			}
			rhs, err := Eval(node.Children[2])
			if err != nil {
				return 0, err
			}
			return lhs + rhs, nil
		}
	default:
		return 0, errors.New("unknown node")
	}
	return 0, nil
}

type Env struct {
	Variables map[string]float64
}

func (env *Env) New() {
	env.Variables = make(map[string]float64)
}

type EvalResult struct {
	HasValue bool
	Value    float64
}

func EvalAst(astNode ast.Ast, env *Env) (EvalResult, error) {
	if astNode.Kind == ast.StmtType {
		err := EvalAstStmt(astNode.Statement, env)
		if err != nil {
			return EvalResult{}, err
		}
		return EvalResult{HasValue: false}, nil
	}
	if astNode.Kind == ast.ExprType {
		val, err := EvalAstExpr(astNode.Expression, *env)
		if err != nil {
			return EvalResult{}, err
		}
		return EvalResult{HasValue: true, Value: val}, nil
	}
	return EvalResult{}, errors.New("?")
}

func EvalAstStmt(node ast.Stmt, env *Env) error {
	switch stmtNode := node.(type) {
	case ast.VarDefStmt:
		result, err := EvalAstExpr(stmtNode.Value, *env)
		if err != nil {
			return err
		}
		env.Variables[stmtNode.Identifier] = result
	default:
		return fmt.Errorf("unknown statement type ", node)
	}
	return nil
}

func BuiltInBinaryOp(f func(float64, float64) float64, lhs ast.Expr, rhs ast.Expr, env Env) (float64, error) {
	lhsValue, err := EvalAstExpr(lhs, env)
	if err != nil {
		return 0, err
	}
	rhsValue, err := EvalAstExpr(rhs, env)
	if err != nil {
		return 0, err
	}
	return f(lhsValue, rhsValue), nil
}

func EvalAstExpr(node ast.Expr, env Env) (float64, error) {
	switch exprNode := node.(type) {
	case ast.NumberExpr:
		return exprNode.Value, nil
	case ast.VarUseExpr:
		if val, ok := env.Variables[exprNode.Identifier]; ok {
			return val, nil
		} else {
			return 0, fmt.Errorf("undeclared variable", exprNode.Identifier)
		}
	case ast.FuncAppExpr:
		switch exprNode.Identifier {
		case "add":
			if len(exprNode.Args) != 2 {
				return 0, errors.New("binary funtion add requires two parameters")
			}
			return BuiltInBinaryOp(func(f1, f2 float64) float64 { return f1 + f2 }, exprNode.Args[0], exprNode.Args[1], env)
		case "sub":
			if len(exprNode.Args) != 2 {
				return 0, errors.New("binary funtion add requires two parameters")
			}
			return BuiltInBinaryOp(func(f1, f2 float64) float64 { return f1 - f2 }, exprNode.Args[0], exprNode.Args[1], env)
		case "mul":
			if len(exprNode.Args) != 2 {
				return 0, errors.New("binary funtion add requires two parameters")
			}
			return BuiltInBinaryOp(func(f1, f2 float64) float64 { return f1 * f2 }, exprNode.Args[0], exprNode.Args[1], env)
		case "div":
			if len(exprNode.Args) != 2 {
				return 0, errors.New("binary funtion add requires two parameters")
			}
			return BuiltInBinaryOp(func(f1, f2 float64) float64 { return f1 / f2 }, exprNode.Args[0], exprNode.Args[1], env)
		default:
			return 0, fmt.Errorf("unknown function", exprNode.Identifier)
		}
	default:
		return 0, fmt.Errorf("unknown expression", node)
	}
}

func RunRepl() {
	reader := bufio.NewReader(os.Stdin)
	env := Env{}
	env.New()

	for {
		fmt.Print("calc> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin: ", err)
			break
		}
		tokens := parser.Tokenise(text)
		parser := parser.Parser{}
		parser.New(tokens)
		expr, err := parser.ParseExpression()
		if err != nil {
			fmt.Println("Parse Error: ", err.Error())
			continue
		}
		ast, err := ast.CreateAst(expr)
		if err != nil {
			fmt.Println("Ast Error: ", err)
			continue
		}
		fmt.Println(ast)

		val, err := EvalAst(ast, &env)
		if err != nil {
			fmt.Println("Eval Error: ", err.Error())
			continue
		}
		if val.HasValue {
			fmt.Println(val.Value)
		}
	}
}

func main() {
	RunRepl()
	tokens := parser.Tokenise("(add (add 14 10) 2)")
	// tokens := tokenise("(define r (add 55 23))(print (plus 10 r))")

	parser := parser.Parser{}
	parser.New(tokens)
	node, err := parser.ParseExpression()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(node)
	}

	util.WriteToFile("syntax.dot", util.ParseTreeToDot(node))

	ast, err := ast.CreateAst(node)
	if err != nil {
		fmt.Println("Ast error: ", err)
	} else {
		fmt.Println(ast)
	}

	val, err := Eval(node)
	if err != nil {
		fmt.Println("Eval error occured", err.Error())
	} else {
		fmt.Println(val)
	}
}
