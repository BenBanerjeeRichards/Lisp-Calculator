package vm

import (
	"errors"
	"fmt"

	"github.com/benbanerjeerichards/lisp-calculator/ast"
	"github.com/benbanerjeerichards/lisp-calculator/types"
)

type Compiler struct {
	GlobalVariables   []Value
	GlobalVariableMap map[string]int
	Functions         []*Frame
	FunctionMap       map[string]int
}

func (c *Compiler) New() {
	c.Functions = make([]*Frame, 0)
	c.FunctionMap = make(map[string]int)
	c.GlobalVariableMap = make(map[string]int)
	c.GlobalVariables = make([]Value, 0)
}

type CompileResult struct {
	Frame           Frame
	Functions       []*Frame
	GlobalVariables []Value
	MainIndex       int
}

// CompileProgram compiles the given AST into bytecode
func (c *Compiler) CompileProgram(astResult ast.AstResult) (CompileResult, error) {
	frame := Frame{}
	frame.New()
	frame.IsRootFrame = true
	mainIndex := -1

	for _, exprOrStmt := range astResult.Asts {
		if exprOrStmt.Kind == ast.StmtType {
			switch stmt := exprOrStmt.Statement.(type) {
			case ast.VarDefStmt:
				err := c.CompileStatement(stmt, &frame)
				if err != nil {
					return CompileResult{}, err
				}
			case ast.FuncDefStmt:
				err := c.CompileStatement(stmt, &frame)
				if err != nil {
					return CompileResult{}, err
				}
			}
		}
	}

	if mainIdx, ok := c.FunctionMap["main"]; ok {
		mainIndex = mainIdx
		if len(c.Functions[mainIdx].FunctionArguments) == 1 {
			frame.Emit(PUSH_ARGS)
		}
		frame.EmitUnary(CALL_FUNCTION, mainIdx)
	} else {
		for _, exprOrStmt := range astResult.Asts {
			if exprOrStmt.Kind == ast.ExprType {
				err := c.CompileExpression(exprOrStmt.Expression, &frame)
				if err != nil {
					return CompileResult{}, err
				}
			} else {
				_, isFunction := exprOrStmt.Statement.(ast.FuncDefStmt)
				_, isGlobal := exprOrStmt.Statement.(ast.VarDefStmt)
				if !isFunction && !isGlobal {
					err := c.CompileStatement(exprOrStmt.Statement, &frame)
					if err != nil {
						return CompileResult{}, err
					}
				}

			}
		}
	}

	return CompileResult{Frame: frame, Functions: c.Functions, GlobalVariables: c.GlobalVariables, MainIndex: mainIndex}, nil
}

func (c *Compiler) compileBlock(asts []ast.Ast, frame *Frame) error {
	for _, exprOrStmt := range asts {
		if exprOrStmt.Kind == ast.ExprType {
			err := c.CompileExpression(exprOrStmt.Expression, frame)
			if err != nil {
				return err
			}
		} else if exprOrStmt.Kind == ast.StmtType {
			err := c.CompileStatement(exprOrStmt.Statement, frame)
			if err != nil {
				return err
			}
		} else {
			return errors.New("unknown expression type")
		}
	}
	return nil
}

