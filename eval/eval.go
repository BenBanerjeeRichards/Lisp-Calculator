package eval

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"

	"github.com/benbanerjeerichards/lisp-calculator/ast"
	"github.com/benbanerjeerichards/lisp-calculator/parser"
	"github.com/benbanerjeerichards/lisp-calculator/util"
)

type Env struct {
	Variables map[string]float64
	Functions map[string]ast.FuncDefStmt
}

func (env *Env) New() {
	env.Variables = make(map[string]float64)
	env.Functions = make(map[string]ast.FuncDefStmt)
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
		if i == len(asts)-1 {
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
	case ast.FuncDefStmt:
		env.Functions[stmtNode.Identifier] = stmtNode
	default:
		return fmt.Errorf("unknown statement type %T", node)
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
			return 0, fmt.Errorf("undeclared variable %s", exprNode.Identifier)
		}
	case ast.FuncAppExpr:
		// First look up in function defintions, then try builtins
		if funcDef, ok := env.Functions[exprNode.Identifier]; ok {
			if len(funcDef.Args) != len(exprNode.Args) {
				return 0, fmt.Errorf("bad funtion application - expected %d arguments but recieved %d", len(funcDef.Args), len(exprNode.Args))
			}
			funcAppEnv := Env{}
			funcAppEnv.New()
			funcAppEnv.Functions = env.Functions
			for i, argName := range funcDef.Args {
				argExpr := exprNode.Args[i]
				argEvalValue, err := evalExpr(argExpr, env)
				if err != nil {
					return 0, fmt.Errorf("failed to eval argument %d - %v", i+1, err)
				}

				funcAppEnv.Variables[argName] = argEvalValue
			}
			for i, funcDefAst := range funcDef.Body {
				evalResult, err := Eval(funcDefAst, &funcAppEnv)
				if err != nil {
					return 0, err
				}
				if i == len(funcDef.Body)-1 {
					if !evalResult.HasValue {
						// TODO define this properly - should be a parse error
						return 0, errors.New("can not use function as expression as it ends with a statement")
					}
					return evalResult.Value, nil
				}
			}
		}
		switch exprNode.Identifier {
		case "+":
			if len(exprNode.Args) != 2 {
				return 0, errors.New("binary funtion add requires two parameters")
			}
			return builtInBinaryOp(func(f1, f2 float64) float64 { return f1 + f2 }, exprNode.Args[0], exprNode.Args[1], env)
		case "-":
			if len(exprNode.Args) != 2 {
				return 0, errors.New("binary funtion add requires two parameters")
			}
			return builtInBinaryOp(func(f1, f2 float64) float64 { return f1 - f2 }, exprNode.Args[0], exprNode.Args[1], env)
		case "*":
			if len(exprNode.Args) != 2 {
				return 0, errors.New("binary funtion add requires two parameters")
			}
			return builtInBinaryOp(func(f1, f2 float64) float64 { return f1 * f2 }, exprNode.Args[0], exprNode.Args[1], env)
		case "/":
			if len(exprNode.Args) != 2 {
				return 0, errors.New("binary funtion add requires two parameters")
			}
			return builtInBinaryOp(func(f1, f2 float64) float64 { return f1 / f2 }, exprNode.Args[0], exprNode.Args[1], env)
		case "^":
			if len(exprNode.Args) != 2 {
				return 0, errors.New("binary funtion add requires two parameters")
			}
			return builtInBinaryOp(func(f1, f2 float64) float64 { return math.Pow(f1, f2) }, exprNode.Args[0], exprNode.Args[1], env)
		case "log":
			if len(exprNode.Args) != 2 {
				return 0, errors.New("binary funtion add requires two parameters")
			}
			return builtInBinaryOp(func(f1, f2 float64) float64 { return math.Log(f2) / math.Log(f1) }, exprNode.Args[0], exprNode.Args[1], env)
		case "sqrt":
			if len(exprNode.Args) != 1 {
				return 0, errors.New("unary funtion add requires two parameters")
			}
			sqrtOf, err := evalExpr(exprNode.Args[0], env)
			if err != nil {
				return 0, err
			}
			return math.Sqrt(sqrtOf), nil
		default:
			return 0, fmt.Errorf("unknown function %s", exprNode.Identifier)
		}
	default:
		return 0, fmt.Errorf("unknown expression %v", node)
	}
}

func RunRepl() {
	reader := bufio.NewReader(os.Stdin)
	env := Env{}
	env.New()
	astConstruct := ast.AstConstructor{}
	astConstruct.New()

	for {
		fmt.Print("calc> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin: ", err)
			break
		}
		text = text[:len(text)-1]
		tokens := parser.Tokenise(text)
		parser := parser.Parser{}
		parser.New(tokens)
		expr, err := parser.ParseExpression()
		if err != nil {
			fmt.Println("Parse Error: ", err.Error())
			continue
		}
		util.WriteToFile("syntax.dot", util.ParseTreeToDot(expr))
		ast, err := astConstruct.CreateExpressionAst(expr)
		if err != nil {
			fmt.Println("Ast Error: ", err)
			continue
		}
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
