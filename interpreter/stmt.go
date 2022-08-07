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
	TypeDecl    Token
}

func (v VarStmt) String() string {
	return fmt.Sprintf("var %s: %s = %s", v.Name.Lexeme, v.TypeDecl.Lexeme, v.Initializer)
}
