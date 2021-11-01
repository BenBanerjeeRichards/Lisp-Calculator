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
	Variables map[string]Value
	Functions map[string]ast.FuncDefStmt
}

func (env *Env) New() {
	env.Variables = make(map[string]Value)
	env.Functions = make(map[string]ast.FuncDefStmt)
}

const (
	NumType    = "num"
	BoolType   = "bool"
	StringType = "string"
	NullType   = "null"
	ListType   = "list"
)

type Value struct {
	Kind   string
	Num    float64
	Bool   bool
	String string
	List   []interface{}
}

func (v *Value) NewNum(value float64) {
	v.Kind = NumType
	v.Num = value
}

func (v *Value) NewString(value string) {
	v.Kind = StringType
	v.String = value
}

func (v *Value) NewBool(value bool) {
	v.Kind = BoolType
	v.Bool = value
}

func (v *Value) NewList(value []interface{}) {
	v.Kind = ListType
	v.List = value
}
func (v *Value) NewNull() {
	v.Kind = NullType
}

// Cant use Stringer interface due to name conflict
func (val Value) ToString() string {
	switch val.Kind {
	case NumType:
		return fmt.Sprint(val.Num)
	case StringType:
		return "\"" + val.String + "\""
	case BoolType:
		if val.Bool {
			return "true"
		}
		return "false"
	case NullType:
		return "null"
	default:
		return "Unknown type"
	}

}

type EvalResult struct {
	HasValue bool
	Value    float64
}

// Eval every ast and return the value of the final one
func EvalProgram(asts []ast.Ast) (Value, error) {
	if len(asts) == 0 {
		return Value{}, errors.New("eval program requires non-zero numbers of asts")
	}
	env := Env{}
	env.New()

	for i, ast := range asts {
		result, err := Eval(ast, &env)
		if err != nil {
			return Value{}, err
		}
		if i == len(asts)-1 {
			return result, nil
		}
	}
	return Value{}, errors.New("unknown error?")
}

