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

type TypeExpr interface {
	Expr
}

type AtomicTypeExprType int

const (
	AtomicTypeInteger AtomicTypeExprType = iota
	AtomicTypeString
	AtomicTypeChar
	AtomicTypeBoolean
)

type AtomicTypeExpr struct {
	Type AtomicTypeExprType
}

func (a AtomicTypeExpr) String() string {
	switch a.Type {
	case AtomicTypeInteger:
		return "integer"
	case AtomicTypeString:
		return "string"
	case AtomicTypeBoolean:
		return "boolean"
	case AtomicTypeChar:
		return "char"
	default:
		panic(fmt.Sprintf("unexpected atomic type: %v", a.Type))
	}
}

type ArrayTypeExpr struct {
	ElementType TypeExpr
	Length      int
}

func (a ArrayTypeExpr) String() string {
	return fmt.Sprintf("array [%d] %s", a.Length, a.ElementType)
}
