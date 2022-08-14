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
