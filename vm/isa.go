package vm

import "fmt"

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