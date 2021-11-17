package eval

import (
	"fmt"
	"math"

	"github.com/benbanerjeerichards/lisp-calculator/ast"
	"github.com/benbanerjeerichards/lisp-calculator/types"
)

func builtInBinaryOp(f func(float64, float64) float64, lhs ast.Expr, rhs ast.Expr, env Env, functions map[string]*ast.FuncDefStmt) (Value, error) {
	lhsValue, err := evalExpr(lhs, env, functions)
	if err != nil {
		return Value{}, err
	}
	rhsValue, err := evalExpr(rhs, env, functions)
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

func builtInBinaryCompare(f func(float64, float64) bool, lhs ast.Expr, rhs ast.Expr, env Env, functions map[string]*ast.FuncDefStmt) (Value, error) {
	lhsValue, err := evalExpr(lhs, env, functions)
	if err != nil {
		return Value{}, err
	}
	rhsValue, err := evalExpr(rhs, env, functions)
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

func EvalBuiltin(funcAppNode ast.FunctionApplicationExpr, env Env, functions map[string]*ast.FuncDefStmt) (Value, error) {
	switch funcAppNode.Identifier {
	case "+":
		if len(funcAppNode.Args) != 2 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(funcAppNode.Args))}
		}
		return builtInBinaryOp(func(f1, f2 float64) float64 { return f1 + f2 }, funcAppNode.Args[0], funcAppNode.Args[1], env, functions)
	case "-":
		if len(funcAppNode.Args) != 2 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(funcAppNode.Args))}
		}
		return builtInBinaryOp(func(f1, f2 float64) float64 { return f1 - f2 }, funcAppNode.Args[0], funcAppNode.Args[1], env, functions)
	case "*":
		if len(funcAppNode.Args) != 2 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(funcAppNode.Args))}
		}
		return builtInBinaryOp(func(f1, f2 float64) float64 { return f1 * f2 }, funcAppNode.Args[0], funcAppNode.Args[1], env, functions)
	case "/":
		if len(funcAppNode.Args) != 2 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(funcAppNode.Args))}
		}
		return builtInBinaryOp(func(f1, f2 float64) float64 { return f1 / f2 }, funcAppNode.Args[0], funcAppNode.Args[1], env, functions)
	case "^":
		if len(funcAppNode.Args) != 2 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(funcAppNode.Args))}
		}
		return builtInBinaryOp(func(f1, f2 float64) float64 { return math.Pow(f1, f2) }, funcAppNode.Args[0], funcAppNode.Args[1], env, functions)
	case "log":
		if len(funcAppNode.Args) != 2 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(funcAppNode.Args))}
		}
		return builtInBinaryOp(func(f1, f2 float64) float64 { return math.Log(f2) / math.Log(f1) }, funcAppNode.Args[0], funcAppNode.Args[1], env, functions)
	case "sqrt":
		if len(funcAppNode.Args) != 1 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Unary function expected one paremters (got %d)", len(funcAppNode.Args))}
		}
		sqrtOf, err := evalExpr(funcAppNode.Args[0], env, functions)
		if err != nil {
			return Value{}, err
		}
		if sqrtOf.Kind != NumType {
			return Value{}, types.Error{Range: funcAppNode.Args[0].GetRange(),
				Simple: fmt.Sprintf("Type error - expected number (got %s)", sqrtOf.Kind)}
		}
		val := Value{}
		val.NewNum(math.Sqrt(sqrtOf.Num))
		return val, nil
	case ">":
		if len(funcAppNode.Args) != 2 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(funcAppNode.Args))}
		}
		return builtInBinaryCompare(func(f1, f2 float64) bool { return f1 > f2 }, funcAppNode.Args[0], funcAppNode.Args[1], env, functions)
	case ">=":
		if len(funcAppNode.Args) != 2 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(funcAppNode.Args))}
		}
		return builtInBinaryCompare(func(f1, f2 float64) bool { return f1 >= f2 }, funcAppNode.Args[0], funcAppNode.Args[1], env, functions)
	case "<":
		if len(funcAppNode.Args) != 2 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(funcAppNode.Args))}
		}
		return builtInBinaryCompare(func(f1, f2 float64) bool { return f1 < f2 }, funcAppNode.Args[0], funcAppNode.Args[1], env, functions)
	case "<=":
		if len(funcAppNode.Args) != 2 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(funcAppNode.Args))}
		}
		return builtInBinaryCompare(func(f1, f2 float64) bool { return f1 <= f2 }, funcAppNode.Args[0], funcAppNode.Args[1], env, functions)
	case "=":
		if len(funcAppNode.Args) != 2 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(funcAppNode.Args))}
		}
		lhsVal, err := evalExpr(funcAppNode.Args[0], env, functions)
		if err != nil {
			return Value{}, nil
		}
		rhsVal, err := evalExpr(funcAppNode.Args[1], env, functions)
		if err != nil {
			return Value{}, nil
		}
		if lhsVal.Kind != rhsVal.Kind {
			return Value{}, types.Error{Range: funcAppNode.Range, Simple: "Operand types to = are different"}
		}
		val := Value{}
		val.NewBool(lhsVal.equals(rhsVal))
		return val, nil
	case "print":
		if len(funcAppNode.Args) != 1 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Unary function `print` expected one paremters (got %d)", len(funcAppNode.Args))}
		}
		val, err := evalExpr(funcAppNode.Args[0], env, functions)
		if err != nil {
			return Value{}, err
		}
		fmt.Println(val.ToString())
		ret := Value{}
		ret.NewNull()
		return ret, nil
	case "length":
		if len(funcAppNode.Args) != 1 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Unary function `length` expected one paremters (got %d)", len(funcAppNode.Args))}
		}
		val, err := evalExpr(funcAppNode.Args[0], env, functions)
		if err != nil {
			return Value{}, err
		}
		if val.Kind != ListType {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Function length requires argument of type list (got %s)", val.Kind)}
		}
		lengthVal := Value{}
		lengthVal.NewNum(float64(len(val.List)))
		return lengthVal, nil
	case "insert":
		if len(funcAppNode.Args) != 3 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Function `insert` expected 3 paremters (got %d)", len(funcAppNode.Args))}
		}
		insertIndexVal, err := evalExpr(funcAppNode.Args[0], env, functions)
		if err != nil {
			return Value{}, err
		}
		if insertIndexVal.Kind != NumType {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Function `insert` expected first argument (index) to be a number (got %s)", insertIndexVal.Kind)}
		}
		valToInsert, err := evalExpr(funcAppNode.Args[1], env, functions)
		if err != nil {
			return Value{}, err
		}
		list, err := evalExpr(funcAppNode.Args[2], env, functions)
		if err != nil {
			return Value{}, err
		}
		if list.Kind != ListType {
			return Value{}, types.Error{Range: funcAppNode.Range,
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
		if len(funcAppNode.Args) != 2 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Function `nth` expected 2 paremters (got %d)", len(funcAppNode.Args))}
		}
		indexToGetVal, err := evalExpr(funcAppNode.Args[0], env, functions)
		if err != nil {
			return Value{}, err
		}
		if indexToGetVal.Kind != NumType {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Function `nth` expected first argument (index) to be a number (got %s)", indexToGetVal.Kind)}
		}
		list, err := evalExpr(funcAppNode.Args[1], env, functions)
		if err != nil {
			return Value{}, err
		}
		if list.Kind != ListType {
			return Value{}, types.Error{Range: funcAppNode.Range,
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
		return Value{}, types.Error{Simple: fmt.Sprintf("Unknown function %s", funcAppNode.Identifier), Range: funcAppNode.GetRange()}
	}
}
