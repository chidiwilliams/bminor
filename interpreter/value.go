package main

import (
	"fmt"
	"strings"
)

type Value interface {
	fmt.Stringer
}

type ArrayValue []Value

func (a ArrayValue) String() string {
	values := make([]string, len(a))
	for i, value := range a {
		values[i] = value.String()
	}
	return "[" + strings.Join(values, " ") + "]"
}

type MapValue map[any]Value

func (m MapValue) String() string {
	pairs := make([]string, 0)
	for key, value := range m {
		pairs = append(pairs, fmt.Sprintf("%s: %s", key, value))
	}
	return fmt.Sprintf("{ %s }", strings.Join(pairs, ", "))
}

type IntegerValue int

func (i IntegerValue) String() string {
	return fmt.Sprintf("%d", i)
}

type StringValue string

func (s StringValue) String() string {
	return string(s)
}

type BooleanValue bool

func (b BooleanValue) String() string {
	return fmt.Sprintf("%t", b)
}

type CharValue rune

func (c CharValue) String() string {
	return fmt.Sprintf("%d", c)
}

type Return struct {
	Value Value
}

type FunctionValue struct {
	Params []string
	Body   []Stmt
}

func (f *FunctionValue) Call(interpreter *Interpreter, args []Value) (value Value) {
	defer func() {
		if r := recover(); r != nil {
			if returnVal, ok := r.(Return); ok {
				value = returnVal.Value
			} else {
				panic(r)
			}
		}
	}()

	interpreter.beginScope()
	for i, param := range f.Params {
		interpreter.env.Define(param, args[i])
	}

	for _, stmt := range f.Body {
		interpreter.interpret(stmt)
	}

	interpreter.endScope()
	return nil
}

func (f *FunctionValue) String() string {
	body := make([]string, len(f.Body))
	for i, stmt := range f.Body {
		body[i] = stmt.String()
	}
	return fmt.Sprintf("( %s ) {%s\n}", strings.Join(f.Params, ", "), strings.Join(body, "\n"))
}

type CallableValue interface {
	Value
	Call(interpreter *Interpreter, args []Value) Value
}

type predeclaredFn func(interpreter *Interpreter, args []Value) Value

func NewPredeclaredFunctionValue(name string, fn predeclaredFn) *PredeclaredFunctionValue {
	return &PredeclaredFunctionValue{name: name, fn: fn}
}

type PredeclaredFunctionValue struct {
	name string
	fn   predeclaredFn
}

func (p *PredeclaredFunctionValue) Call(interpreter *Interpreter, args []Value) Value {
	return p.fn(interpreter, args)
}

func (p *PredeclaredFunctionValue) String() string {
	return p.name
}
