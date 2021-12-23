package vm

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/benbanerjeerichards/lisp-calculator/types"
)

type Frame struct {
	Code              []Instruction
	Constants         []Value
	Variables         []Value
	VariableMap       map[string]int
	FunctionArguments []string
	// The root node of the frame hierarchy
	IsRootFrame bool
	// LineMap maps from opcode index to line number
	LineMap      []int
	FunctionName string
}

func (f *Frame) New() {
	f.Code = make([]Instruction, 0)
	f.Constants = make([]Value, 0)
	f.Variables = make([]Value, 0)
	f.VariableMap = make(map[string]int)
	f.FunctionArguments = make([]string, 0)
	f.IsRootFrame = false
	f.LineMap = []int{}
	f.FunctionName = "."
}

func (f *Frame) Emit(opcode int, lineNumber int) {
	f.LineMap = append(f.LineMap, lineNumber)
	f.Code = append(f.Code, Instruction{Opcode: opcode})
}

func (f *Frame) EmitUnary(opcode int, arg1 int, lineNumber int) {
	f.LineMap = append(f.LineMap, lineNumber)
	f.Code = append(f.Code, Instruction{Opcode: opcode, Arg1: arg1})
}

func (f *Frame) EmitBinary(opcode int, arg1 int, arg2 int, lineNumber int) {
	f.LineMap = append(f.LineMap, lineNumber)
	f.Code = append(f.Code, Instruction{Opcode: opcode, Arg1: arg1, Arg2: arg2})
}

func Eval(compileRes CompileResult, programArgs []string, debug bool) (Value, error) {
	evalulator := Evalulator{stack: []Value{},
		globalVariables: &compileRes.GlobalVariables,
		functions:       compileRes.Functions,
		programArgs:     programArgs,
		functionNames:   compileRes.FunctionNames,
		printProfile:    debug,
		profileWriter:   tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)}

	val, err := evalulator.evalInstructions(compileRes.Frame)
	if debug {
		evalulator.profileWriter.Flush()
		fmt.Println("Final stack: ", stackToString(evalulator.stack))
	}
	return val, err
}

type Evalulator struct {
	programArgs     []string
	globalVariables *[]Value
	functions       []*Frame
	stack           []Value
	functionNames   []string

	// printProfile is true if we should print a profile of all instructions executed, along with the resulting stack
	printProfile  bool
	profileWriter *tabwriter.Writer
}

