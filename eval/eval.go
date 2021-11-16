package eval

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/benbanerjeerichards/lisp-calculator/ast"
	"github.com/benbanerjeerichards/lisp-calculator/parser"
	"github.com/benbanerjeerichards/lisp-calculator/types"
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
	NumType     = "num"
	BoolType    = "bool"
	StringType  = "string"
	NullType    = "null"
	ListType    = "list"
	ClosureType = "closure"
)

type ClosureValue struct {
	ClosureEnv Env
	Args       []string
	Body       []ast.Ast
}

type Value struct {
	Kind    string
	Num     float64
	Bool    bool
	String  string
	List    []Value
	Closure ClosureValue
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

func (v *Value) NewList(value []Value) {
	v.Kind = ListType
	v.List = value
}
func (v *Value) NewNull() {
	v.Kind = NullType
}
func (v *Value) NewClosure(args []string, body []ast.Ast, env Env) {
	v.Kind = ClosureType
	v.Closure = ClosureValue{Args: args, Body: body, ClosureEnv: env}
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
	case ListType:
		var listStrBuilder strings.Builder
		listStrBuilder.WriteString("(")
		for i, item := range val.List {
			listStrBuilder.WriteString(item.ToString())
			if i != len(val.List)-1 {
				listStrBuilder.WriteString(" ")
			}
		}
		listStrBuilder.WriteString(")")
		return listStrBuilder.String()
	case ClosureType:
		var argString strings.Builder
		for i, arg := range val.Closure.Args {
			argString.WriteString(arg)
			if i != len(val.Closure.Args)-1 {
				argString.WriteString(" ")
			}
		}
		return fmt.Sprintf("lambda(%s)", argString.String())
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
	case ast.WhileStmt:
		cond, err := evalExpr(stmtNode.Condition, *env)
		if err != nil {
			return err
		}
		if cond.Kind != BoolType {
			return types.Error{Range: stmtNode.Condition.GetRange(),
				Simple: fmt.Sprintf("Type Error - while loop condition is not a boolean (got %s)", cond.Kind)}
		}
		for cond.Bool {
			for _, ast := range stmtNode.Body {
				Eval(ast, env)
			}
			cond, err = evalExpr(stmtNode.Condition, *env)
			if err != nil {
				return err
			}
			if cond.Kind != BoolType {
				return types.Error{Range: stmtNode.Condition.GetRange(),
					Simple: fmt.Sprintf("Type Error - while loop condition is not a boolean (got %s)", cond.Kind)}
			}
		}
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
		return Value{}, types.Error{Range: lhs.GetRange(),
			Simple: fmt.Sprintf("Type Error - LHS expected number (got %s)", lhsValue.Kind)}

	}
	if rhsValue.Kind != NumType {
		return Value{}, types.Error{Range: rhs.GetRange(),
			Simple: fmt.Sprintf("Type Error - RHS expected number (got %s)", rhsValue.Kind)}
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
		return Value{}, types.Error{Range: rhs.GetRange(),
			Simple: fmt.Sprintf("Type Error - LHS  expected number (got %s)", lhsValue.Kind)}
	}
	if rhsValue.Kind != NumType {
		return Value{}, types.Error{Range: rhs.GetRange(),
			Simple: fmt.Sprintf("Type Error - RHS expected number (got %s)", rhsValue.Kind)}
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
	case ast.StringExpr:
		val := Value{}
		val.NewString(exprNode.Value)
		return val, nil
	case ast.BoolExpr:
		val := Value{}
		val.NewBool(exprNode.Value)
		return val, nil
	case ast.NullExpr:
		val := Value{}
		val.NewNull()
		return val, nil
	case ast.ListExpr:
		val := Value{Kind: ListType, List: make([]Value, len(exprNode.Value))}
		for i, expr := range exprNode.Value {
			itemValue, err := evalExpr(expr, env)
			if err != nil {
				return Value{}, err
			}
			val.List[i] = itemValue
		}
		return val, nil
	case ast.VarUseExpr:
		if val, ok := env.Variables[exprNode.Identifier]; ok {
			return val, nil
		} else {
			return Value{}, types.Error{Range: exprNode.Range,
				Simple: fmt.Sprintf("Undeclared variable %s", exprNode.Identifier)}
		}
	case ast.IfElseExpr:
		condRes, err := evalExpr(exprNode.Condition, env)
		if err != nil {
			return Value{}, err
		}
		if condRes.Kind != BoolType {
			return Value{}, types.Error{Range: exprNode.Condition.GetRange(),
				Simple: fmt.Sprintf("Type error - expected boolean for IfElse condition (got %s)", condRes.Kind)}
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
			return Value{}, types.Error{Range: exprNode.Condition.GetRange(),
				Simple: fmt.Sprintf("Type error - expected boolean for If condition (got %s)", condRes.Kind)}
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
	case ast.ClosureDefExpr:
		// Return value of closure with captured env
		value := Value{}
		closureEnv := Env{}
		closureEnv.New()
		for funcName, f := range env.Functions {
			closureEnv.Functions[funcName] = f
		}
		for varName, v := range env.Variables {
			closureEnv.Variables[varName] = v
		}
		value.NewClosure(exprNode.Args, exprNode.Body, closureEnv)
		return value, nil
	case ast.ClosureApplicationExpr:
		closureVal, err := evalExpr(exprNode.Closure, env)
		if err != nil {
			return Value{}, err
		}
		if closureVal.Kind != ClosureType {
			return Value{}, types.Error{Range: exprNode.Range, Simple: fmt.Sprintf("Can not apply arguments to type %s (expected closure)", closureVal.Kind)}
		}
		closure := closureVal.Closure
		if len(closure.Args) != len(exprNode.Args) {
			return Value{}, types.Error{Range: exprNode.Range, Simple: fmt.Sprintf("Expected %d arguments to closure applicatoin, got %d", len(closure.Args), len(exprNode.Args))}
		}
		return evalClosure(closure, exprNode.Args, env, exprNode.Range)
	case ast.FunctionApplicationExpr:
		// First look up in function defintions, then try builtins
		if funcDef, ok := env.Functions[exprNode.Identifier]; ok {
			if len(funcDef.Args) != len(exprNode.Args) {
				return Value{}, types.Error{Range: exprNode.Range,
					Simple: fmt.Sprintf("Bad funtion application - expected %d arguments but recieved %d", len(funcDef.Args), len(exprNode.Args))}
			}
			funcAppEnv := Env{}
			funcAppEnv.New()
			funcAppEnv.Functions = env.Functions
			for i, argName := range funcDef.Args {
				argExpr := exprNode.Args[i]
				argEvalValue, err := evalExpr(argExpr, env)
				if err != nil {
					return Value{}, err
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
		} else {
			// Try to look up variable instead
			if val, ok := env.Variables[exprNode.Identifier]; ok {
				if len(exprNode.Args) == 0 {
					return val, nil
				} else if val.Kind == ClosureType {
					return evalClosure(val.Closure, exprNode.Args, env, exprNode.Range)
				}
			}
		}

		// Check for closure application
		switch exprNode.Identifier {
		case "+":
			if len(exprNode.Args) != 2 {
				return Value{}, types.Error{Range: exprNode.Range,
					Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(exprNode.Args))}
			}
			return builtInBinaryOp(func(f1, f2 float64) float64 { return f1 + f2 }, exprNode.Args[0], exprNode.Args[1], env)
		case "-":
			if len(exprNode.Args) != 2 {
				return Value{}, types.Error{Range: exprNode.Range,
					Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(exprNode.Args))}
			}
			return builtInBinaryOp(func(f1, f2 float64) float64 { return f1 - f2 }, exprNode.Args[0], exprNode.Args[1], env)
		case "*":
			if len(exprNode.Args) != 2 {
				return Value{}, types.Error{Range: exprNode.Range,
					Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(exprNode.Args))}
			}
			return builtInBinaryOp(func(f1, f2 float64) float64 { return f1 * f2 }, exprNode.Args[0], exprNode.Args[1], env)
		case "/":
			if len(exprNode.Args) != 2 {
				return Value{}, types.Error{Range: exprNode.Range,
					Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(exprNode.Args))}
			}
			return builtInBinaryOp(func(f1, f2 float64) float64 { return f1 / f2 }, exprNode.Args[0], exprNode.Args[1], env)
		case "^":
			if len(exprNode.Args) != 2 {
				return Value{}, types.Error{Range: exprNode.Range,
					Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(exprNode.Args))}
			}
			return builtInBinaryOp(func(f1, f2 float64) float64 { return math.Pow(f1, f2) }, exprNode.Args[0], exprNode.Args[1], env)
		case "log":
			if len(exprNode.Args) != 2 {
				return Value{}, types.Error{Range: exprNode.Range,
					Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(exprNode.Args))}
			}
			return builtInBinaryOp(func(f1, f2 float64) float64 { return math.Log(f2) / math.Log(f1) }, exprNode.Args[0], exprNode.Args[1], env)
		case "sqrt":
			if len(exprNode.Args) != 1 {
				return Value{}, types.Error{Range: exprNode.Range,
					Simple: fmt.Sprintf("Unary function expected one paremters (got %d)", len(exprNode.Args))}
			}
			sqrtOf, err := evalExpr(exprNode.Args[0], env)
			if err != nil {
				return Value{}, err
			}
			if sqrtOf.Kind != NumType {
				return Value{}, types.Error{Range: exprNode.Args[0].GetRange(),
					Simple: fmt.Sprintf("Type error - expected number (got %s)", sqrtOf.Kind)}
			}
			val := Value{}
			val.NewNum(math.Sqrt(sqrtOf.Num))
			return val, nil
		case ">":
			if len(exprNode.Args) != 2 {
				return Value{}, types.Error{Range: exprNode.Range,
					Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(exprNode.Args))}
			}
			return builtInBinaryCompare(func(f1, f2 float64) bool { return f1 > f2 }, exprNode.Args[0], exprNode.Args[1], env)
		case ">=":
			if len(exprNode.Args) != 2 {
				return Value{}, types.Error{Range: exprNode.Range,
					Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(exprNode.Args))}
			}
			return builtInBinaryCompare(func(f1, f2 float64) bool { return f1 >= f2 }, exprNode.Args[0], exprNode.Args[1], env)
		case "<":
			if len(exprNode.Args) != 2 {
				return Value{}, types.Error{Range: exprNode.Range,
					Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(exprNode.Args))}
			}
			return builtInBinaryCompare(func(f1, f2 float64) bool { return f1 < f2 }, exprNode.Args[0], exprNode.Args[1], env)
		case "<=":
			if len(exprNode.Args) != 2 {
				return Value{}, types.Error{Range: exprNode.Range,
					Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(exprNode.Args))}
			}
			return builtInBinaryCompare(func(f1, f2 float64) bool { return f1 <= f2 }, exprNode.Args[0], exprNode.Args[1], env)
		case "=":
			if len(exprNode.Args) != 2 {
				return Value{}, types.Error{Range: exprNode.Range,
					Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(exprNode.Args))}
			}
			lhsVal, err := evalExpr(exprNode.Args[0], env)
			if err != nil {
				return Value{}, nil
			}
			rhsVal, err := evalExpr(exprNode.Args[1], env)
			if err != nil {
				return Value{}, nil
			}
			if lhsVal.Kind != rhsVal.Kind {
				return Value{}, types.Error{Range: exprNode.Range, Simple: "Operand types to = are different"}
			}
			val := Value{}
			val.NewBool(lhsVal.equals(rhsVal))
			return val, nil
		case "print":
			if len(exprNode.Args) != 1 {
				return Value{}, types.Error{Range: exprNode.Range,
					Simple: fmt.Sprintf("Unary function `print` expected one paremters (got %d)", len(exprNode.Args))}
			}
			val, err := evalExpr(exprNode.Args[0], env)
			if err != nil {
				return Value{}, err
			}
			fmt.Println(val.ToString())
			ret := Value{}
			ret.NewNull()
			return ret, nil
		case "length":
			if len(exprNode.Args) != 1 {
				return Value{}, types.Error{Range: exprNode.Range,
					Simple: fmt.Sprintf("Unary function `length` expected one paremters (got %d)", len(exprNode.Args))}
			}
			val, err := evalExpr(exprNode.Args[0], env)
			if err != nil {
				return Value{}, err
			}
			if val.Kind != ListType {
				return Value{}, types.Error{Range: exprNode.Range,
					Simple: fmt.Sprintf("Function length requires argument of type list (got %s)", val.Kind)}
			}
			lengthVal := Value{}
			lengthVal.NewNum(float64(len(val.List)))
			return lengthVal, nil
		case "insert":
			if len(exprNode.Args) != 3 {
				return Value{}, types.Error{Range: exprNode.Range,
					Simple: fmt.Sprintf("Function `insert` expected 3 paremters (got %d)", len(exprNode.Args))}
			}
			insertIndexVal, err := evalExpr(exprNode.Args[0], env)
			if err != nil {
				return Value{}, err
			}
			if insertIndexVal.Kind != NumType {
				return Value{}, types.Error{Range: exprNode.Range,
					Simple: fmt.Sprintf("Function `insert` expected first argument (index) to be a number (got %s)", insertIndexVal.Kind)}
			}
			valToInsert, err := evalExpr(exprNode.Args[1], env)
			if err != nil {
				return Value{}, err
			}
			list, err := evalExpr(exprNode.Args[2], env)
			if err != nil {
				return Value{}, err
			}
			if list.Kind != ListType {
				return Value{}, types.Error{Range: exprNode.Range,
					Simple: fmt.Sprintf("Function `insert` expected third argument (list) to be a list (got %s)", list.Kind)}
			}
			idx := int(insertIndexVal.Num)
			if idx < 0 {
				idx = 0
			}
			var newList []Value
			if idx >= len(list.List) {
				newList = append(list.List, valToInsert)
			} else {
				newList = append(list.List[:idx+1], list.List[idx:]...)
				newList[idx] = valToInsert
			}
			newListVal := Value{}
			newListVal.NewList(newList)
			return newListVal, nil
		case "nth":
			if len(exprNode.Args) != 2 {
				return Value{}, types.Error{Range: exprNode.Range,
					Simple: fmt.Sprintf("Function `nth` expected 2 paremters (got %d)", len(exprNode.Args))}
			}
			indexToGetVal, err := evalExpr(exprNode.Args[0], env)
			if err != nil {
				return Value{}, err
			}
			if indexToGetVal.Kind != NumType {
				return Value{}, types.Error{Range: exprNode.Range,
					Simple: fmt.Sprintf("Function `nth` expected first argument (index) to be a number (got %s)", indexToGetVal.Kind)}
			}
			list, err := evalExpr(exprNode.Args[1], env)
			if err != nil {
				return Value{}, err
			}
			if list.Kind != ListType {
				return Value{}, types.Error{Range: exprNode.Range,
					Simple: fmt.Sprintf("Function `nth` expected second argument (list) to be a list (got %s)", list.Kind)}
			}
			idx := int(indexToGetVal.Num)
			if idx < 0 || idx >= len(list.List) {
				v := Value{}
				v.NewNull()
				return v, nil
			}
			return list.List[idx], nil
		default:
			return Value{}, types.Error{Simple: fmt.Sprintf("Unknown function %s", exprNode.Identifier), Range: exprNode.GetRange()}
		}
	}
	return Value{}, errors.New("?")
}

