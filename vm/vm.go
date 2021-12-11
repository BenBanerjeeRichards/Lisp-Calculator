package vm

import (
	"fmt"
	"time"

	"github.com/benbanerjeerichards/lisp-calculator/calc"
	"github.com/benbanerjeerichards/lisp-calculator/eval"
)

type Instruction struct {
	Opcode int
	Arg1   int
	Arg2   int
}

func (i Instruction) String() string {
	return fmt.Sprintf("%s %d %d", opcodeToString(i.Opcode), i.Arg1, i.Arg2)
}

type Frame struct {
	Code              []Instruction
	Constants         []Value
	Variables         []Value
	VariableMap       map[string]int
	FunctionArguments []string
	// The root node of the frame hierarchy
	IsRootFrame bool
}

func (f *Frame) New() {
	f.Code = make([]Instruction, 0)
	f.Constants = make([]Value, 0)
	f.Variables = make([]Value, 0)
	f.VariableMap = make(map[string]int)
	f.FunctionArguments = make([]string, 0)
	f.IsRootFrame = false
}

func (f *Frame) Emit(opcode int) {
	f.Code = append(f.Code, Instruction{Opcode: opcode})
}

func (f *Frame) EmitUnary(opcode int, arg1 int) {
	f.Code = append(f.Code, Instruction{Opcode: opcode, Arg1: arg1})
}

func (f *Frame) EmitBinary(opcode int, arg1 int, arg2 int) {
	f.Code = append(f.Code, Instruction{Opcode: opcode, Arg1: arg1, Arg2: arg2})
}

type ClosureValue struct {
	Args []string
	Body *Frame
}

func EvalInstructions(compileRes CompileResult) Value {
	stack := []Value{}
	return evalInstructions(&compileRes.GlobalVariables, compileRes.Functions, compileRes.Frame, stack)
}

func evalInstructions(globalVariables *[]Value, functions []*Frame, frame Frame, stack []Value) Value {
	// POC, very inefficient (especially the stack)
	// fmt.Println("==========================")
	// for _, instr := range frame.Code {
	// 	fmt.Println(instr)
	// }
	// fmt.Println("--------------------------")

	pc := 0
	// TODO error handling for runtime errors
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
		case LOAD_GLOBAL:
			stack = append(stack, (*globalVariables)[instr.Arg1])
		case STORE_GLOBAL:
			(*globalVariables)[instr.Arg1] = stack[len(stack)-1]
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
			function := functions[instr.Arg1]
			val := evalInstructions(globalVariables, functions, *function, stack)
			stack = append(stack, val)
		case PUSH_CLOSURE_VAR:
			closure := stack[len(stack)-1]
			if closure.Kind != ClosureType {
				// TODO error handling
				panic("bad type - expected closure")
			}
			closure.Closure.Body.Variables[instr.Arg2] = frame.Variables[instr.Arg1]
			stack[len(stack)-1] = closure
		case PUSH_GLOBAL_CLOSURE_VAR:
			closure := stack[len(stack)-1]
			if closure.Kind != ClosureType {
				// TODO error handling
				panic("bad type - expected closure")
			}
			closure.Closure.Body.Variables[instr.Arg2] = (*globalVariables)[instr.Arg2]
			stack[len(stack)-1] = closure

		case CALL_CLOSURE:
			closure := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if closure.Kind != ClosureType {
				panic("bad type closure")
			}
			val := evalInstructions(globalVariables, functions, *closure.Closure.Body, stack)
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
	(defun isPrime (n)
    (def prime true)
    (if (= n 1)
        false
        ((def i 2)
        (while (and prime (<= i (sqrt n)))
            (if (= 0 (mod n i))(def prime false))
            (def i (+ i 1)))))
    (prime)
	)	

	(defun euler7 ()
		(def primeCount 0)
		(def i 2)
		(while (not (= primeCount 10001))
			(if (isPrime i)(
				(def primeCount (+ 1 primeCount))))
			(def i (+ i 1))
		)
		(- i 1)
	)
	(euler7)
	`)
	if err != nil {
		fmt.Println("AST error", err)
		return
	}

	c := Compiler{}
	c.New()
	compileStart := time.Now()
	frame, err := c.CompileProgram(astRes.Asts)
	if err != nil {
		fmt.Println("COMPILE error", err)
		return
	}
	fmt.Printf("Compiled in %s\n", time.Since(compileStart))

	runStart := time.Now()
	val := EvalInstructions(frame)
	fmt.Printf("Ran in %s\n", time.Since(runStart))

	fmt.Println(val.ToString())
}
