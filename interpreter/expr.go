package main

import (
	"fmt"
	"strings"
)

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

// TOOD: shouldn't this just be a literal expr with a map/array value
type ArrayExpr struct {
	Elements []Expr
}

func (a ArrayExpr) String() string {
	return fmt.Sprint(a.Elements)
}

type MapExpr struct {
	Pairs []Pair
}

func (m MapExpr) String() string {
	pairs := make([]string, len(m.Pairs))
	for i := range m.Pairs {
		pairs[i] = m.Pairs[i].String()
	}

	return fmt.Sprintf("{ %s }", strings.Join(pairs, ", "))
}

type Pair struct {
	Key   Expr
	Value Expr
}

func (p Pair) String() string {
	return fmt.Sprintf("%s : %s", p.Key, p.Value)
}

type GetExpr struct {
	Object Expr
	Name   Expr
}

func (g GetExpr) String() string {
	return fmt.Sprintf("%s[%s]", g.Object, g.Name)
}

type SetExpr struct {
	Object Expr
	Name   Expr
	Value  Expr
}

func (s SetExpr) String() string {
	return fmt.Sprintf("%s[%s] = %s", s.Object, s.Name, s.Value)
}

type AssignExpr struct {
	Name  Token
	Value Expr
}

func (a AssignExpr) String() string {
	return fmt.Sprintf("%s = %s", a.Name.Lexeme, a.Value)
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

type MapTypeExpr struct {
	KeyType   TypeExpr
	ValueType TypeExpr
}

func (m MapTypeExpr) String() string {
	return fmt.Sprintf("map %s %s", m.KeyType, m.ValueType)
}