func evalClosure(closureDef ClosureValue, args []ast.Expr, env Env, cRange types.FileRange) (Value, error) {
	// Closure method application
	if len(closureDef.Args) != len(args) {
		return Value{}, types.Error{Range: cRange,
			Simple: fmt.Sprintf("Expected %d arguments to closure application, got %d", len(closureDef.Args), len(args))}
	}
	// Construct closure environment, which is based on environment when closure was declared (the captured scope)
	closureEnv := closureDef.ClosureEnv

	for funcName, f := range env.Functions {
		if _, ok := closureEnv.Functions[funcName]; !ok {
			closureEnv.Functions[funcName] = f
		}
	}
	for varName, v := range env.Variables {
		if _, ok := closureEnv.Variables[varName]; !ok {
			closureEnv.Variables[varName] = v
		}
	}
	for i, argName := range closureDef.Args {
		argExpr := args[i]
		argEvalValue, err := evalExpr(argExpr, env)
		if err != nil {
			return Value{}, err
		}
		closureEnv.Variables[argName] = argEvalValue
	}
	for i, closureAst := range closureDef.Body {
		evalResult, err := Eval(closureAst, &closureEnv)
		if err != nil {
			return Value{}, err
		}
		if i == len(closureDef.Body)-1 {
			return evalResult, nil
		}
	}
	return Value{}, errors.New("??")
}

func (a Value) equals(b Value) bool {
	if a.Kind != b.Kind {
		return false
	}
	switch a.Kind {
	case NumType:
		return a.Num == b.Num
	case StringType:
		return a.String == b.String
	case BoolType:
		return a.Bool == b.Bool
	case NullType:
		return true
	case ListType:
		if len(a.List) != len(b.List) {
			return false
		}
		for i := range a.List {
			if !a.List[i].equals(b.List[i]) {
				return false
			}
		}
		return true
	}
	return false
}

func RunRepl() {
	reader := bufio.NewReader(os.Stdin)
	env := Env{}
	env.New()
	astConstruct := ast.AstConstructor{}
	// Otherwise really annoying
	astConstruct.AllowFunctionRedeclaration = true
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
		asts, err := astConstruct.CreateAst(expr)
		for _, ast := range asts.Asts {
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
}
