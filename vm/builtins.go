package vm

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"

	"github.com/benbanerjeerichards/lisp-calculator/types"
	"github.com/benbanerjeerichards/lisp-calculator/util"
)

type Builtin struct {
	Function   func([]Value) (Value, error)
	NumArgs    int
	Identifier string
}

func checKTypes(values []Value, expected []string) error {
	for i, val := range values {
		if val.Kind != expected[i] {
			return types.Error{Simple: fmt.Sprintf("Type error for argument %d - expected %s but got %s", i+1, expected[i], val.Kind)}
		}
	}
	return nil
}

var Builtins []Builtin = []Builtin{
	{
		Identifier: "+",
		NumArgs:    2,
		Function: func(v []Value) (Value, error) {
			err := checKTypes(v, []string{NumType, NumType})
			if err != nil {
				return Value{}, err
			}
			res := Value{}
			res.NewNum(v[0].Num + v[1].Num)
			return res, nil
		},
	},
	{
		Identifier: "-",
		NumArgs:    2,
		Function: func(v []Value) (Value, error) {
			err := checKTypes(v, []string{NumType, NumType})
			if err != nil {
				return Value{}, err
			}
			res := Value{}
			res.NewNum(v[0].Num - v[1].Num)
			return res, nil
		},
	},
	{
		Identifier: "/",
		NumArgs:    2,
		Function: func(v []Value) (Value, error) {
			err := checKTypes(v, []string{NumType, NumType})
			if err != nil {
				return Value{}, err
			}
			res := Value{}
			res.NewNum(v[0].Num / v[1].Num)
			return res, nil
		},
	},
	{
		Identifier: "*",
		NumArgs:    2,
		Function: func(v []Value) (Value, error) {
			err := checKTypes(v, []string{NumType, NumType})
			if err != nil {
				return Value{}, err
			}
			res := Value{}
			res.NewNum(v[0].Num * v[1].Num)
			return res, nil
		},
	},
	{
		Identifier: "^",
		NumArgs:    2,
		Function: func(v []Value) (Value, error) {
			err := checKTypes(v, []string{NumType, NumType})
			if err != nil {
				return Value{}, err
			}
			res := Value{}
			res.NewNum(math.Pow(v[0].Num, v[1].Num))
			return res, nil
		},
	},
	{
		Identifier: "mod",
		NumArgs:    2,
		Function: func(v []Value) (Value, error) {
			err := checKTypes(v, []string{NumType, NumType})
			if err != nil {
				return Value{}, err
			}
			res := Value{}
			res.NewNum(math.Mod(v[0].Num, v[1].Num))
			return res, nil
		},
	},
	{
		Identifier: "log",
		NumArgs:    2,
		Function: func(v []Value) (Value, error) {
			err := checKTypes(v, []string{NumType, NumType})
			if err != nil {
				return Value{}, err
			}
			res := Value{}
			res.NewNum(math.Log(v[1].Num) / math.Log(v[0].Num))
			return res, nil
		},
	},
	{
		Identifier: "sqrt",
		NumArgs:    1,
		Function: func(v []Value) (Value, error) {
			err := checKTypes(v, []string{NumType})
			if err != nil {
				return Value{}, err
			}
			res := Value{}
			res.NewNum(math.Sqrt(v[0].Num))
			return res, nil
		},
	},
	{
		Identifier: "rng",
		NumArgs:    0,
		Function: func(v []Value) (Value, error) {
			// FIXME
			res := Value{}
			res.NewNum(rand.Float64())
			return res, nil
		},
	},
	{
		Identifier: "floor",
		NumArgs:    1,
		Function: func(v []Value) (Value, error) {
			err := checKTypes(v, []string{NumType})
			if err != nil {
				return Value{}, err
			}
			res := Value{}
			res.NewNum(math.Floor(v[0].Num))
			return res, nil
		},
	},
	{
		Identifier: "ceil",
		NumArgs:    1,
		Function: func(v []Value) (Value, error) {
			err := checKTypes(v, []string{NumType})
			if err != nil {
				return Value{}, err
			}
			res := Value{}
			res.NewNum(math.Ceil(v[0].Num))
			return res, nil
		},
	},
	{
		Identifier: ">",
		NumArgs:    2,
		Function: func(v []Value) (Value, error) {
			err := checKTypes(v, []string{NumType, NumType})
			if err != nil {
				return Value{}, err
			}
			res := Value{}
			res.NewBool(v[0].Num > v[1].Num)
			return res, nil
		},
	},
	{
		Identifier: ">=",
		NumArgs:    2,
		Function: func(v []Value) (Value, error) {
			err := checKTypes(v, []string{NumType, NumType})
			if err != nil {
				return Value{}, err
			}
			res := Value{}
			res.NewBool(v[0].Num >= v[1].Num)
			return res, nil
		},
	},
	{
		Identifier: "<",
		NumArgs:    2,
		Function: func(v []Value) (Value, error) {
			err := checKTypes(v, []string{NumType, NumType})
			if err != nil {
				return Value{}, err
			}
			res := Value{}
			res.NewBool(v[0].Num < v[1].Num)
			return res, nil
		},
	},
	{
		Identifier: "<=",
		NumArgs:    2,
		Function: func(v []Value) (Value, error) {
			err := checKTypes(v, []string{NumType, NumType})
			if err != nil {
				return Value{}, err
			}
			res := Value{}
			res.NewBool(v[0].Num <= v[1].Num)
			return res, nil
		},
	},
	{
		Identifier: "=",
		NumArgs:    2,
		Function: func(v []Value) (Value, error) {
			res := Value{}
			res.NewBool(v[0].equals(v[1]))
			return res, nil
		},
	},
	{
		Identifier: "not",
		NumArgs:    1,
		Function: func(v []Value) (Value, error) {
			err := checKTypes(v, []string{BoolType})
			if err != nil {
				return Value{}, err
			}
			res := Value{}
			res.NewBool(!v[0].Bool)
			return res, nil
		},
	},
	{
		Identifier: "and",
		NumArgs:    2,
		Function: func(v []Value) (Value, error) {
			err := checKTypes(v, []string{BoolType, BoolType})
			if err != nil {
				return Value{}, err
			}
			res := Value{}
			res.NewBool(v[0].Bool && v[1].Bool)
			return res, nil
		},
	},
	{
		Identifier: "or",
		NumArgs:    2,
		Function: func(v []Value) (Value, error) {
			err := checKTypes(v, []string{BoolType, BoolType})
			if err != nil {
				return Value{}, err
			}
			res := Value{}
			res.NewBool(v[0].Bool || v[1].Bool)
			return res, nil
		},
	},
	{
		Identifier: "concat",
		NumArgs:    2,
		Function: func(v []Value) (Value, error) {
			lString := v[0].ToString()
			if v[0].Kind == StringType {
				lString = v[0].String
			}
			rString := v[1].ToString()
			if v[1].Kind == StringType {
				rString = v[1].String
			}
			val := Value{}
			val.NewString(fmt.Sprintf("%s%s", lString, rString))
			return val, nil
		},
	},
	{
		Identifier: "panic",
		NumArgs:    1,
		Function: func(v []Value) (Value, error) {
			err := checKTypes(v, []string{StringType})
			if err != nil {
				return Value{}, err
			}
			return Value{}, types.Error{Simple: fmt.Sprintf("panic - %s", v[0].String)}
		},
	},
	{
		Identifier: "print",
		NumArgs:    1,
		Function: func(v []Value) (Value, error) {
			if v[0].Kind == StringType {
				fmt.Print(v[0].String)
			} else {
				fmt.Print(v[0].ToString())
			}
			val := Value{}
			val.NewNull()
			return val, nil
		},
	},
	{
		Identifier: "length",
		NumArgs:    1,
		Function: func(v []Value) (Value, error) {
			val := v[0]
			if val.Kind != ListType && val.Kind != StringType {
				return Value{}, types.Error{
					Simple: fmt.Sprintf("Function length requires argument of type list or string(got %s)", val.Kind)}
			}
			lengthVal := Value{}
			if val.Kind == ListType {
				lengthVal.NewNum(float64(len(val.List)))
			} else {
				lengthVal.NewNum(float64(len(val.String)))
			}
			return lengthVal, nil
		},
	},
	{
		Identifier: "chr",
		NumArgs:    1,
		Function: func(v []Value) (Value, error) {
			err := checKTypes(v, []string{NumType})
			if err != nil {
				return Value{}, err
			}
			val := Value{}
			val.NewString(string(int(v[0].Num)))
			return val, nil
		},
	},
	{
		Identifier: "ord",
		NumArgs:    1,
		Function: func(v []Value) (Value, error) {
			err := checKTypes(v, []string{StringType})
			if err != nil {
				return Value{}, err
			}
			if len(v[0].String) != 1 {
				return Value{}, types.Error{Simple: "ord expected string of length 1"}
			}
			val := Value{}
			val.NewNum(float64(int(v[0].String[0])))
			return val, nil
		},
	},
	{
		Identifier: "readFile",
		NumArgs:    1,
		Function: func(v []Value) (Value, error) {
			err := checKTypes(v, []string{StringType})
			if err != nil {
				return Value{}, err
			}
			contents, err := util.ReadFile(v[0].String)
			if err != nil {
				return Value{}, types.Error{Simple: fmt.Sprintf("Failed to read from file %s", v[0].String)}
			}
			val := Value{}
			val.NewString(contents)
			return val, nil
		},
	},
	{
		Identifier: "input",
		NumArgs:    0,
		Function: func(v []Value) (Value, error) {
			// TODO make this work with go-prompt for REPL
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			val := Value{}
			val.NewString(text)
			return val, nil
		},
	},
	{
		Identifier: "insert",
		NumArgs:    3,
		Function: func(v []Value) (Value, error) {
			if v[0].Kind != NumType {
				return Value{}, types.Error{Simple: fmt.Sprintf("Type error - argument 1 of insert expected Num, got %s", v[0].Kind)}
			}
			if v[2].Kind != ListType {
				return Value{}, types.Error{Simple: fmt.Sprintf("Type error - argument 3 of insert expected type List , got %s", v[0].Kind)}
			}
			idx := int(v[0].Num)
			if idx < 0 {
				idx = 0
			}
			var newList []Value
			if idx >= len(v[2].List) {
				newList = append(v[2].List, v[1])
			} else {
				newList = append(v[2].List[:idx+1], v[2].List[idx:]...)
				newList[idx] = v[1]
			}
			newListVal := Value{}
			newListVal.NewList(newList)
			return newListVal, nil

		},
	},
	{
		Identifier: "nth",
		NumArgs:    2,
		Function: func(v []Value) (Value, error) {
			err := checKTypes(v[0:1], []string{NumType})
			if err != nil {
				return Value{}, err
			}
			if v[1].Kind != StringType && v[1].Kind != ListType {
				return Value{}, types.Error{Simple: fmt.Sprintf("Type error - argument 1 of nth: expected type String or List , got %s", v[0].Kind)}
			}
			idx := int(v[0].Num)
			if idx < 0 || (v[1].Kind == ListType && idx >= len(v[1].List)) || (v[1].Kind == StringType && idx >= len(v[1].String)) {
				v := Value{}
				v.NewNull()
				return v, nil
			}
			if v[1].Kind == StringType {
				c := v[1].String[idx]
				val := Value{}
				val.NewString(string(c))
				return val, nil
			}
			return v[1].List[idx], nil
		},
	},
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
