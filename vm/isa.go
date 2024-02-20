package vm

import "fmt"

const (
	// POP
	// Remove element from top of stack
	POP = iota

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

	// LOAD_GLOBAL <globalIdx>
	// Push global at given index onto top of stack
	LOAD_GLOBAL

	// STORE_GLOBAL <globalIdx>
	// Store value at top of stack into given global index
	STORE_GLOBAL

	// PUSH_CLOSURE_VAR <sourceIdx> <targetIdx>
	// Store variable at current from index sourceIdx into closure (at TOS) frame index targetIdx
	PUSH_CLOSURE_VAR

	// PUSH_GLOBAL_CLOSURE_VAR <globalIdx> <targetIdx>
	// Push global at globalIdx into closure (at TOS) variable targetIdx (This means closures capture global)
	PUSH_GLOBAL_CLOSURE_VAR

	// Call the closure at the top of stack
	CALL_CLOSURE

	// PUSH_ARGS pushes command line arguments onto stack as List<String>
	PUSH_ARGS

	// CREATE_STRUCT <structDeclarationIndex>
	// Create a new struct of type at index structDeclarationIndex
	CREATE_STRUCT

	// SET_STRUCT_FIELD
	// Given a struct at TOS-2, a struct field index at TOS-1 and a value at TOS, sets the struct field value
	SET_STRUCT_FIELD

	// STRUCT_FIELD_INDEX <nameIndex>
	// For a struct at TOS, pushes the index of the field at names[nameIndex] of struct TOS.
	// Struct NOT popped from stack
	STRUCT_FIELD_INDEX

	// GET_STRUCT_FIELD <fieldIdx>
	// For struct at top of stack, push field value at fieldIdx onto top of stack
	GET_STRUCT_FIELD
)

func opcodeToString(op int) string {
	switch op {
	case POP:
		return "POP"
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
	case STORE_GLOBAL:
		return "STORE_GLOBAL"
	case LOAD_GLOBAL:
		return "LOAD_GLOBAL"
	case PUSH_CLOSURE_VAR:
		return "PUSH_CLOSURE_VAR"
	case PUSH_GLOBAL_CLOSURE_VAR:
		return "PUSH_GLOBAL_CLOSURE_VAR"
	case CALL_CLOSURE:
		return "CALL_CLOSURE"
	case PUSH_ARGS:
		return "PUSH_ARGS"
	case CREATE_STRUCT:
		return "CREATE_STRUCT"
	case SET_STRUCT_FIELD:
		return "SET_STRUCT_FIELD"
	case GET_STRUCT_FIELD:
		return "GET_STRUCT_FIELD"
	case STRUCT_FIELD_INDEX:
		return "STRUCT_FIELD_INDEX"
	default:
		return fmt.Sprintf("<%d>", op)
	}
}

type Instruction struct {
	Opcode int
	Arg1   int
	Arg2   int
}

func (i Instruction) String() string {
	return fmt.Sprintf("%s %d %d", opcodeToString(i.Opcode), i.Arg1, i.Arg2)
}

func (i Instruction) DetailedString(frame *Frame, functionNames []string) string {
	str := opcodeToString(i.Opcode)
	return str + i.Detail(frame, functionNames)
}

func (i Instruction) Detail(frame *Frame, functionNames []string) string {
	showArg1 := true
	showArg2 := i.Opcode == PUSH_GLOBAL_CLOSURE_VAR || i.Opcode == PUSH_CLOSURE_VAR
	detail := ""
	if i.Opcode == LOAD_CONST {
		detail = frame.Constants[i.Arg1].ToString()
	}
	if i.Opcode == CALL_BUILTIN {
		detail = Builtins[i.Arg1].Identifier
	}
	if i.Opcode == CALL_FUNCTION {
		detail = functionNames[i.Arg1]
	}
	if i.Opcode == LOAD_VAR {
		if len(frame.Variables) <= i.Arg1 {
			detail = "<ERROR>"
		} else {
			v := frame.Variables[i.Arg1]
			if v.Kind != "" {
				detail = v.ToString()
			}
		}
	}
	str := ""

	if showArg1 {
		str += " " + fmt.Sprintf("%d", i.Arg1)
	}
	if showArg2 {
		str += " " + fmt.Sprintf("%d", i.Arg1)
	}
	if len(detail) > 0 {
		str += " (" + detail + ")"
	}

	return str

}
