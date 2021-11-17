package eval

import (
	"errors"
	"fmt"
	"strings"

	"github.com/benbanerjeerichards/lisp-calculator/ast"
	"github.com/benbanerjeerichards/lisp-calculator/types"
)

type Env struct {
	Variables map[string]Value
	Functions map[string]ast.FuncDefStmt
}

func (env *Env) New() {
	env.Variables = make(map[string]Value)
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
	Value     Value
	Variables map[string]*ast.VarDefStmt
}

// Evalulate a program
// If a main function exists, it executes it and returns the result of the main function
// Otherwise, it evaluates every ast in turn and returns the value of the final expression/statement
func EvalProgram(astResult ast.AstResult, programArgs []string) (Value, error) {
	if len(astResult.Asts) == 0 {
		return Value{}, errors.New("eval program requires non-zero numbers of asts")
	}

	// Search for main function
	if mainFunc, ok := astResult.Functions["main"]; ok {
		if len(mainFunc.Args) >= 2 {
			return Value{}, types.Error{Range: mainFunc.Range, Simple: "Main function must take either zero arguments or one argument (for args)"}
		}

		// To call main, all we do is construct a fake main() call (with or without arguments, depending on definition)
		argList := []ast.Expr{}
		if len(mainFunc.Args) == 1 {
			argsExprs := []ast.Expr{}
			for _, arg := range programArgs {
				argsExprs = append(argsExprs, ast.StringExpr{Value: arg})
			}
			argList = append(argList, ast.ListExpr{Value: argsExprs})
		}

		funcAppExpr := ast.FunctionApplicationExpr{
			Range: types.FileRange{Start: types.FilePos{Line: 0, Col: 0, Position: 0}, End: types.FilePos{Line: 0, Col: 0, Position: 0}},
			Args:  argList,
		}
		env := Env{}
		env.New()
		return evalFunctionApplication(funcAppExpr, *mainFunc, env, astResult.Functions)
	}
	env := Env{}
	env.New()

	for i, ast := range astResult.Asts {
		result, err := Eval(ast, &env, astResult.Functions)
		if err != nil {
			return Value{}, err
		}
		if i == len(astResult.Asts)-1 {
			return result, nil
		}
	}
	return Value{}, errors.New("unknown error?")
}

func Eval(astNode ast.Ast, env *Env, functions map[string]*ast.FuncDefStmt) (Value, error) {
	if astNode.Kind == ast.StmtType {
		err := evalStmt(astNode.Statement, env, functions)
		if err != nil {
			return Value{}, err
		}

		return Value{Kind: NullType}, nil
	}
	if astNode.Kind == ast.ExprType {
		val, err := evalExpr(astNode.Expression, *env, functions)
		if err != nil {
			return Value{}, err
		}
		return val, nil
	}
	return Value{}, errors.New("?")
}

func evalStmt(node ast.Stmt, env *Env, functions map[string]*ast.FuncDefStmt) error {
	switch stmtNode := node.(type) {
	case ast.VarDefStmt:
		result, err := evalExpr(stmtNode.Value, *env, functions)
		if err != nil {
			return err
		}
		env.Variables[stmtNode.Identifier] = result
	case ast.FuncDefStmt:
		// NOP - already handled by AST
	case ast.WhileStmt:
		cond, err := evalExpr(stmtNode.Condition, *env, functions)
		if err != nil {
			return err
		}
		if cond.Kind != BoolType {
			return types.Error{Range: stmtNode.Condition.GetRange(),
				Simple: fmt.Sprintf("Type Error - while loop condition is not a boolean (got %s)", cond.Kind)}
		}
		for cond.Bool {
			for _, ast := range stmtNode.Body {
				Eval(ast, env, functions)
			}
			cond, err = evalExpr(stmtNode.Condition, *env, functions)
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

func evalExpr(node ast.Expr, env Env, functions map[string]*ast.FuncDefStmt) (Value, error) {
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
			itemValue, err := evalExpr(expr, env, functions)
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
		condRes, err := evalExpr(exprNode.Condition, env, functions)
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
			evalResult, err := Eval(ast, &env, functions)
			if err != nil {
				return Value{}, err
			}
			if i == len(branch)-1 {
				return evalResult, nil
			}
		}
	case ast.IfOnlyExpr:
		condRes, err := evalExpr(exprNode.Condition, env, functions)
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
			evalResult, err := Eval(ast, &env, functions)
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
		for varName, v := range env.Variables {
			closureEnv.Variables[varName] = v
		}
		value.NewClosure(exprNode.Args, exprNode.Body, closureEnv)
		return value, nil
	case ast.ClosureApplicationExpr:
		closureVal, err := evalExpr(exprNode.Closure, env, functions)
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
		return evalClosure(closure, exprNode.Args, env, functions, exprNode.Range)
	case ast.FunctionApplicationExpr:
		// First look up in function defintions, then try builtins
		if funcDef, ok := functions[exprNode.Identifier]; ok {
			return evalFunctionApplication(exprNode, *funcDef, env, functions)
		} else {
			// Try to look up variable instead
			if val, ok := env.Variables[exprNode.Identifier]; ok {
				if len(exprNode.Args) == 0 {
					return val, nil
				} else if val.Kind == ClosureType {
					return evalClosure(val.Closure, exprNode.Args, env, functions, exprNode.Range)
				}
			}
			return EvalBuiltin(exprNode, env, functions)
		}
	}
	return Value{}, errors.New("?")
}

func evalFunctionApplication(funcAppNode ast.FunctionApplicationExpr, funcDef ast.FuncDefStmt, env Env, functions map[string]*ast.FuncDefStmt) (Value, error) {
	if len(funcDef.Args) != len(funcAppNode.Args) {
		return Value{}, types.Error{Range: funcAppNode.Range,
			Simple: fmt.Sprintf("Bad funtion application - expected %d arguments but recieved %d", len(funcDef.Args), len(funcAppNode.Args))}
	}
	funcAppEnv := Env{}
	funcAppEnv.New()
	for i, argName := range funcDef.Args {
		argExpr := funcAppNode.Args[i]
		argEvalValue, err := evalExpr(argExpr, env, functions)
		if err != nil {
			return Value{}, err
		}

		funcAppEnv.Variables[argName] = argEvalValue
	}
	for i, funcDefAst := range funcDef.Body {
		evalResult, err := Eval(funcDefAst, &funcAppEnv, functions)
		if err != nil {
			return Value{}, err
		}
		if i == len(funcDef.Body)-1 {
			return evalResult, nil
		}
	}
	return Value{}, errors.New("?????")
}

func evalClosure(closureDef ClosureValue, args []ast.Expr, env Env, functions map[string]*ast.FuncDefStmt, cRange types.FileRange) (Value, error) {
	// Closure method application
	if len(closureDef.Args) != len(args) {
		return Value{}, types.Error{Range: cRange,
			Simple: fmt.Sprintf("Expected %d arguments to closure application, got %d", len(closureDef.Args), len(args))}
	}
	// Construct closure environment, which is based on environment when closure was declared (the captured scope)
	closureEnv := closureDef.ClosureEnv

	for varName, v := range env.Variables {
		if _, ok := closureEnv.Variables[varName]; !ok {
			closureEnv.Variables[varName] = v
		}
	}
	for i, argName := range closureDef.Args {
		argExpr := args[i]
		argEvalValue, err := evalExpr(argExpr, env, functions)
		if err != nil {
			return Value{}, err
		}
		closureEnv.Variables[argName] = argEvalValue
	}
	for i, closureAst := range closureDef.Body {
		evalResult, err := Eval(closureAst, &closureEnv, functions)
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
