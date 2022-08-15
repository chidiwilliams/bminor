package main

import (
	"fmt"
	"strings"
)

type Stmt interface {
	fmt.Stringer
}

type PrintStmt struct {
	Expressions []Expr
}

func (p PrintStmt) String() string {
	return fmt.Sprintf("print %v;", p.Expressions)
}

type VarStmt struct {
	Name        Token
	Initializer Expr
	Type        TypeExpr
}

func (v VarStmt) String() string {
	return fmt.Sprintf("var %s: %s = %s;", v.Name.Lexeme, v.Type, v.Initializer)
}

type ExprStmt struct {
	Expr Expr
}

func (e ExprStmt) String() string {
	return fmt.Sprintf("%s;", e.Expr)
}

type BlockStmt struct {
	Statements []Stmt
}

func (b BlockStmt) String() string {
	s := "{\n"
	for _, stmt := range b.Statements {
		s += stmt.String() + "\n"
	}
	s += "\n}"
	return s
}

type IfStmt struct {
	Condition Expr
	Body      Stmt
}

func (i IfStmt) String() string {
	return fmt.Sprintf("if (%s) {\n%s\n}", i.Condition, i.Body)
}

type ReturnStmt struct {
	Value Value
}

func (r ReturnStmt) String() string {
	return fmt.Sprintf("return %s;", r.Value)
}

type FunctionStmt struct {
	Body     []Stmt
	TypeExpr FunctionTypeExpr
	Name     Token
}

func (f FunctionStmt) String() string {
	params := make([]string, len(f.TypeExpr.Params))
	for i, param := range f.TypeExpr.Params {
		params[i] = param.String()
	}

	body := make([]string, len(f.Body))
	for i, stmt := range f.Body {
		body[i] = stmt.String()
	}

	return fmt.Sprintf("%s: function %s ( %s ) {\n%s}",
		f.Name.Lexeme, f.TypeExpr.ReturnType, strings.Join(params, ", "), strings.Join(body, "\n"))
}
