package vm

import (
	"fmt"
	"strings"

	"github.com/benbanerjeerichards/lisp-calculator/ast"
	"github.com/benbanerjeerichards/lisp-calculator/eval"
)

const (
	// POP
	// Remove element from top of stack
	POP = iota

	// ADD
	// Adds top two elements numbers of the stack
	ADD

	// JUMP <relativeOffset>
	// Unconditional jump to PC + relativeOffset
	JUMP

	// COND_JUMP <relativeOffset>
	// Jumps by relativeOffset amount from current PC only if top of stack is true
	COND_JUMP

	// LOAD_CONST <constantRef>
	// Pushes the constant at index constantRef to to the top of the stack
	LOAD_CONST

	// LOAD_VAR <varIndex>
	// Load variable at index varIndex to the top of the stack
	LOAD_VAR

	// STORE_VAR <varIndex>
	// Store value at top of stack into variable at index varIndex
	STORE_VAR
)

const (
	NumType     = "num"
	BoolType    = "bool"
	StringType  = "string"
	NullType    = "null"
	ListType    = "list"
	ClosureType = "closure"
)

type Instruction struct {
	Opcode int
	Arg1   int
}

type Frame struct {
	code      []Instruction
	constants []Value
	variables []Value
}

func (f *Frame) New() {
	f.code = make([]Instruction, 0)
	f.constants = make([]Value, 0)
	f.variables = make([]Value, 0)
}

func (f *Frame) Emit(opcode int) {
	f.code = append(f.code, Instruction{Opcode: opcode})
}

func (f *Frame) EmitUnary(opcode int, arg1 int) {
	f.code = append(f.code, Instruction{Opcode: opcode, Arg1: arg1})
}

type ClosureValue struct {
	Args []string
	Body Frame
}

// Value is a runtime value
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
func (v *Value) NewClosure(args []string, body []ast.Ast) {
	v.Kind = ClosureType
	v.Closure = ClosureValue{Args: args}
}

// Cant use Stringer interface due to name conflict
func (val Value) ToString() string {
	switch val.Kind {
	case NumType:
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
	default:
		return "Unknown type"
	}

}

func EvalInstructions(functions []*Frame, frame Frame) {
	// POC, very inefficient (especially the stack)
	pc := 0
	stack := []Value{}

	for pc < len(frame.code) {
		instr := frame.code[pc]
		switch instr.Opcode {
		case LOAD_CONST:
			stack = append(stack, frame.constants[instr.Arg1])
		case POP:
			stack = stack[0 : len(stack)-1]
		case ADD:
			lhs := stack[len(stack)-1]
			rhs := stack[len(stack)-2]
			stack = stack[0 : len(stack)-2]
			val := Value{}
			checkType(eval.NumType, lhs.Kind)
			checkType(eval.NumType, rhs.Kind)
			val.NewNum(lhs.Num + rhs.Num)
			stack = append(stack, val)
		case COND_JUMP:
			val := stack[len(stack)-1]
			checkType(eval.BoolType, val.Kind)
			stack = stack[0 : len(stack)-1]
			if val.Bool {
				pc += instr.Arg1
			}
		case JUMP:
			pc += instr.Arg1
		case LOAD_VAR:
			stack = append(stack, frame.variables[instr.Arg1])
		case STORE_VAR:
			frame.variables[instr.Arg1] = stack[len(stack)-1]
			stack = stack[0 : len(stack)-1]
		default:
			fmt.Println("Unknown instruction", instr)
		}
		pc += 1
	}
	printStack(stack)
}

// TODO remove this, just for easy debugging
func checkType(expected string, actual string) {
	if expected != actual {
		panic(fmt.Sprintf("Bad type - expected %s got %s", expected, actual))
	}
}

func printStack(stack []Value) {
	fmt.Println("============= stack =============")
	for i, item := range stack {
		fmt.Println(i, item.ToString())
	}
}

func Main() {
	c := Compiler{}

	c.CompileFunction()
}
