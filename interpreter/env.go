package main

import "fmt"

type Environment[Value any] struct {
	enclosing *Environment[Value]
	values    map[string]Value
}

func NewEnvironment[Value any](enclosing *Environment[Value]) *Environment[Value] {
	return &Environment[Value]{enclosing: enclosing, values: make(map[string]Value)}
}

func (e *Environment[Value]) Define(name string, value Value) {
	e.values[name] = value
}

func (e *Environment[Value]) Get(name string) Value {
	if v, ok := e.values[name]; ok {
		return v
	}

	if e.enclosing != nil {
		return e.enclosing.Get(name)
	}

	panic(runtimeError{message: fmt.Sprintf("'%s' is not defined", name)})
}

func (e *Environment[Value]) Assign(name string, value Value) {
	if _, ok := e.values[name]; ok {
		e.values[name] = value
		return
	}

	if e.enclosing != nil {
		e.enclosing.Assign(name, value)
		return
	}

	panic(runtimeError{message: fmt.Sprintf("'%s' is not defined", name)})
}
