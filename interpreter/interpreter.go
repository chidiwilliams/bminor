package main

import (
	"fmt"
	"io"
)

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
	default:
		panic(fmt.Sprintf("unexpected statement type: %s", stmt))
	}
}

func (i *Interpreter) interpretExpr(expr Expr) interface{} {
	switch expr := expr.(type) {
	case LiteralExpr:
		return expr.Value
	case VariableExpr:
		return i.env[expr.Name.Lexeme]
	}
	panic(fmt.Sprintf("unexpected expression type: %s", expr))
}
