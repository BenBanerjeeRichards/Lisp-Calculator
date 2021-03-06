package vm

import (
	"fmt"
	"math"
	"strings"
)

const (
	NumType     = "num"
	BoolType    = "bool"
	StringType  = "string"
	NullType    = "null"
	ListType    = "list"
	ClosureType = "closure"
	StructType  = "struct"
)

// Value is a runtime value
type Value struct {
	Kind    string
	Num     float64
	Bool    bool
	String  string
	List    []Value
	Closure ClosureValue
	Struct  StructValue
}

type ClosureValue struct {
	Args []string
	Body *Frame
}

type StructValue struct {
	TypeName    string
	FieldNames  []string
	FieldValues []Value
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
func (v *Value) NewClosure(args []string, body *Frame) {
	v.Kind = ClosureType
	v.Closure = ClosureValue{Args: args, Body: body}
}

func (v *Value) NewStruct(structType string, fieldNames []string) {
	v.Kind = StructType
	v.Struct = StructValue{TypeName: structType, FieldNames: fieldNames, FieldValues: make([]Value, len(fieldNames))}
}

// Cant use Stringer interface due to name conflict
func (val Value) ToString() string {
	switch val.Kind {
	case NumType:
		if math.Abs(float64(int(val.Num))-val.Num) < 0.000001 {
			return fmt.Sprint(int(val.Num))
		}
		return fmt.Sprintf("%f", val.Num)
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
	case StructType:
		var str strings.Builder
		str.WriteString(fmt.Sprintf("%s{", val.Struct.TypeName))
		for i, fieldName := range val.Struct.FieldNames {
			str.WriteString(fmt.Sprintf("%s:%s,", fieldName, val.Struct.FieldValues[i].ToString()))
		}
		str.WriteString("}")
		return str.String()
	case "":
		return "<undef>"
	default:
		return fmt.Sprintf("Unknown type %v", val)
	}

}
