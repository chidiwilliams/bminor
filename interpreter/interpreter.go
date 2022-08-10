package main

import (
	"fmt"
	"io"
	"strings"
)

type Map map[interface{}]interface{}

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
	return "\"" + string(s) + "\""
}

func NewInterpreter(typeChecker *TypeChecker, stdOut io.Writer) *Interpreter {
	return &Interpreter{
		env:         map[string]interface{}{},
		typeChecker: typeChecker,
		stdOut:      stdOut,
	}
}

type Interpreter struct {
	env         map[string]interface{}
	typeChecker *TypeChecker // should this be passed in this way. or should it work similar to the resolved locals in Lox
	stdOut      io.Writer
}

func (i *Interpreter) Interpret(statements []Stmt) error {
	for _, statement := range statements {
		i.interpretStatement(statement)
	}
	return nil
}

func (i *Interpreter) interpretStatement(stmt Stmt) {
	switch stmt := stmt.(type) {
	case VarStmt:
		var value interface{}
		if stmt.Initializer == nil {
			value = i.typeChecker.getType(stmt.Type).ZeroValue()
		} else {
			value = i.interpretExpr(stmt.Initializer)
		}
		i.env[stmt.Name.Lexeme] = value
	case PrintStmt:
		for _, expr := range stmt.Expressions {
			value := i.interpretExpr(expr)
			_, _ = i.stdOut.Write([]byte(fmt.Sprint(value)))
		}
	case ExprStmt:
		i.interpretExpr(stmt.Expr)
	default:
		panic(fmt.Sprintf("unexpected statement type: %s", stmt))
	}
}

func (i *Interpreter) interpretExpr(expr Expr) interface{} {
	switch expr := expr.(type) {
	case LiteralExpr:
		// TODO: fix from parser
		switch value := expr.Value.(type) {
		case int:
			return Integer(value)
		default:
			return expr.Value
		}
	case VariableExpr:
		return i.env[expr.Name.Lexeme]
	case ArrayExpr:
		array := make([]interface{}, len(expr.Elements))
		for j, element := range expr.Elements {
			array[j] = i.interpretExpr(element)
		}
		return array
	case MapExpr:
		mapValue := make(Map)
		for _, pair := range expr.Pairs {
			key := i.interpretExpr(pair.Key)
			value := i.interpretExpr(pair.Value)
			mapValue[key] = value
		}
		return mapValue
	case SetExpr:
		objectValue := i.interpretExpr(expr.Object)
		name := i.interpretExpr(expr.Name)
		value := i.interpretExpr(expr.Value)
		if array, ok := objectValue.([]interface{}); ok {
			if name, ok := name.(int); ok {
				array[name] = value
				return value
			}
		}
		if mapValue, ok := objectValue.(Map); ok {
			mapValue[name] = value
			return value
		}
	case GetExpr:
		objectValue := i.interpretExpr(expr.Object)
		name := i.interpretExpr(expr.Name)
		if array, ok := objectValue.([]interface{}); ok {
			if name, ok := name.(int); ok {
				return array[name]
			}
		}
		if mapValue, ok := objectValue.(Map); ok {
			return mapValue[name]
		}
	}
	panic(fmt.Sprintf("unexpected expression type: %s", expr))
}
