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

// FunctionValue represents a function's declaration and its closure
type FunctionValue struct {
	closure     *Environment[Value]
	declaration *FunctionStmt
}

// Call interprets the function declaration. It creates a new environment for the call,
// sets up the values of the function params with the call arguments, and interprets the
// body of the function.
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

	callEnv := NewEnvironment(f.closure)
	for i, param := range f.declaration.TypeExpr.Params {
		callEnv.Define(param.Name.Lexeme, args[i])
	}

	interpreter.interpretBlock(f.declaration.Body, callEnv)

	// only reachable from void functions
	return nil
}

func (f *FunctionValue) String() string {
	return f.declaration.String()
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
