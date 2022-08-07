package main

import "fmt"

type Expr interface {
	fmt.Stringer
}

type VariableExpr struct {
	Name Token
}

func (v VariableExpr) String() string {
	return v.Name.Lexeme
}

type LiteralExpr struct {
	Value interface{}
}

func (l LiteralExpr) String() string {
	return fmt.Sprint(l.Value)
}
