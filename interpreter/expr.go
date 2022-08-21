package main

import (
	"fmt"
	"strings"
)

type Node interface {
	fmt.Stringer
	StartLine() int
}

type Expr interface {
	Node
}

type VariableExpr struct {
	Name Token
}

func (v *VariableExpr) StartLine() int {
	return v.Name.Line
}

func (v *VariableExpr) String() string {
	return v.Name.Lexeme
}

type LiteralExpr struct {
	Value     Value
	BeginLine int
}

func (l *LiteralExpr) StartLine() int {
	return l.BeginLine
}

func (l *LiteralExpr) String() string {
	return fmt.Sprint(l.Value)
}

type ArrayExpr struct {
	Elements  []Expr
	BeginLine int
}

func (a *ArrayExpr) StartLine() int {
	return a.BeginLine
}

func (a *ArrayExpr) String() string {
	return fmt.Sprint(a.Elements)
}

type MapExpr struct {
	Pairs     []Pair
	BeginLine int
}

func (m *MapExpr) StartLine() int {
	return m.BeginLine
}

func (m *MapExpr) String() string {
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

func (p Pair) StartLine() int {
	return p.Key.StartLine()
}

func (p Pair) String() string {
	return fmt.Sprintf("%s : %s", p.Key, p.Value)
}

type GetExpr struct {
	Object Expr
	Name   Expr
}

func (g *GetExpr) StartLine() int {
	return g.Object.StartLine()
}

func (g *GetExpr) String() string {
	return fmt.Sprintf("%s[%s]", g.Object, g.Name)
}

type SetExpr struct {
	Object Expr
	Name   Expr
	Value  Expr
}

func (s *SetExpr) StartLine() int {
	return s.Object.StartLine()
}

func (s *SetExpr) String() string {
	return fmt.Sprintf("%s[%s] = %s", s.Object, s.Name, s.Value)
}

type AssignExpr struct {
	Name  Token
	Value Expr
}

func (a *AssignExpr) StartLine() int {
	return a.Name.Line
}

func (a *AssignExpr) String() string {
	return fmt.Sprintf("%s = %s", a.Name.Lexeme, a.Value)
}

type PrefixExpr struct {
	Operator Token
	Right    Expr
}

func (e *PrefixExpr) StartLine() int {
	return e.Operator.Line
}

func (e *PrefixExpr) String() string {
	return fmt.Sprintf("%s%s", e.Operator.Lexeme, e.Right)
}

type PostfixExpr struct {
	Operator Token
	Left     *VariableExpr
}

func (e *PostfixExpr) StartLine() int {
	return e.Left.StartLine()
}

func (e *PostfixExpr) String() string {
	return fmt.Sprintf("%s%s", e.Left, e.Operator.Lexeme)
}

type BinaryExpr struct {
	Left     Expr
	Right    Expr
	Operator Token
}

func (b *BinaryExpr) StartLine() int {
	return b.Left.StartLine()
}

func (b *BinaryExpr) String() string {
	return fmt.Sprintf("%s %s %s", b.Left, b.Operator.Lexeme, b.Right)
}

type LogicalExpr struct {
	Left     Expr
	Right    Expr
	Operator Token
}

func (l *LogicalExpr) StartLine() int {
	return l.Left.StartLine()
}

func (l *LogicalExpr) String() string {
	return fmt.Sprintf("%s %s %s", l.Left, l.Operator.Lexeme, l.Right)
}

type CallExpr struct {
	Callee    Expr
	Paren     Token
	Arguments []Expr
}

func (c *CallExpr) StartLine() int {
	return c.Callee.StartLine()
}

func (c *CallExpr) String() string {
	args := make([]string, len(c.Arguments))
	for i, arg := range c.Arguments {
		args[i] = arg.String()
	}
	return fmt.Sprintf("%s(%s)", c.Callee, strings.Join(args, ", "))
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
	Type      AtomicTypeExprType
	BeginLine int
}

func (a AtomicTypeExpr) StartLine() int {
	return a.BeginLine
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
	BeginLine   int
	Reference   bool
}

func (a ArrayTypeExpr) StartLine() int {
	return a.BeginLine
}

func (a ArrayTypeExpr) String() string {
	return fmt.Sprintf("array [%d] %s", a.Length, a.ElementType)
}

type MapTypeExpr struct {
	KeyType   TypeExpr
	ValueType TypeExpr
	BeginLine int
}

func (m MapTypeExpr) StartLine() int {
	return m.BeginLine
}

func (m MapTypeExpr) String() string {
	return fmt.Sprintf("map %s %s", m.KeyType, m.ValueType)
}

type ParamTypeExpr struct {
	Name Token
	Type TypeExpr
}

func (p ParamTypeExpr) String() string {
	return fmt.Sprintf("%s: %s", p.Name.Lexeme, p.Type)
}

type FunctionTypeExpr struct {
	Params     []ParamTypeExpr
	ReturnType TypeExpr
	BeginLine  int
}

func (f FunctionTypeExpr) StartLine() int {
	return f.BeginLine
}

func (f FunctionTypeExpr) String() string {
	params := make([]string, len(f.Params))
	for i, param := range f.Params {
		params[i] = param.String()
	}
	return fmt.Sprintf("function %s ( %s )", f.ReturnType, strings.Join(params, ", "))
}

type VoidTypeExpr struct {
	BeginLine int
}

func (v VoidTypeExpr) String() string {
	return "void"
}

func (v VoidTypeExpr) StartLine() int {
	return v.BeginLine
}
