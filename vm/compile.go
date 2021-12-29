package vm

import (
	"errors"
	"fmt"

	"github.com/benbanerjeerichards/lisp-calculator/ast"
	"github.com/benbanerjeerichards/lisp-calculator/types"
	"github.com/davecgh/go-spew/spew"
)

type Compiler struct {
	GlobalVariables   []Value
	GlobalVariableMap map[string]int
	Functions         []*Frame
	FunctionMap       map[string]int
	FunctionNames     []string
	Structs           [][]string
	StructMap         map[string]int
}

func (c *Compiler) New() {
	c.Functions = make([]*Frame, 0)
	c.FunctionMap = make(map[string]int)
	c.GlobalVariableMap = make(map[string]int)
	c.GlobalVariables = make([]Value, 0)
	c.Structs = make([][]string, 0)
	c.StructMap = make(map[string]int)
}

type CompileResult struct {
	Frame           Frame
	Functions       []*Frame
	GlobalVariables []Value
	MainIndex       int
	FunctionNames   []string
	Structs         []StructDecl
}

type StructDecl struct {
	Name       string
	FieldNames []string
}

// CompileProgram compiles the given AST into bytecode
func (c *Compiler) CompileProgram(asts []ast.Ast) (CompileResult, error) {
	frame := Frame{}
	frame.New()
	frame.IsRootFrame = true
	mainIndex := -1

	for _, exprOrStmt := range asts {
		if exprOrStmt.Kind == ast.StmtType {
			switch stmt := exprOrStmt.Statement.(type) {
			case ast.FuncDefStmt:
				c.Functions = append(c.Functions, &Frame{})
				c.FunctionMap[stmt.Identifier] = len(c.Functions) - 1
				c.FunctionNames = append(c.FunctionNames, stmt.Identifier)
			case ast.StructDefStmt:
				c.Structs = append(c.Structs, stmt.FieldNames)
				c.StructMap[stmt.Identifier] = len(c.Structs) - 1
			}
		}
	}

	for _, exprOrStmt := range asts {
		if exprOrStmt.Kind == ast.StmtType {
			switch stmt := exprOrStmt.Statement.(type) {
			case ast.VarDefStmt, ast.FuncDefStmt:
				err := c.compileStatement(stmt, &frame)
				if err != nil {
					return CompileResult{}, err
				}
			}
		}
	}

	if mainIdx, ok := c.FunctionMap["main"]; ok {
		mainIndex = mainIdx
		if len(c.Functions[mainIdx].FunctionArguments) == 1 {
			frame.Emit(PUSH_ARGS, -1)
		} else if len(c.Functions[mainIdx].FunctionArguments) > 1 {
			return CompileResult{}, types.Error{Simple: "Main function must take zero of one argument"}
		}
		frame.EmitUnary(CALL_FUNCTION, mainIdx, -1)
	} else {
		for _, exprOrStmt := range asts {
			if exprOrStmt.Kind == ast.ExprType {
				err := c.compileExpression(exprOrStmt.Expression, &frame)
				if err != nil {
					return CompileResult{}, err
				}
			} else {
				_, isFunction := exprOrStmt.Statement.(ast.FuncDefStmt)
				_, isGlobal := exprOrStmt.Statement.(ast.VarDefStmt)
				_, isStructDef := exprOrStmt.Statement.(ast.StructDefStmt)
				if !isFunction && !isGlobal && !isStructDef {
					err := c.compileStatement(exprOrStmt.Statement, &frame)
					if err != nil {
						return CompileResult{}, err
					}
				}

			}
		}
	}

	structs := make([]StructDecl, len(c.Structs))
	for name, fieldIdx := range c.StructMap {
		structs[fieldIdx] = StructDecl{Name: name, FieldNames: c.Structs[fieldIdx]}
	}

	return CompileResult{Frame: frame, Functions: c.Functions, GlobalVariables: c.GlobalVariables,
		MainIndex: mainIndex, FunctionNames: c.FunctionNames, Structs: structs}, nil
}

func (c *Compiler) compileBlock(asts []ast.Ast, frame *Frame) error {
	for _, exprOrStmt := range asts {
		if exprOrStmt.Kind == ast.ExprType {
			err := c.compileExpression(exprOrStmt.Expression, frame)
			if err != nil {
				return err
			}
		} else if exprOrStmt.Kind == ast.StmtType {
			err := c.compileStatement(exprOrStmt.Statement, frame)
			if err != nil {
				return err
			}
		} else {
			return errors.New("unknown expression type")
		}
	}
	return nil
}