func (c *Compiler) CompileExpression(exprNode ast.Expr, frame *Frame) error {
	switch expr := exprNode.(type) {
	case ast.NumberExpr:
		val := Value{}
		val.NewNum(expr.Value)
		frame.Constants = append(frame.Constants, val)
		frame.EmitUnary(LOAD_CONST, len(frame.Constants)-1)
	case ast.BoolExpr:
		val := Value{}
		val.NewBool(expr.Value)
		frame.Constants = append(frame.Constants, val)
		frame.EmitUnary(LOAD_CONST, len(frame.Constants)-1)
	case ast.StringExpr:
		val := Value{}
		val.NewString(expr.Value)
		frame.Constants = append(frame.Constants, val)
		frame.EmitUnary(LOAD_CONST, len(frame.Constants)-1)
	case ast.NullExpr:
		val := Value{}
		val.NewNull()
		frame.Constants = append(frame.Constants, val)
		frame.EmitUnary(LOAD_CONST, len(frame.Constants)-1)
	case ast.ListExpr:
		for _, listItem := range expr.Value {
			err := c.CompileExpression(listItem, frame)
			if err != nil {
				return err
			}
		}
		frame.EmitUnary(CREATE_LIST, len(expr.Value))
	case ast.IfElseExpr:
		err := c.CompileExpression(expr.Condition, frame)
		if err != nil {
			return err
		}
		frame.EmitUnary(COND_JUMP_FALSE, 0)
		condJumpInstrIdx := len(frame.Code) - 1
		err = c.compileBlock(expr.IfBranch, frame)
		if err != nil {
			return err
		}
		frame.Code[condJumpInstrIdx].Arg1 = len(frame.Code) - condJumpInstrIdx
		frame.EmitUnary(JUMP, 0)
		ifJumpIndx := len(frame.Code) - 1
		err = c.compileBlock(expr.ElseBranch, frame)
		if err != nil {
			return err
		}
		frame.Code[ifJumpIndx].Arg1 = len(frame.Code) - (ifJumpIndx + 1)
	case ast.IfOnlyExpr:
		err := c.CompileExpression(expr.Condition, frame)
		if err != nil {
			return err
		}
		frame.EmitUnary(COND_JUMP_FALSE, 0)
		condJumpInstrIdx := len(frame.Code) - 1
		err = c.compileBlock(expr.IfBranch, frame)
		if err != nil {
			return err
		}
		frame.Code[condJumpInstrIdx].Arg1 = len(frame.Code) - condJumpInstrIdx
		frame.EmitUnary(JUMP, 1)
		frame.Emit(STORE_NULL)
	case ast.VarUseExpr:
		if idx, ok := frame.VariableMap[expr.Identifier]; ok {
			frame.EmitUnary(LOAD_VAR, idx)
		} else if idx, ok := c.GlobalVariableMap[expr.Identifier]; ok {
			frame.EmitUnary(LOAD_GLOBAL, idx)
		} else {
			return errors.New("unknown variable " + expr.Identifier)
		}
	case ast.ClosureDefExpr:
		closureFrame := Frame{}
		closureFrame.New()
		// Capture all variables in current scope
		closureFrame.VariableMap = frame.VariableMap
		closureFrame.Variables = frame.Variables
		// Capture globals as vars
		for range c.GlobalVariables {
			closureFrame.Variables = append(closureFrame.Variables, Value{})
		}
		for globalName, globalIdx := range c.GlobalVariableMap {
			closureFrame.VariableMap[globalName] = globalIdx + len(frame.Variables)
		}

		// Push arguments onto stack
		for _, argName := range expr.Args {
			closureFrame.Variables = append(closureFrame.Variables, Value{})
			closureFrame.VariableMap[argName] = len(closureFrame.Variables) - 1
			closureFrame.EmitUnary(STORE_VAR, len(closureFrame.Variables)-1)
		}
		err := c.compileBlock(expr.Body, &closureFrame)
		if err != nil {
			return err
		}

		// Closure is a value that needs to be pushed to top of stack
		closureValue := Value{}
		closureValue.NewClosure(expr.Args, &closureFrame)
		frame.Constants = append(frame.Constants, closureValue)
		frame.EmitUnary(LOAD_CONST, len(frame.Constants)-1)

		// Now capture the values of the variables
		for sourceIndex := range frame.Variables {
			frame.EmitBinary(PUSH_CLOSURE_VAR, sourceIndex, sourceIndex)
		}
		// Now capture globals - closures can not access or modify globals, only capture them into variables
		// From langauge user POV, this means that globals can only be read from closure and and changes exist only within
		// the closure
		for globalIndex := range c.GlobalVariables {
			// Target index set  - see above capturing logic.
			// Variables are in this order: <captured vars><captured globals><lambda arguments><closure arguments>
			frame.EmitBinary(PUSH_GLOBAL_CLOSURE_VAR, globalIndex, len(frame.Variables)+globalIndex)
		}
	case ast.ClosureApplicationExpr:
		for _, arg := range expr.Args {
			err := c.CompileExpression(arg, frame)
			if err != nil {
				return err
			}
		}
		err := c.CompileExpression(expr.Closure, frame)
		if err != nil {
			return err
		}
		frame.Emit(CALL_CLOSURE)
	case ast.FunctionApplicationExpr:
		for _, arg := range expr.Args {
			err := c.CompileExpression(arg, frame)
			if err != nil {
				return err
			}
		}
		if expr.Identifier == "+" {
			if len(expr.Args) != 2 {
				return types.Error{Range: expr.GetRange(), Simple: fmt.Sprintf("Expected 2 arguments, got %d", len(expr.Args))}
			}
			frame.Emit(ADD)
		} else {
			if idx, builtinFunc, ok := lookupBuiltin(expr.Identifier); ok {
				if len(expr.Args) != builtinFunc.NumArgs {
					return types.Error{Range: expr.GetRange(), Simple: fmt.Sprintf("Expected %d arguments, got %d", builtinFunc.NumArgs, len(expr.Args))}
				}
				frame.EmitUnary(CALL_BUILTIN, idx)
			} else if idx, ok := frame.VariableMap[expr.Identifier]; ok {
				frame.EmitUnary(LOAD_VAR, idx)
			} else if idx, ok := c.GlobalVariableMap[expr.Identifier]; ok {
				frame.EmitUnary(LOAD_GLOBAL, idx)
			} else if idx, ok := c.FunctionMap[expr.Identifier]; ok {
				frame.EmitUnary(CALL_FUNCTION, idx)
			} else {
				return errors.New("unknown variable or function")

			}
		}
	default:
		return errors.New(fmt.Sprintf("unsupported ast type %s", exprNode))
	}

	return nil
}