func Eval(astNode ast.Ast, env *Env) (Value, error) {
	if astNode.Kind == ast.StmtType {
		err := evalStmt(astNode.Statement, env)
		if err != nil {
			return Value{}, err
		}

		return Value{Kind: NullType}, nil
	}
	if astNode.Kind == ast.ExprType {
		val, err := evalExpr(astNode.Expression, *env)
		if err != nil {
			return Value{}, err
		}
		return val, nil
	}
	return Value{}, errors.New("?")
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

func builtInBinaryOp(f func(float64, float64) float64, lhs ast.Expr, rhs ast.Expr, env Env) (Value, error) {
	lhsValue, err := evalExpr(lhs, env)
	if err != nil {
		return Value{}, err
	}
	rhsValue, err := evalExpr(rhs, env)
	if err != nil {
		return Value{}, err
	}
	if lhsValue.Kind != NumType {
		return Value{}, errors.New("type error for lhs to binary op")
	}
	if rhsValue.Kind != NumType {
		return Value{}, errors.New("type error for rhs to binary op")
	}
	val := Value{}
	val.NewNum(f(lhsValue.Num, rhsValue.Num))
	return val, nil
}

func builtInBinaryCompare(f func(float64, float64) bool, lhs ast.Expr, rhs ast.Expr, env Env) (Value, error) {
	lhsValue, err := evalExpr(lhs, env)
	if err != nil {
		return Value{}, err
	}
	rhsValue, err := evalExpr(rhs, env)
	if err != nil {
		return Value{}, err
	}
	if lhsValue.Kind != NumType {
		return Value{}, errors.New("type error for lhs to binary op")
	}
	if rhsValue.Kind != NumType {
		return Value{}, errors.New("type error for rhs to binary op")
	}
	val := Value{}
	val.NewBool(f(lhsValue.Num, rhsValue.Num))
	return val, nil
}

func evalExpr(node ast.Expr, env Env) (Value, error) {
	switch exprNode := node.(type) {
	case ast.NumberExpr:
		val := Value{}
		val.NewNum(exprNode.Value)
		return val, nil
	case ast.BoolExpr:
		val := Value{}
		val.NewBool(exprNode.Value)
		return val, nil
	case ast.VarUseExpr:
		if val, ok := env.Variables[exprNode.Identifier]; ok {
			return val, nil
		} else {
			return Value{}, fmt.Errorf("undeclared variable %s", exprNode.Identifier)
		}
	case ast.IfElseExpr:
		condRes, err := evalExpr(exprNode.Condition, env)
		if err != nil {
			return Value{}, err
		}
		if condRes.Kind != BoolType {
			return Value{}, errors.New("bad type for ifelse condition")
		}
		branch := exprNode.IfBranch
		if !condRes.Bool {
			branch = exprNode.ElseBranch
		}
		for i, ast := range branch {
			evalResult, err := Eval(ast, &env)
			if err != nil {
				return Value{}, err
			}
			if i == len(branch)-1 {
				return evalResult, nil
			}
		}
	case ast.IfOnlyExpr:
		condRes, err := evalExpr(exprNode.Condition, env)
		if err != nil {
			return Value{}, err
		}
		if condRes.Kind != BoolType {
			return Value{}, errors.New("bad type for ifelse condition")
		}
		if !condRes.Bool {
			val := Value{}
			val.NewNull()
			return val, nil
		}
		for i, ast := range exprNode.IfBranch {
			evalResult, err := Eval(ast, &env)
			if err != nil {
				return Value{}, err
			}
			if i == len(exprNode.IfBranch)-1 {
				return evalResult, nil
			}
		}
	case ast.FuncAppExpr:
		// First look up in function defintions, then try builtins
		if funcDef, ok := env.Functions[exprNode.Identifier]; ok {
			if len(funcDef.Args) != len(exprNode.Args) {
				return Value{}, fmt.Errorf("bad funtion application - expected %d arguments but recieved %d", len(funcDef.Args), len(exprNode.Args))
			}
			funcAppEnv := Env{}
			funcAppEnv.New()
			funcAppEnv.Functions = env.Functions
			for i, argName := range funcDef.Args {
				argExpr := exprNode.Args[i]
				argEvalValue, err := evalExpr(argExpr, env)
				if err != nil {
					return Value{}, fmt.Errorf("failed to eval argument %d - %v", i+1, err)
				}

				funcAppEnv.Variables[argName] = argEvalValue
			}
			for i, funcDefAst := range funcDef.Body {
				evalResult, err := Eval(funcDefAst, &funcAppEnv)
				if err != nil {
					return Value{}, err
				}
				if i == len(funcDef.Body)-1 {
					return evalResult, nil
				}
			}
		}
		switch exprNode.Identifier {
		case "+":
			if len(exprNode.Args) != 2 {
				return Value{}, errors.New("binary funtion add requires two parameters")
			}
			return builtInBinaryOp(func(f1, f2 float64) float64 { return f1 + f2 }, exprNode.Args[0], exprNode.Args[1], env)
		case "-":
			if len(exprNode.Args) != 2 {
				return Value{}, errors.New("binary funtion add requires two parameters")
			}
			return builtInBinaryOp(func(f1, f2 float64) float64 { return f1 - f2 }, exprNode.Args[0], exprNode.Args[1], env)
		case "*":
			if len(exprNode.Args) != 2 {
				return Value{}, errors.New("binary funtion add requires two parameters")
			}
			return builtInBinaryOp(func(f1, f2 float64) float64 { return f1 * f2 }, exprNode.Args[0], exprNode.Args[1], env)
		case "/":
			if len(exprNode.Args) != 2 {
				return Value{}, errors.New("binary funtion add requires two parameters")
			}
			return builtInBinaryOp(func(f1, f2 float64) float64 { return f1 / f2 }, exprNode.Args[0], exprNode.Args[1], env)
		case "^":
			if len(exprNode.Args) != 2 {
				return Value{}, errors.New("binary funtion add requires two parameters")
			}
			return builtInBinaryOp(func(f1, f2 float64) float64 { return math.Pow(f1, f2) }, exprNode.Args[0], exprNode.Args[1], env)
		case "log":
			if len(exprNode.Args) != 2 {
				return Value{}, errors.New("binary funtion add requires two parameters")
			}
			return builtInBinaryOp(func(f1, f2 float64) float64 { return math.Log(f2) / math.Log(f1) }, exprNode.Args[0], exprNode.Args[1], env)
		case "sqrt":
			if len(exprNode.Args) != 1 {
				return Value{}, errors.New("unary funtion add requires two parameters")
			}
			sqrtOf, err := evalExpr(exprNode.Args[0], env)
			if err != nil {
				return Value{}, err
			}
			if sqrtOf.Kind != NumType {
				return Value{}, errors.New("type error for arg to sqrt")
			}
			val := Value{}
			val.NewNum(math.Sqrt(sqrtOf.Num))
			return val, nil
		case ">":
			if len(exprNode.Args) != 2 {
				return Value{}, errors.New("binary funtion add requires two parameters")
			}
			return builtInBinaryCompare(func(f1, f2 float64) bool { return f1 > f2 }, exprNode.Args[0], exprNode.Args[1], env)
		case ">=":
			if len(exprNode.Args) != 2 {
				return Value{}, errors.New("binary funtion add requires two parameters")
			}
			return builtInBinaryCompare(func(f1, f2 float64) bool { return f1 >= f2 }, exprNode.Args[0], exprNode.Args[1], env)
		case "<":
			if len(exprNode.Args) != 2 {
				return Value{}, errors.New("binary funtion add requires two parameters")
			}
			return builtInBinaryCompare(func(f1, f2 float64) bool { return f1 < f2 }, exprNode.Args[0], exprNode.Args[1], env)
		case "<=":
			if len(exprNode.Args) != 2 {
				return Value{}, errors.New("binary funtion add requires two parameters")
			}
			return builtInBinaryCompare(func(f1, f2 float64) bool { return f1 <= f2 }, exprNode.Args[0], exprNode.Args[1], env)
		case "=":
			if len(exprNode.Args) != 2 {
				return Value{}, errors.New("binary funtion add requires two parameters")
			}
			return builtInBinaryCompare(func(f1, f2 float64) bool { return f1 == f2 }, exprNode.Args[0], exprNode.Args[1], env)
		default:
			return Value{}, fmt.Errorf("unknown function %s", exprNode.Identifier)
		}

	default:
		return Value{}, fmt.Errorf("unknown expression %v", node)
	}
	return Value{}, errors.New("?")
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
		fmt.Println(val.ToString())
	}
}