func (c *Compiler) compileExpression(exprNode ast.Expr, frame *Frame) error {
	switch expr := exprNode.(type) {
	case ast.NumberExpr:
		val := Value{}
		val.NewNum(expr.Value)
		frame.Constants = append(frame.Constants, val)
		frame.EmitUnary(LOAD_CONST, len(frame.Constants)-1, expr.Range.Start.Line)
	case ast.BoolExpr:
		val := Value{}
		val.NewBool(expr.Value)
		frame.Constants = append(frame.Constants, val)
		frame.EmitUnary(LOAD_CONST, len(frame.Constants)-1, expr.Range.Start.Line)
	case ast.StringExpr:
		val := Value{}
		val.NewString(expr.Value)
		frame.Constants = append(frame.Constants, val)
		frame.EmitUnary(LOAD_CONST, len(frame.Constants)-1, expr.Range.Start.Line)
	case ast.NullExpr:
		val := Value{}
		val.NewNull()
		frame.Constants = append(frame.Constants, val)
		frame.EmitUnary(LOAD_CONST, len(frame.Constants)-1, expr.Range.Start.Line)
	case ast.ListExpr:
		for _, listItem := range expr.Value {
			err := c.compileExpression(listItem, frame)
			if err != nil {
				return err
			}
		}
		frame.EmitUnary(CREATE_LIST, len(expr.Value), expr.Range.Start.Line)
	case ast.IfElseExpr:
		err := c.compileExpression(expr.Condition, frame)
		if err != nil {
			return err
		}
		frame.EmitUnary(COND_JUMP_FALSE, 0, expr.Range.Start.Line)
		condJumpInstrIdx := len(frame.Code) - 1
		err = c.compileBlock(expr.IfBranch, frame)
		if err != nil {
			return err
		}
		frame.Code[condJumpInstrIdx].Arg1 = len(frame.Code) - condJumpInstrIdx
		frame.EmitUnary(JUMP, 0, expr.Range.Start.Line)
		ifJumpIndx := len(frame.Code) - 1
		err = c.compileBlock(expr.ElseBranch, frame)
		if err != nil {
			return err
		}
		frame.Code[ifJumpIndx].Arg1 = len(frame.Code) - (ifJumpIndx + 1)
	case ast.IfOnlyExpr:
		err := c.compileExpression(expr.Condition, frame)
		if err != nil {
			return err
		}
		frame.EmitUnary(COND_JUMP_FALSE, 0, expr.Range.Start.Line)
		condJumpInstrIdx := len(frame.Code) - 1
		err = c.compileBlock(expr.IfBranch, frame)
		if err != nil {
			return err
		}
		frame.Code[condJumpInstrIdx].Arg1 = len(frame.Code) - condJumpInstrIdx
		frame.EmitUnary(JUMP, 1, expr.Range.Start.Line)
		frame.Emit(STORE_NULL, expr.Range.Start.Line)
	case ast.VarUseExpr:
		if idx, ok := frame.VariableMap[expr.Identifier]; ok {
			frame.EmitUnary(LOAD_VAR, idx, expr.Range.Start.Line)
		} else if idx, ok := c.GlobalVariableMap[expr.Identifier]; ok {
			frame.EmitUnary(LOAD_GLOBAL, idx, expr.Range.Start.Line)
		} else {
			return types.Error{Range: expr.GetRange(), Simple: fmt.Sprintf("Unknown variable %s", expr.Identifier)}
		}
	case ast.StructExpr:
		if structIdx, ok := c.StructMap[expr.StructIdentifier]; ok {
			frame.EmitUnary(CREATE_STRUCT, structIdx, expr.Range.Start.Line)
			structFields := c.Structs[structIdx]
			// First check to see if any values exist that don't exist on struct
			for valueFieldName, valueFieldExpr := range expr.Values {
				found := false
				for _, structFieldName := range structFields {
					if structFieldName == valueFieldName {
						found = true
						break
					}
				}
				if !found {
					return types.Error{Range: valueFieldExpr.GetRange(), Simple: fmt.Sprintf("Struct has no field %s", valueFieldName)}
				}
			}
			for _, fieldName := range structFields {
				frame.EmitUnary(STRUCT_FIELD_INDEX, getNameIndex(fieldName, frame), expr.Range.Start.Line)
				if valExpr, ok := expr.Values[fieldName]; ok {
					err := c.compileExpression(valExpr, frame)
					if err != nil {
						return err
					}
				} else {
					frame.Emit(STORE_NULL, expr.Range.Start.Line)
				}
				frame.Emit(SET_STRUCT_FIELD, expr.Range.Start.Line)
			}
		} else {
			return types.Error{Range: expr.Range, Simple: fmt.Sprintf("Use of undeclared struct %s", expr.StructIdentifier)}
		}
	case ast.StructAccessorExpr:
		err := c.compileExpression(expr.Struct, frame)
		if err != nil {
			return err
		}
		frame.EmitUnary(STRUCT_FIELD_INDEX, getNameIndex(expr.FieldIdentifier, frame), expr.Range.Start.Line)
		frame.Emit(GET_STRUCT_FIELD, expr.Range.Start.Line)
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
		for i := range expr.Args {
			argName := expr.Args[len(expr.Args)-(i+1)]
			closureFrame.Variables = append(closureFrame.Variables, Value{})
			closureFrame.VariableMap[argName] = len(closureFrame.Variables) - 1
			closureFrame.EmitUnary(STORE_VAR, len(closureFrame.Variables)-1, expr.Range.Start.Line)
		}
		err := c.compileBlock(expr.Body, &closureFrame)
		if err != nil {
			return err
		}

		// Closure is a value that needs to be pushed to top of stack
		closureValue := Value{}
		closureValue.NewClosure(expr.Args, &closureFrame)
		frame.Constants = append(frame.Constants, closureValue)
		frame.EmitUnary(LOAD_CONST, len(frame.Constants)-1, expr.Range.Start.Line)

		// Now capture the values of the variables
		for sourceIndex := range frame.Variables {
			frame.EmitBinary(PUSH_CLOSURE_VAR, sourceIndex, sourceIndex, expr.Range.Start.Line)
		}
		// Now capture globals - closures can not access or modify globals, only capture them into variables
		// From langauge user POV, this means that globals can only be read from closure and and changes exist only within
		// the closure
		for globalIndex := range c.GlobalVariables {
			// Target index set  - see above capturing logic.
			// Variables are in this order: <captured vars><captured globals><lambda arguments><closure variables>
			frame.EmitBinary(PUSH_GLOBAL_CLOSURE_VAR, globalIndex, len(frame.Variables)+globalIndex, expr.Range.Start.Line)
		}
	case ast.ClosureApplicationExpr:
		for _, arg := range expr.Args {
			err := c.compileExpression(arg, frame)
			if err != nil {
				return err
			}
		}
		err := c.compileExpression(expr.Closure, frame)
		if err != nil {
			return err
		}
		frame.Emit(CALL_CLOSURE, expr.Range.Start.Line)
	case ast.FunctionApplicationExpr:
		for _, arg := range expr.Args {
			err := c.compileExpression(arg, frame)
			if err != nil {
				return err
			}
		}
		if expr.Identifier == "+" {
			if len(expr.Args) != 2 {
				return types.Error{Range: expr.GetRange(), Simple: fmt.Sprintf("Expected 2 arguments, got %d", len(expr.Args))}
			}
			frame.Emit(ADD, expr.Range.Start.Line)
		} else {
			if idx, builtinFunc, ok := lookupBuiltin(expr.Identifier); ok {
				if len(expr.Args) != builtinFunc.NumArgs {
					return types.Error{Range: expr.GetRange(), Simple: fmt.Sprintf("Expected %d arguments, got %d", builtinFunc.NumArgs, len(expr.Args))}
				}
				frame.EmitUnary(CALL_BUILTIN, idx, expr.Range.Start.Line)
			} else if idx, ok := frame.VariableMap[expr.Identifier]; ok {
				frame.EmitUnary(LOAD_VAR, idx, expr.Range.Start.Line)
			} else if idx, ok := c.GlobalVariableMap[expr.Identifier]; ok {
				frame.EmitUnary(LOAD_GLOBAL, idx, expr.Range.Start.Line)
			} else if idx, ok := c.FunctionMap[expr.Identifier]; ok {
				frame.EmitUnary(CALL_FUNCTION, idx, expr.Range.Start.Line)
			} else {
				return types.Error{Range: expr.Range, Simple: fmt.Sprintf("Unknown identifier %s", expr.Identifier)}
			}
		}
	default:
		return errors.New(fmt.Sprintf("unsupported ast type %s", exprNode))
	}

	return nil
}

