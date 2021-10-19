package eval

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/benbanerjeerichards/lisp-calculator/ast"
	"github.com/benbanerjeerichards/lisp-calculator/parser"
)

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

// Eval every ast and return the value of the final one
func EvalProgram(asts []ast.Ast) (EvalResult, error) {
	if len(asts) == 0 {
		return EvalResult{}, errors.New("eval program requires non-zero numbers of asts")
	}
	env := Env{}
	env.New()

	for i, ast := range asts {
		result, err := Eval(ast, &env)
		if err != nil {
			return EvalResult{}, err
		}
		if i == len(asts) - 1 {
			return result, nil
		}
	}
	return EvalResult{}, errors.New("unknown error?")
}

func Eval(astNode ast.Ast, env *Env) (EvalResult, error) {
	if astNode.Kind == ast.StmtType {
		err := evalStmt(astNode.Statement, env)
		if err != nil {
			return EvalResult{}, err
		}
		return EvalResult{HasValue: false}, nil
	}
	if astNode.Kind == ast.ExprType {
		val, err := evalExpr(astNode.Expression, *env)
		if err != nil {
			return EvalResult{}, err
		}
		return EvalResult{HasValue: true, Value: val}, nil
	}
	return EvalResult{}, errors.New("?")
}

func evalStmt(node ast.Stmt, env *Env) error {
	switch stmtNode := node.(type) {
	case ast.VarDefStmt:
		result, err := evalExpr(stmtNode.Value, *env)
		if err != nil {
			return err
		}
		env.Variables[stmtNode.Identifier] = result
	default:
		return fmt.Errorf("unknown statement type ", node)
	}
	return nil
}

func builtInBinaryOp(f func(float64, float64) float64, lhs ast.Expr, rhs ast.Expr, env Env) (float64, error) {
	lhsValue, err := evalExpr(lhs, env)
	if err != nil {
		return 0, err
	}
	rhsValue, err := evalExpr(rhs, env)
	if err != nil {
		return 0, err
	}
	return f(lhsValue, rhsValue), nil
}

func evalExpr(node ast.Expr, env Env) (float64, error) {
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
			return builtInBinaryOp(func(f1, f2 float64) float64 { return f1 + f2 }, exprNode.Args[0], exprNode.Args[1], env)
		case "sub":
			if len(exprNode.Args) != 2 {
				return 0, errors.New("binary funtion add requires two parameters")
			}
			return builtInBinaryOp(func(f1, f2 float64) float64 { return f1 - f2 }, exprNode.Args[0], exprNode.Args[1], env)
		case "mul":
			if len(exprNode.Args) != 2 {
				return 0, errors.New("binary funtion add requires two parameters")
			}
			return builtInBinaryOp(func(f1, f2 float64) float64 { return f1 * f2 }, exprNode.Args[0], exprNode.Args[1], env)
		case "div":
			if len(exprNode.Args) != 2 {
				return 0, errors.New("binary funtion add requires two parameters")
			}
			return builtInBinaryOp(func(f1, f2 float64) float64 { return f1 / f2 }, exprNode.Args[0], exprNode.Args[1], env)
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

		val, err := Eval(ast, &env)
		if err != nil {
			fmt.Println("Eval Error: ", err.Error())
			continue
		}
		if val.HasValue {
			fmt.Println(val.Value)
		}
	}
}
