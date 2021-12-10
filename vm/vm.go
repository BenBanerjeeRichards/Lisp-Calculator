package vm

import (
	"fmt"
	"strings"

	"github.com/benbanerjeerichards/lisp-calculator/ast"
	"github.com/benbanerjeerichards/lisp-calculator/calc"
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

	COND_JUMP_FALSE

	// LOAD_CONST <constantRef>
	// Pushes the constant at index constantRef to to the top of the stack
	LOAD_CONST

	// LOAD_VAR <varIndex>
	// Load variable at index varIndex to the top of the stack
	LOAD_VAR

	// STORE_VAR <varIndex>
	// Store value at top of stack into variable at index varIndex
	STORE_VAR

	// CALL_FUNCTION <functionIndex>
	// Calls the function <functionIndex>
	// Arguments must be passed into stack first (same as with builtins)
	CALL_FUNCTION

	// CALL_BUILTIN <builtinIdx>
	// Call builtin at index in vm.Builtins[builtinIdx]
	CALL_BUILTIN

	// CREATE_LIST <N>
	// Create a list of N elements (takes N elements from stack)
	CREATE_LIST

	// STORE_NULL
	// Push Null onto the stach
	STORE_NULL

	// RETURN
	// Return from running current frame
	RETURN
)

func opcodeToString(op int) string {
	switch op {
	case POP:
		return "POP"
	case ADD:
		return "ADD"
	case JUMP:
		return "JUMP"
	case COND_JUMP:
		return "COND_JUMP"
	case COND_JUMP_FALSE:
		return "COND_JUMP_FALSE"

	case LOAD_CONST:
		return "LOAD_CONST"
	case LOAD_VAR:
		return "LOAD_VAR"
	case STORE_VAR:
		return "STORE_VAR"
	case CALL_FUNCTION:
		return "CALL_FUNCTION"
	case CALL_BUILTIN:
		return "CALL_BUILTIN"
	case CREATE_LIST:
		return "CREATE_LIST"
	case STORE_NULL:
		return "STORE_NULL"
	case RETURN:
		return "RETURN"
	default:
		return fmt.Sprintf("<%d>", op)
	}
}

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

func (i Instruction) String() string {
	return fmt.Sprintf("%s %d", opcodeToString(i.Opcode), i.Arg1)
}

type Frame struct {
	Code              []Instruction
	Constants         []Value
	Variables         []Value
	VariableMap       map[string]int
	Functions         []*Frame
	FunctionMap       map[string]int
	FunctionArguments []string
}

func (f *Frame) New() {
	f.Code = make([]Instruction, 0)
	f.Constants = make([]Value, 0)
	f.Variables = make([]Value, 0)
	f.VariableMap = make(map[string]int)
	f.Functions = make([]*Frame, 0)
	f.FunctionMap = make(map[string]int)
	f.FunctionArguments = make([]string, 0)
}

func (f *Frame) Emit(opcode int) {
	f.Code = append(f.Code, Instruction{Opcode: opcode})
}

func (f *Frame) EmitUnary(opcode int, arg1 int) {
	f.Code = append(f.Code, Instruction{Opcode: opcode, Arg1: arg1})
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
		return fmt.Sprintf("Unknown type %v", val)
	}

}

func EvalInstructions(frame Frame) Value {
	stack := []Value{}
	return evalInstructions(frame, stack)
}

func evalInstructions(frame Frame, stack []Value) Value {
	// POC, very inefficient (especially the stack)
	// for _, instr := range frame.Code {
	// 	fmt.Println(instr)
	// }
	pc := 0

out:
	for pc < len(frame.Code) {
		instr := frame.Code[pc]
		switch instr.Opcode {
		case LOAD_CONST:
			stack = append(stack, frame.Constants[instr.Arg1])
		case STORE_NULL:
			v := Value{}
			v.NewNull()
			stack = append(stack, v)
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
		case COND_JUMP_FALSE:
			val := stack[len(stack)-1]
			checkType(eval.BoolType, val.Kind)
			stack = stack[0 : len(stack)-1]
			if !val.Bool {
				pc += instr.Arg1
			}
		case JUMP:
			pc += instr.Arg1
		case LOAD_VAR:
			stack = append(stack, frame.Variables[instr.Arg1])
		case STORE_VAR:
			frame.Variables[instr.Arg1] = stack[len(stack)-1]
			stack = stack[0 : len(stack)-1]
		case CALL_BUILTIN:
			builtin := Builtins[instr.Arg1]
			// FIXME handle error
			res, _ := builtin.Function(stack[len(stack)-(builtin.NumArgs):])
			stack = stack[0 : len(stack)-(builtin.NumArgs)]
			stack = append(stack, res)
		case CREATE_LIST:
			list := make([]Value, instr.Arg1)
			for i := 0; i < instr.Arg1; i++ {
				list[instr.Arg1-(i+1)] = stack[len(stack)-(1+i)]
			}
			stack = stack[0 : len(stack)-instr.Arg1]
			val := Value{}
			val.NewList(list)
			stack = append(stack, val)
		case RETURN:
			break out
		case CALL_FUNCTION:
			// TODO handle globals
			function := frame.Functions[instr.Arg1]
			val := evalInstructions(*function, stack)
			stack = append(stack, val)
		default:
			fmt.Println("Unknown instruction", instr)
		}
		pc += 1
	}

	return stack[len(stack)-1]
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
	astRes, err := calc.Ast(`
	(defun f (a b) (/ a b))
	(f 10 2)`)
	if err != nil {
		fmt.Println("AST error", err)
		return
	}

	c := Compiler{}
	frame, err := c.CompileFunction(astRes.Asts)
	if err != nil {
		fmt.Println("COMPILE error", err)
		return
	}

	val := EvalInstructions(frame)
	fmt.Println(val.ToString())
}