func (c *Compiler) CompileStatement(stmtExpr ast.Stmt, frame *Frame) error {
	switch stmt := stmtExpr.(type) {
	case ast.VarDefStmt:
		err := c.CompileExpression(stmt.Value, frame)
		if err != nil {
			return err
		}
		// TODO clean this up
		if frame.IsRootFrame {
			idx, ok := c.GlobalVariableMap[stmt.Identifier]
			if !ok {
				c.GlobalVariables = append(c.GlobalVariables, Value{})
				idx = len(c.GlobalVariables) - 1
				c.GlobalVariableMap[stmt.Identifier] = idx
			}
			frame.EmitUnary(STORE_GLOBAL, idx)
		} else {
			idx, ok := frame.VariableMap[stmt.Identifier]
			if !ok {
				frame.Variables = append(frame.Variables, Value{})
				idx = len(frame.Variables) - 1
				frame.VariableMap[stmt.Identifier] = idx
			}
			frame.EmitUnary(STORE_VAR, idx)
		}
		frame.Emit(STORE_NULL)
	case ast.ImportStmt:
		frame.Emit(STORE_NULL)
	case ast.WhileStmt:
		condStartIdx := len(frame.Code) - 1
		err := c.CompileExpression(stmt.Condition, frame)
		if err != nil {
			return err
		}
		frame.EmitUnary(COND_JUMP_FALSE, 0)
		condJumpIdx := len(frame.Code) - 1
		err = c.compileBlock(stmt.Body, frame)
		if err != nil {
			return err
		}
		frame.Code[condJumpIdx].Arg1 = len(frame.Code) - condJumpIdx
		frame.EmitUnary(JUMP, condStartIdx-len(frame.Code))
		frame.Emit(STORE_NULL)
	case ast.FuncDefStmt:
		functionFrame := Frame{}
		functionFrame.New()
		for i, argName := range stmt.Args {
			functionFrame.VariableMap[argName] = i
			functionFrame.Variables = append(functionFrame.Variables, Value{})
			functionFrame.FunctionArguments = stmt.Args
			// Store each argument from the stack into the variables array
			functionFrame.EmitUnary(STORE_VAR, len(stmt.Args)-(i+1))
		}
		err := c.compileBlock(stmt.Body, &functionFrame)
		if err != nil {
			return err
		}
		c.Functions = append(c.Functions, &functionFrame)
		functionId := len(c.Functions) - 1
		c.FunctionMap[stmt.Identifier] = functionId
		frame.Emit(STORE_NULL)
	default:
		return errors.New("unsupported statement")
	}
	return nil
}

func lookupBuiltin(identifier string) (int, Builtin, bool) {
	for i, item := range Builtins {
		if item.Identifier == identifier {
			return i, item, true
		}
	}

	return 0, Builtin{}, false
}
