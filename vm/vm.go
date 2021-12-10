package vm

import (
	"fmt"

	"github.com/benbanerjeerichards/lisp-calculator/calc"
	"github.com/benbanerjeerichards/lisp-calculator/eval"
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
