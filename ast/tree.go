package ast

import "github.com/benbanerjeerichards/lisp-calculator/types"

type Ast struct {
	Expression Expr
	Statement  Stmt
	Kind       string
}

type VarDefStmt struct {
	Identifier string
	Value      Expr
	Range      types.FileRange
}

type FuncDefStmt struct {
	Identifier string
	Args       []string
	Body       []Ast
	Range      types.FileRange
}

// No information is stored in this node as it is immedialy added to global ast construct env
type ImportStmt struct {
	Range types.FileRange
}

type VarUseExpr struct {
	Identifier string
	Range      types.FileRange
}

type FunctionApplicationExpr struct {
	Identifier string
	Args       []Expr
	Range      types.FileRange
}

type ClosureApplicationExpr struct {
	Closure Expr
	Args    []Expr
	Range   types.FileRange
}

type ClosureDefExpr struct {
	Args  []string
	Body  []Ast
	Range types.FileRange
}

type NumberExpr struct {
	Value float64
	Range types.FileRange
}

type StringExpr struct {
	Value string
	Range types.FileRange
}
type BoolExpr struct {
	Value bool
	Range types.FileRange
}

type NullExpr struct {
	Range types.FileRange
}

type ListExpr struct {
	Value []Expr
	Range types.FileRange
}

type IfElseExpr struct {
	Condition  Expr
	IfBranch   []Ast
	ElseBranch []Ast
	Range      types.FileRange
}

type IfOnlyExpr struct {
	Condition Expr
	IfBranch  []Ast
	Range     types.FileRange
}

type WhileStmt struct {
	Condition Expr
	Body      []Ast
	Range     types.FileRange
}

type StructDefStmt struct {
	Identifier string
	FieldNames []string
	Range      types.FileRange
}

type StructAccessorExpr struct {
	Struct          Expr
	FieldIdentifier string
	Range           types.FileRange
}

type StructFieldDeclarationStmt struct {
	StructIdentifier string
	FieldIdentifier  string
	Value            Expr
	Range            types.FileRange
}

type StructExpr struct {
	StructIdentifier string
	Values           map[string]Expr
	Range            types.FileRange
}

type ReturnStmt struct {
	Range types.FileRange
}

type ReturnValueStmt struct {
	Value Expr
	Range types.FileRange
}

func (v VarDefStmt) GetRange() types.FileRange {
	return v.Range
}

func (v FuncDefStmt) GetRange() types.FileRange {
	return v.Range
}

func (v VarUseExpr) GetRange() types.FileRange {
	return v.Range
}

func (v FunctionApplicationExpr) GetRange() types.FileRange {
	return v.Range
}

func (v NumberExpr) GetRange() types.FileRange {
	return v.Range
}

func (v StringExpr) GetRange() types.FileRange {
	return v.Range
}

func (v BoolExpr) GetRange() types.FileRange {
	return v.Range
}

func (v NullExpr) GetRange() types.FileRange {
	return v.Range
}

func (v ListExpr) GetRange() types.FileRange {
	return v.Range
}

func (v IfElseExpr) GetRange() types.FileRange {
	return v.Range
}

func (v IfOnlyExpr) GetRange() types.FileRange {
	return v.Range
}

func (v WhileStmt) GetRange() types.FileRange {
	return v.Range
}

func (v ImportStmt) GetRange() types.FileRange {
	return v.Range
}

func (v ClosureDefExpr) GetRange() types.FileRange {
	return v.Range
}

func (v ClosureApplicationExpr) GetRange() types.FileRange {
	return v.Range
}

func (v StructDefStmt) GetRange() types.FileRange {
	return v.Range
}

func (v StructAccessorExpr) GetRange() types.FileRange {
	return v.Range
}
func (v StructFieldDeclarationStmt) GetRange() types.FileRange {
	return v.Range
}
func (v StructExpr) GetRange() types.FileRange {
	return v.Range
}

func (v ReturnStmt) GetRange() types.FileRange {
	return v.Range
}

func (v ReturnValueStmt) GetRange() types.FileRange {
	return v.Range
}

// These are just to prevent assigning a statement to an expression
// Same as what go compiler does
func (FunctionApplicationExpr) exprType() {}
func (NumberExpr) exprType()              {}
func (VarUseExpr) exprType()              {}
func (BoolExpr) exprType()                {}
func (IfElseExpr) exprType()              {}
func (IfOnlyExpr) exprType()              {}
func (StringExpr) exprType()              {}
func (ListExpr) exprType()                {}
func (NullExpr) exprType()                {}
func (ClosureDefExpr) exprType()          {}
func (ClosureApplicationExpr) exprType()  {}
func (StructAccessorExpr) exprType()      {}
func (StructExpr) exprType()              {}

func (VarDefStmt) stmtType()                 {}
func (FuncDefStmt) stmtType()                {}
func (WhileStmt) stmtType()                  {}
func (ImportStmt) stmtType()                 {}
func (StructDefStmt) stmtType()              {}
func (StructFieldDeclarationStmt) stmtType() {}
func (ReturnStmt) stmtType()                 {}
func (ReturnValueStmt) stmtType()            {}

func (ast *Ast) newStatement(stmt Stmt) {
	ast.Kind = StmtType
	ast.Statement = stmt
}

func (ast *Ast) newExpression(expr Expr) {
	// Keep track of declared functions. This is needed to differentiate between function application
	// and variable usage
	ast.Kind = ExprType
	ast.Expression = expr
}
