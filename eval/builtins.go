package eval

import (
	"fmt"
	"math"

	"github.com/benbanerjeerichards/lisp-calculator/ast"
	"github.com/benbanerjeerichards/lisp-calculator/types"
	"github.com/benbanerjeerichards/lisp-calculator/util"
)

func (evalulator Evalulator) builtInBinaryOp(f func(float64, float64) float64, lhs ast.Expr, rhs ast.Expr, env Env) (Value, error) {
	lhsValue, err := evalulator.evalExpr(lhs, env)
	if err != nil {
		return Value{}, err
	}
	rhsValue, err := evalulator.evalExpr(rhs, env)
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

func (evalulator Evalulator) builtInBinaryCompare(f func(float64, float64) bool, lhs ast.Expr, rhs ast.Expr, env Env) (Value, error) {
	lhsValue, err := evalulator.evalExpr(lhs, env)
	if err != nil {
		return Value{}, err
	}
	rhsValue, err := evalulator.evalExpr(rhs, env)
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

func (evalulator Evalulator) EvalBuiltin(funcAppNode ast.FunctionApplicationExpr, env Env) (Value, error) {
	switch funcAppNode.Identifier {
	case "+":
		if len(funcAppNode.Args) != 2 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(funcAppNode.Args))}
		}
		return evalulator.builtInBinaryOp(func(f1, f2 float64) float64 { return f1 + f2 }, funcAppNode.Args[0], funcAppNode.Args[1], env)
	case "-":
		if len(funcAppNode.Args) != 2 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(funcAppNode.Args))}
		}
		return evalulator.builtInBinaryOp(func(f1, f2 float64) float64 { return f1 - f2 }, funcAppNode.Args[0], funcAppNode.Args[1], env)
	case "*":
		if len(funcAppNode.Args) != 2 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(funcAppNode.Args))}
		}
		return evalulator.builtInBinaryOp(func(f1, f2 float64) float64 { return f1 * f2 }, funcAppNode.Args[0], funcAppNode.Args[1], env)
	case "/":
		if len(funcAppNode.Args) != 2 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(funcAppNode.Args))}
		}
		return evalulator.builtInBinaryOp(func(f1, f2 float64) float64 { return f1 / f2 }, funcAppNode.Args[0], funcAppNode.Args[1], env)
	case "^":
		if len(funcAppNode.Args) != 2 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(funcAppNode.Args))}
		}
		return evalulator.builtInBinaryOp(func(f1, f2 float64) float64 { return math.Pow(f1, f2) }, funcAppNode.Args[0], funcAppNode.Args[1], env)
	case "log":
		if len(funcAppNode.Args) != 2 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(funcAppNode.Args))}
		}
		return evalulator.builtInBinaryOp(func(f1, f2 float64) float64 { return math.Log(f2) / math.Log(f1) }, funcAppNode.Args[0], funcAppNode.Args[1], env)
	case "sqrt":
		if len(funcAppNode.Args) != 1 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Unary function expected one paremters (got %d)", len(funcAppNode.Args))}
		}
		sqrtOf, err := evalulator.evalExpr(funcAppNode.Args[0], env)
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
		return evalulator.builtInBinaryCompare(func(f1, f2 float64) bool { return f1 > f2 }, funcAppNode.Args[0], funcAppNode.Args[1], env)
	case ">=":
		if len(funcAppNode.Args) != 2 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(funcAppNode.Args))}
		}
		return evalulator.builtInBinaryCompare(func(f1, f2 float64) bool { return f1 >= f2 }, funcAppNode.Args[0], funcAppNode.Args[1], env)
	case "<":
		if len(funcAppNode.Args) != 2 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(funcAppNode.Args))}
		}
		return evalulator.builtInBinaryCompare(func(f1, f2 float64) bool { return f1 < f2 }, funcAppNode.Args[0], funcAppNode.Args[1], env)
	case "<=":
		if len(funcAppNode.Args) != 2 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(funcAppNode.Args))}
		}
		return evalulator.builtInBinaryCompare(func(f1, f2 float64) bool { return f1 <= f2 }, funcAppNode.Args[0], funcAppNode.Args[1], env)
	case "=":
		if len(funcAppNode.Args) != 2 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Binary function expected two paremters (got %d)", len(funcAppNode.Args))}
		}
		lhsVal, err := evalulator.evalExpr(funcAppNode.Args[0], env)
		if err != nil {
			return Value{}, nil
		}
		rhsVal, err := evalulator.evalExpr(funcAppNode.Args[1], env)
		if err != nil {
			return Value{}, nil
		}
		val := Value{}
		val.NewBool(lhsVal.equals(rhsVal))
		return val, nil
	case "not":
		if len(funcAppNode.Args) != 1 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Binary function `not` expected one parameter (got %d)", len(funcAppNode.Args))}
		}
		val, err := evalulator.evalExpr(funcAppNode.Args[0], env)
		if err != nil {
			return Value{}, err
		}
		if val.Kind != BoolType {
			return Value{}, types.Error{Range: funcAppNode.Range, Simple: fmt.Sprintf("Function not can only be applied to type Bool (got %s)", val.Kind)}
		}
		val.NewBool(!val.Bool)
		return val, nil
	case "and":
		if len(funcAppNode.Args) != 2 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Binary function `and` expected two parameters (got %d)", len(funcAppNode.Args))}
		}
		valLhs, err := evalulator.evalExpr(funcAppNode.Args[0], env)
		if err != nil {
			return Value{}, err
		}
		valRhs, err := evalulator.evalExpr(funcAppNode.Args[1], env)
		if err != nil {
			return Value{}, err
		}
		if valLhs.Kind != BoolType {
			return Value{}, types.Error{Range: funcAppNode.Range, Simple: fmt.Sprintf("Function `and` LHS can only be applied to type Bool (got %s)", valLhs.Kind)}
		}
		if valRhs.Kind != BoolType {
			return Value{}, types.Error{Range: funcAppNode.Range, Simple: fmt.Sprintf("Function `and` RHS can only be applied to type Bool (got %s)", valRhs.Kind)}
		}
		val := Value{}
		val.NewBool(valRhs.Bool && valLhs.Bool)
		return val, nil
	case "or":
		if len(funcAppNode.Args) != 2 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Binary function `or` expected two parameters (got %d)", len(funcAppNode.Args))}
		}
		valLhs, err := evalulator.evalExpr(funcAppNode.Args[0], env)
		if err != nil {
			return Value{}, err
		}
		valRhs, err := evalulator.evalExpr(funcAppNode.Args[1], env)
		if err != nil {
			return Value{}, err
		}
		if valLhs.Kind != BoolType {
			return Value{}, types.Error{Range: funcAppNode.Range, Simple: fmt.Sprintf("Function `or` LHS can only be applied to type Bool (got %s)", valLhs.Kind)}
		}
		if valRhs.Kind != BoolType {
			return Value{}, types.Error{Range: funcAppNode.Range, Simple: fmt.Sprintf("Function `or` RHS can only be applied to type Bool (got %s)", valRhs.Kind)}
		}
		val := Value{}
		val.NewBool(valRhs.Bool || valLhs.Bool)
		return val, nil
	case "concat":
		if len(funcAppNode.Args) != 2 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Binary function `concat` expected two parameters (got %d)", len(funcAppNode.Args))}
		}
		valLhs, err := evalulator.evalExpr(funcAppNode.Args[0], env)
		if err != nil {
			return Value{}, err
		}
		valRhs, err := evalulator.evalExpr(funcAppNode.Args[1], env)
		if err != nil {
			return Value{}, err
		}
		val := Value{}
		lString := valLhs.ToString()
		if valLhs.Kind == StringType {
			lString = valLhs.String
		}
		rString := valRhs.ToString()
		if valRhs.Kind == StringType {
			rString = valRhs.String
		}

		val.NewString(fmt.Sprintf("%s%s", lString, rString))
		return val, nil

	case "panic":
		if len(funcAppNode.Args) != 1 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Unary function `panic` expected one paremters (got %d)", len(funcAppNode.Args))}
		}
		val, err := evalulator.evalExpr(funcAppNode.Args[0], env)
		if err != nil {
			return Value{}, err
		}
		if val.Kind != StringType {
			return Value{}, types.Error{Range: funcAppNode.Range, Simple: fmt.Sprintf("Function not can only be applied to type String (got %s)", val.Kind)}
		}
		return Value{}, types.Error{Range: funcAppNode.Range,
			Simple: fmt.Sprintf("%v panic - %s", funcAppNode.Range, val.String)}
	case "print":
		if len(funcAppNode.Args) != 1 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Unary function `print` expected one paremters (got %d)", len(funcAppNode.Args))}
		}
		val, err := evalulator.evalExpr(funcAppNode.Args[0], env)
		if err != nil {
			return Value{}, err
		}
		str := val.ToString()
		if val.Kind == StringType {
			str = val.String
		}
		fmt.Print(str)
		ret := Value{}
		ret.NewNull()
		return ret, nil
	case "length":
		if len(funcAppNode.Args) != 1 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Unary function `length` expected one paremters (got %d)", len(funcAppNode.Args))}
		}
		val, err := evalulator.evalExpr(funcAppNode.Args[0], env)
		if err != nil {
			return Value{}, err
		}
		if val.Kind != ListType && val.Kind != StringType {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Function length requires argument of type list or string(got %s)", val.Kind)}
		}
		lengthVal := Value{}
		if val.Kind == ListType {
			lengthVal.NewNum(float64(len(val.List)))
		} else {
			lengthVal.NewNum(float64(len(val.String)))
		}
		return lengthVal, nil
	case "chr":
		if len(funcAppNode.Args) != 1 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Unary function `chr` expected one paremters (got %d)", len(funcAppNode.Args))}
		}
		codeVal, err := evalulator.evalExpr(funcAppNode.Args[0], env)
		if err != nil {
			return Value{}, err
		}
		if codeVal.Kind != NumType {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Function chr requires argument of type number (got %s)", codeVal.Kind)}
		}

		val := Value{}
		val.NewString(string(int(codeVal.Num)))
		return val, nil
	case "ord":
		if len(funcAppNode.Args) != 1 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Unary function `ord` expected one paremters (got %d)", len(funcAppNode.Args))}
		}
		codeVal, err := evalulator.evalExpr(funcAppNode.Args[0], env)
		if err != nil {
			return Value{}, err
		}
		if codeVal.Kind != StringType {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Function ord requires argument of type string (got %s)", codeVal.Kind)}
		}

		val := Value{}
		val.NewNum(float64(int(codeVal.String[0])))
		return val, nil
	case "readFile":
		if len(funcAppNode.Args) != 1 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Unary function `readFile` expected one paremters (got %d)", len(funcAppNode.Args))}
		}
		pathVal, err := evalulator.evalExpr(funcAppNode.Args[0], env)
		if err != nil {
			return Value{}, err
		}
		if pathVal.Kind != StringType {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Function readFile requires argument of type string (got %s)", pathVal.Kind)}
		}
		contents, err := util.ReadFile(pathVal.String)
		if err != nil {
			return Value{}, types.Error{Range: funcAppNode.Args[0].GetRange(), Simple: fmt.Sprintf("Failed to read from file %s", pathVal.String)}
		}
		val := Value{}
		val.NewString(contents)
		return val, nil

	case "insert":
		if len(funcAppNode.Args) != 3 {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Function `insert` expected 3 paremters (got %d)", len(funcAppNode.Args))}
		}
		insertIndexVal, err := evalulator.evalExpr(funcAppNode.Args[0], env)
		if err != nil {
			return Value{}, err
		}
		if insertIndexVal.Kind != NumType {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Function `insert` expected first argument (index) to be a number (got %s)", insertIndexVal.Kind)}
		}
		valToInsert, err := evalulator.evalExpr(funcAppNode.Args[1], env)
		if err != nil {
			return Value{}, err
		}
		list, err := evalulator.evalExpr(funcAppNode.Args[2], env)
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
		indexToGetVal, err := evalulator.evalExpr(funcAppNode.Args[0], env)
		if err != nil {
			return Value{}, err
		}
		if indexToGetVal.Kind != NumType {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Function `nth` expected first argument (index) to be a number (got %s)", indexToGetVal.Kind)}
		}
		subject, err := evalulator.evalExpr(funcAppNode.Args[1], env)
		if err != nil {
			return Value{}, err
		}
		if subject.Kind != ListType && subject.Kind != StringType {
			return Value{}, types.Error{Range: funcAppNode.Range,
				Simple: fmt.Sprintf("Function `nth` expected second argument (list) to be a list (got %s)", subject.Kind)}
		}
		idx := int(indexToGetVal.Num)
		if idx < 0 || (subject.Kind == ListType && idx >= len(subject.List)) || (subject.Kind == StringType && idx >= len(subject.String)) {
			v := Value{}
			v.NewNull()
			return v, nil
		}
		if subject.Kind == StringType {
			c := subject.String[idx]
			val := Value{}
			val.NewString(string(c))
			return val, nil
		}
		return subject.List[idx], nil
	default:
		return Value{}, types.Error{Simple: fmt.Sprintf("Unknown function %s", funcAppNode.Identifier), Range: funcAppNode.GetRange()}
	}
}
