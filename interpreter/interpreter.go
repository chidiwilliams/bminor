package main

import (
	"fmt"
	"io"
)

type runtimeError struct {
	message string
}

func (r runtimeError) Error() string {
	return fmt.Sprintf("error: %s", r.message)
}

func NewInterpreter(typeChecker *TypeChecker, stdOut io.Writer) *Interpreter {
	return &Interpreter{
		env:         map[string]Value{},
		typeChecker: typeChecker,
		stdOut:      stdOut,
	}
}

type Interpreter struct {
	env         map[string]Value
	typeChecker *TypeChecker // should this be passed in this way. or should it work similar to the resolved locals in Lox
	stdOut      io.Writer
}

func (i *Interpreter) Interpret(statements []Stmt) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if recovered, ok := r.(runtimeError); ok {
				err = recovered
			} else {
				panic(err)
			}
		}
	}()
	for _, statement := range statements {
		i.interpretStatement(statement)
	}
	return nil
}

func (i *Interpreter) interpretStatement(stmt Stmt) {
	switch stmt := stmt.(type) {
	case VarStmt:
		var value Value
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

func (i *Interpreter) interpretExpr(expr Expr) Value {
	switch expr := expr.(type) {
	case LiteralExpr:
		return expr.Value
	case VariableExpr:
		return i.env[expr.Name.Lexeme]
	case ArrayExpr:
		array := make([]Value, len(expr.Elements))
		for j, element := range expr.Elements {
			array[j] = i.interpretExpr(element)
		}
		return Array(array)
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
		if array, ok := objectValue.(Array); ok {
			if name, ok := name.(Integer); ok {
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
		if array, ok := objectValue.(Array); ok {
			if name, ok := name.(Integer); ok {
				return array[name]
			}
		}
		if mapValue, ok := objectValue.(Map); ok {
			return mapValue[name]
		}
	case BinaryExpr:
		left := i.interpretExpr(expr.Left)
		right := i.interpretExpr(expr.Right)
		switch expr.Operator.TokenType {
		case TokenPlus:
			return left.(Integer) + right.(Integer)
		case TokenMinus:
			return left.(Integer) - right.(Integer)
		case TokenStar:
			return left.(Integer) * right.(Integer)
		case TokenSlash:
			return left.(Integer) / right.(Integer)
		case TokenPercent:
			return left.(Integer) % right.(Integer)
		case TokenCaret:
			return left.(Integer) ^ right.(Integer)
		case TokenLess:
			return Boolean(left.(Integer) < right.(Integer))
		case TokenLessEqual:
			return Boolean(left.(Integer) <= right.(Integer))
		case TokenGreater:
			return Boolean(left.(Integer) > right.(Integer))
		case TokenGreaterEqual:
			return Boolean(left.(Integer) >= right.(Integer))
		case TokenEqualEqual:
			return Boolean(left == right)
		case TokenBangEqual:
			return Boolean(left != right)
		}
	case PrefixExpr:
		right := i.interpretExpr(expr.Right)
		switch expr.Operator.TokenType {
		case TokenMinus:
			return -right.(Integer)
		}
	case LogicalExpr:
		left := i.interpretExpr(expr.Left)
		right := i.interpretExpr(expr.Right)
		switch expr.Operator.TokenType {
		case TokenOr:
			return left.(Boolean) || right.(Boolean)
		case TokenAnd:
			return left.(Boolean) && right.(Boolean)
		}
	case PostfixExpr:
		name := expr.Left.Name.Lexeme
		val := i.env[name].(Integer)
		switch expr.Operator.TokenType {
		case TokenPlusPlus:
			i.env[name] = val + 1
			return val
		case TokenMinusMinus:
			i.env[name] = val - 1
			return val
		}
	}
	panic(i.error(fmt.Sprintf("unexpected expression type: %s", expr)))
}

func (i *Interpreter) error(message string) error {
	return runtimeError{message: message}
}
