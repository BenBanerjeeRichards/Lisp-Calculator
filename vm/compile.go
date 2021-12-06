package vm

import (
	"errors"

	"github.com/benbanerjeerichards/lisp-calculator/ast"
)

type Compiler struct {
	GlobalVariables map[string]Value
	Functions       map[string]*ast.FuncDefStmt
}

// CompileFunction compiles the given AST into bytecode
func (c Compiler) CompileFunction(asts []ast.Ast) (Frame, error) {
	frame := Frame{}
	frame.New()

	for _, exprOrStmt := range asts {
		if exprOrStmt.Kind == ast.ExprType {
			err := c.CompileExpression(exprOrStmt.Expression, &frame)
			if err != nil {
				return Frame{}, err
			}
		} else if exprOrStmt.Kind == ast.StmtType {
			err := c.CompileStatement(&frame)
			if err != nil {
				return Frame{}, err
			}
		} else {
			panic("")
		}
	}

	return frame, nil
}

func (c Compiler) CompileExpression(exprNode ast.Expr, frame *Frame) error {
	switch expr := exprNode.(type) {
	case ast.NumberExpr:
		val := Value{}
		val.NewNum(expr.Value)
		frame.constants = append(frame.constants, val)
		frame.EmitUnary(LOAD_CONST, len(frame.constants)-1)
	case ast.FunctionApplicationExpr:
		if expr.Identifier == "+" {
			frame.Emit(ADD)
		} else {
			panic(expr.Identifier)
		}
	}

	return errors.New("unsupported expression")
}
func (c Compiler) CompileStatement(frame *Frame) error {
	return errors.New("unsupported statement")
}
