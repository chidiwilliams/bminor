package main

import "fmt"

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
	return fmt.Sprintf("var %s: %s = %s", v.Name.Lexeme, v.Type, v.Initializer)
}