func (c *Compiler) compileStatement(stmtExpr ast.Stmt, frame *Frame) error {
	switch stmt := stmtExpr.(type) {
	case ast.VarDefStmt:
		err := c.compileExpression(stmt.Value, frame)
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
			frame.EmitUnary(STORE_GLOBAL, idx, stmt.Range.Start.Line)
		} else {
			idx, ok := frame.VariableMap[stmt.Identifier]
			if !ok {
				frame.Variables = append(frame.Variables, Value{})
				idx = len(frame.Variables) - 1
				frame.VariableMap[stmt.Identifier] = idx
			}
			frame.EmitUnary(STORE_VAR, idx, stmt.Range.Start.Line)
		}
		frame.Emit(STORE_NULL, stmt.Range.Start.Line)
	case ast.ImportStmt:
		// NOP
	case ast.WhileStmt:
		condStartIdx := len(frame.Code) - 1
		err := c.compileExpression(stmt.Condition, frame)
		if err != nil {
			return err
		}
		frame.EmitUnary(COND_JUMP_FALSE, 0, stmt.Range.Start.Line)
		condJumpIdx := len(frame.Code) - 1
		err = c.compileBlock(stmt.Body, frame)
		if err != nil {
			return err
		}
		frame.Code[condJumpIdx].Arg1 = len(frame.Code) - condJumpIdx
		frame.EmitUnary(JUMP, condStartIdx-len(frame.Code), stmt.Range.Start.Line)
		frame.Emit(STORE_NULL, stmt.Range.Start.Line)
	case ast.FuncDefStmt:
		functionFrame := Frame{}
		functionFrame.New()
		for i, argName := range stmt.Args {
			functionFrame.VariableMap[argName] = i
			functionFrame.Variables = append(functionFrame.Variables, Value{})
			functionFrame.FunctionArguments = stmt.Args
			functionFrame.FunctionName = stmt.Identifier
			// Store each argument from the stack into the variables array
			functionFrame.EmitUnary(STORE_VAR, len(stmt.Args)-(i+1), stmt.Range.Start.Line)
		}
		err := c.compileBlock(stmt.Body, &functionFrame)
		if err != nil {
			return err
		}
		if funcIdx, ok := c.FunctionMap[stmt.Identifier]; ok {
			c.Functions[funcIdx] = &functionFrame
		} else {
			c.Functions = append(c.Functions, &functionFrame)
			c.FunctionMap[stmt.Identifier] = len(c.Functions) - 1
			c.FunctionNames = append(c.FunctionNames, stmt.Identifier)
		}
	case ast.StructDefStmt:
		if _, ok := c.StructMap[stmt.Identifier]; ok {
			return types.Error{Range: stmt.Range, Simple: fmt.Sprintf("Duplicate declaration of struct %s", stmt.Identifier)}
		}
		c.Structs = append(c.Structs, stmt.FieldNames)
		c.StructMap[stmt.Identifier] = len(c.Structs) - 1
	case ast.StructFieldDeclarationStmt:
		if variableIdx, ok := frame.VariableMap[stmt.StructIdentifier]; ok {
			frame.EmitUnary(LOAD_VAR, variableIdx, stmt.Range.Start.Line)
		} else {
			if globalIdx, ok := c.GlobalVariableMap[stmt.StructIdentifier]; ok {
				frame.EmitUnary(LOAD_GLOBAL, globalIdx, stmt.Range.Start.Line)
			} else {
				return types.Error{Range: stmt.Range, Simple: fmt.Sprintf("Unknown variable %s", stmt.StructIdentifier)}
			}
		}
		frame.EmitUnary(STRUCT_FIELD_INDEX, getNameIndex(stmt.FieldIdentifier, frame), stmt.Range.Start.Line)
		err := c.compileExpression(stmt.Value, frame)
		if err != nil {
			return err
		}
		frame.Emit(SET_STRUCT_FIELD, stmt.Range.Start.Line)
	default:
		spew.Dump(stmt)
		return errors.New("unsupported statement")
	}
	return nil
}

func getNameIndex(nameToFind string, frame *Frame) int {
	idx := -1
	for i, name := range frame.Names {
		if name == nameToFind {
			idx = i
			break
		}
	}
	if idx == -1 {
		frame.Names = append(frame.Names, nameToFind)
		idx = len(frame.Names) - 1
	}
	return idx
}

func lookupBuiltin(identifier string) (int, Builtin, bool) {
	for i, item := range Builtins {
		if item.Identifier == identifier {
			return i, item, true
		}
	}

	return 0, Builtin{}, false
}
