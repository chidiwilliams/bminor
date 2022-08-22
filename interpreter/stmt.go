package main

import (
	"fmt"
	"strings"
)

type Stmt interface {
	Node
}

type PrintStmt struct {
	Expressions []Expr
	BeginLine   int
}

func (p *PrintStmt) StartLine() int {
	return p.BeginLine
}

func (p *PrintStmt) String() string {
	return fmt.Sprintf("print %v;", p.Expressions)
}

type VarStmt struct {
	Name          Token
	Initializer   Expr
	Type          TypeExpr
	BeginLine     int
	IsFnPrototype bool
}

func (v *VarStmt) StartLine() int {
	return v.BeginLine
}

func (v *VarStmt) String() string {
	return fmt.Sprintf("var %s: %s = %s;", v.Name.Lexeme, v.Type, v.Initializer)
}

type ExprStmt struct {
	Expr Expr
}

func (e *ExprStmt) StartLine() int {
	return e.Expr.StartLine()
}

func (e *ExprStmt) String() string {
	return fmt.Sprintf("%s;", e.Expr)
}

type BlockStmt struct {
	Statements []Stmt
	BeginLine  int
}

func (b *BlockStmt) StartLine() int {
	return b.BeginLine
}

func (b *BlockStmt) String() string {
	s := "{\n"
	for _, stmt := range b.Statements {
		s += stmt.String() + "\n"
	}
	s += "\n}"
	return s
}

type IfStmt struct {
	Condition   Expr
	Consequent  Stmt
	Alternative Stmt
	BeginLine   int
}

func (i *IfStmt) StartLine() int {
	return i.BeginLine
}

func (i *IfStmt) String() string {
	return fmt.Sprintf("if (%s) {\n%s\n}", i.Condition, i.Consequent)
}

type ReturnStmt struct {
	Value     Expr
	BeginLine int
}

func (r *ReturnStmt) StartLine() int {
	return r.BeginLine
}

func (r *ReturnStmt) String() string {
	return fmt.Sprintf("return %s;", r.Value)
}

type FunctionStmt struct {
	Body      *BlockStmt
	TypeExpr  FunctionTypeExpr
	Name      Token
	BeginLine int
}

func (f *FunctionStmt) StartLine() int {
	return f.BeginLine
}

func (f *FunctionStmt) String() string {
	params := make([]string, len(f.TypeExpr.Params))
	for i, param := range f.TypeExpr.Params {
		params[i] = param.String()
	}

	return fmt.Sprintf("%s: function %s ( %s ) {\n%s}",
		f.Name.Lexeme, f.TypeExpr.ReturnType, strings.Join(params, ", "), f.Body)
}

type ForStmt struct {
	Condition Expr
	Body      Stmt
	BeginLine int
}

func (f *ForStmt) String() string {
	return fmt.Sprintf("for (%s) %s", f.Condition, f.Body)
}

func (f *ForStmt) StartLine() int {
	return f.BeginLine
}