func (e *Evalulator) evalInstructions(frame Frame) (Value, error) {
	pc := 0

out:
	for pc < len(frame.Code) {
		instr := frame.Code[pc]
		if e.printProfile {
			e.profileInstruction(pc, instr, &frame)
		}
		switch instr.Opcode {
		case LOAD_CONST:
			e.stack = append(e.stack, frame.Constants[instr.Arg1])
		case STORE_NULL:
			v := Value{}
			v.NewNull()
			e.stack = append(e.stack, v)
		case POP:
			e.stack = e.stack[0 : len(e.stack)-1]
		case ADD:
			lhs := e.stack[len(e.stack)-1]
			rhs := e.stack[len(e.stack)-2]
			e.stack = e.stack[0 : len(e.stack)-2]
			val := Value{}
			if lhs.Kind != NumType {
				return Value{}, RuntimeError{Line: frame.LineMap[pc], Simple: fmt.Sprintf("Type error -  expected type Num for first argument to add, got %s", lhs.Kind)}
			}
			if rhs.Kind != NumType {
				return Value{}, RuntimeError{Line: frame.LineMap[pc], Simple: fmt.Sprintf("Type error -  expected type Num for second argument to add, got %s", lhs.Kind)}
			}
			val.NewNum(lhs.Num + rhs.Num)
			e.stack = append(e.stack, val)
		case COND_JUMP:
			val := e.stack[len(e.stack)-1]
			if val.Kind != BoolType {
				return Value{}, RuntimeError{Line: frame.LineMap[pc], Simple: fmt.Sprintf("Type error -  expected type Bool for condition, got %s", val.Kind)}
			}
			e.stack = e.stack[0 : len(e.stack)-1]
			if val.Bool {
				pc += instr.Arg1
			}
		case COND_JUMP_FALSE:
			val := e.stack[len(e.stack)-1]
			if val.Kind != BoolType {
				return Value{}, RuntimeError{Line: frame.LineMap[pc], Simple: fmt.Sprintf("Type error -  expected type Bool for condition, got %s", val.Kind)}
			}
			e.stack = e.stack[0 : len(e.stack)-1]
			if !val.Bool {
				pc += instr.Arg1
			}
		case JUMP:
			pc += instr.Arg1
		case LOAD_VAR:
			e.stack = append(e.stack, frame.Variables[instr.Arg1])
		case STORE_VAR:
			frame.Variables[instr.Arg1] = e.stack[len(e.stack)-1]
			e.stack = e.stack[0 : len(e.stack)-1]
		case LOAD_GLOBAL:
			e.stack = append(e.stack, (*e.globalVariables)[instr.Arg1])
		case STORE_GLOBAL:
			(*e.globalVariables)[instr.Arg1] = e.stack[len(e.stack)-1]
			e.stack = e.stack[0 : len(e.stack)-1]
		case CALL_BUILTIN:
			builtin := Builtins[instr.Arg1]
			res, err := builtin.Function(e.stack[len(e.stack)-(builtin.NumArgs):])
			if err != nil {
				if stdErr, ok := err.(types.Error); ok {
					return Value{}, RuntimeError{Simple: stdErr.Simple, Detail: stdErr.Detail, Line: frame.LineMap[pc]}
				}
				return Value{}, nil
			}
			e.stack = e.stack[0 : len(e.stack)-(builtin.NumArgs)]
			e.stack = append(e.stack, res)
		case CREATE_LIST:
			list := make([]Value, instr.Arg1)
			for i := 0; i < instr.Arg1; i++ {
				list[instr.Arg1-(i+1)] = e.stack[len(e.stack)-(1+i)]
			}
			e.stack = e.stack[0 : len(e.stack)-instr.Arg1]
			val := Value{}
			val.NewList(list)
			e.stack = append(e.stack, val)
		case RETURN:
			break out
		case PUSH_ARGS:
			argVal := Value{}
			argsAsValues := []Value{}
			for _, arg := range e.programArgs {
				val := Value{}
				val.NewString(arg)
				argsAsValues = append(argsAsValues, val)
			}
			argVal.NewList(argsAsValues)
			e.stack = append(e.stack, argVal)
		case CALL_FUNCTION:
			if e.printProfile {
				e.profileNewLine()
			}
			function := e.functions[instr.Arg1]
			stackIndex := len(e.stack) - (len(function.FunctionArguments) + 1)
			val, err := e.evalInstructions(*function)
			if err != nil {
				return Value{}, err
			}
			e.stack = e.stack[:stackIndex+1]
			e.stack = append(e.stack, val)
		case PUSH_CLOSURE_VAR:
			closure := e.stack[len(e.stack)-1]
			if closure.Kind != ClosureType {
				return Value{}, RuntimeError{Line: frame.LineMap[pc], Simple: fmt.Sprintf("Type error -  expected Closure, got %s", closure.Kind)}
			}
			closure.Closure.Body.Variables[instr.Arg2] = frame.Variables[instr.Arg1]
			e.stack[len(e.stack)-1] = closure
		case PUSH_GLOBAL_CLOSURE_VAR:
			closure := e.stack[len(e.stack)-1]
			if closure.Kind != ClosureType {
				if closure.Kind != ClosureType {
					return Value{}, RuntimeError{Line: frame.LineMap[pc], Simple: fmt.Sprintf("Type error -  expected Closure, got %s", closure.Kind)}
				}
			}
			closure.Closure.Body.Variables[instr.Arg2] = (*e.globalVariables)[instr.Arg2]
			e.stack[len(e.stack)-1] = closure

		case CALL_CLOSURE:
			closure := e.stack[len(e.stack)-1]
			e.stack = e.stack[:len(e.stack)-1]
			if closure.Kind != ClosureType {
				if closure.Kind != ClosureType {
					return Value{}, RuntimeError{Line: frame.LineMap[pc], Simple: fmt.Sprintf("Type error -  expected Closure, got %s", closure.Kind)}
				}
			}
			val, err := e.evalInstructions(*closure.Closure.Body)
			if err != nil {
				return Value{}, err
			}
			e.stack = append(e.stack, val)
		default:
			fmt.Println("Unknown instruction", instr)
		}
		pc += 1
	}
	if e.printProfile {
		e.profileNewLine()
	}
	val := e.stack[len(e.stack)-1]
	e.stack = e.stack[:len(e.stack)-1]
	return val, nil
}

func (e *Evalulator) profileInstruction(pc int, instr Instruction, frame *Frame) {
	str := fmt.Sprintf("%d\t%s\t%s\t%s\t%s\t", pc, frame.FunctionName, opcodeToString(instr.Opcode), instr.Detail(frame, e.functionNames), stackToString(e.stack))
	str = strings.ReplaceAll(str, "\n", "\\n")
	fmt.Fprintf(e.profileWriter, str+"\n")
}
func (e *Evalulator) profileNewLine() {
	fmt.Fprint(e.profileWriter, "\t\t\t\t\n")

}
func printStack(stack []Value) {
	fmt.Println("============= stack =============")
	for i, item := range stack {
		fmt.Println(i, item.ToString())
	}
}

func stackToString(stack []Value) string {
	var sb strings.Builder
	for _, val := range stack {
		sb.WriteString(val.ToString())
		sb.WriteString(" ")
	}
	return sb.String()
}
