package main

import (
	"fmt"
	"strings"
)

type Value interface {
	fmt.Stringer
}

type Array []Value

func (a Array) String() string {
	values := make([]string, len(a))
	for i, value := range a {
		values[i] = value.String()
	}
	return "[" + strings.Join(values, " ") + "]"
}

type Map map[any]Value

func (m Map) String() string {
	pairs := make([]string, 0)
	for key, value := range m {
		pairs = append(pairs, fmt.Sprintf("%s: %s", key, value))
	}
	return fmt.Sprintf("{ %s }", strings.Join(pairs, ", "))
}

type Integer int

func (i Integer) String() string {
	return fmt.Sprintf("%d", i)
}

type String string

func (s String) String() string {
	return string(s)
}

type Boolean bool

func (b Boolean) String() string {
	return fmt.Sprintf("%t", b)
}

type Char rune

func (c Char) String() string {
	return fmt.Sprintf("%d", c)
}

type Return struct {
	Value Value
}

type Function struct {
	Params []string
	Body   []Stmt
}

func (f *Function) Call(interpreter *Interpreter, args []Value) (value Value) {
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
		interpreter.interpretStatement(stmt)
	}

	interpreter.endScope()
	return nil
}

func (f *Function) String() string {
	body := make([]string, len(f.Body))
	for i, stmt := range f.Body {
		body[i] = stmt.String()
	}
	return fmt.Sprintf("( %s ) {%s\n}", strings.Join(f.Params, ", "), strings.Join(body, "\n"))
}

type Callable interface {
	Call(interpreter *Interpreter, args []Value) Value
}
