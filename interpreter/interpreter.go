package main

import (
	"fmt"
	"io"
	"math"
)

type runtimeError struct {
	message string
}

func (r runtimeError) Error() string {
	return fmt.Sprintf("error: %s", r.message)
}

// NewInterpreter returns a new interpreter
func NewInterpreter(stdOut io.Writer) *Interpreter {
	env := NewEnvironment[Value](nil)
	interpreter := &Interpreter{
		env:    env,
		stdOut: stdOut,
	}
	interpreter.predeclareFunction("puts", NewPredeclaredFunctionValue("puts", puts))
	return interpreter
}

type Interpreter struct {
	env          *Environment[Value]
	enclosingEnv *Environment[Value]
	stdOut       io.Writer
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
		i.interpret(statement)
	}
	return nil
}

func (i *Interpreter) interpret(stmt Stmt) {
	switch stmt := stmt.(type) {
	case *VarStmt:
		if !stmt.IsFnPrototype {
			value := i.evaluate(stmt.Initializer)
			i.env.Define(stmt.Name.Lexeme, value)
		}
	case *PrintStmt:
		for _, expr := range stmt.Expressions {
			value := i.evaluate(expr)
			_, _ = i.stdOut.Write([]byte(fmt.Sprint(value)))
		}
	case *ExprStmt:
		i.evaluate(stmt.Expr)
	case *IfStmt:
		condition := i.evaluate(stmt.Condition).(BooleanValue)
		if condition {
			i.interpret(stmt.Body)
		}
	case *BlockStmt:
		i.beginScope()
		for _, stmt := range stmt.Statements {
			i.interpret(stmt)
		}
		i.endScope()
	case *FunctionStmt:
		params := make([]string, len(stmt.TypeExpr.Params))
		for i, param := range stmt.TypeExpr.Params {
			params[i] = param.Name.Lexeme
		}
		fnValue := &FunctionValue{Params: params, Body: stmt.Body}
		i.env.Define(stmt.Name.Lexeme, fnValue)
	case *ReturnStmt:
		value := i.evaluate(stmt.Value)
		panic(Return{Value: value})
	case *ForStmt:
		for i.evaluate(stmt.Condition).(BooleanValue) {
			i.interpret(stmt.Body)
		}
	default:
		panic(i.error("unexpected statement type: %s", stmt))
	}
}

func (i *Interpreter) evaluate(expr Expr) Value {
	switch expr := expr.(type) {
	case *LiteralExpr:
		return expr.Value
	case *VariableExpr:
		return i.env.Get(expr.Name.Lexeme)
	case *ArrayExpr:
		array := make([]Value, len(expr.Elements))
		for j, element := range expr.Elements {
			array[j] = i.evaluate(element)
		}
		return ArrayValue(array)
	case *MapExpr:
		mapValue := make(MapValue)
		for _, pair := range expr.Pairs {
			key := i.evaluate(pair.Key)
			value := i.evaluate(pair.Value)
			mapValue[key] = value
		}
		return mapValue
	case *SetExpr:
		objectValue := i.evaluate(expr.Object)
		name := i.evaluate(expr.Name)
		value := i.evaluate(expr.Value)
		if array, ok := objectValue.(ArrayValue); ok {
			if name, ok := name.(IntegerValue); ok {
				array[name] = value
				return value
			}
		}
		if mapValue, ok := objectValue.(MapValue); ok {
			mapValue[name] = value
			return value
		}
	case *GetExpr:
		objectValue := i.evaluate(expr.Object)
		name := i.evaluate(expr.Name)
		if array, ok := objectValue.(ArrayValue); ok {
			if name, ok := name.(IntegerValue); ok {
				return array[name]
			}
		}
		if mapValue, ok := objectValue.(MapValue); ok {
			return mapValue[name]
		}
	case *BinaryExpr:
		left := i.evaluate(expr.Left)
		right := i.evaluate(expr.Right)
		switch expr.Operator.TokenType {
		case TokenPlus:
			return left.(IntegerValue) + right.(IntegerValue)
		case TokenMinus:
			return left.(IntegerValue) - right.(IntegerValue)
		case TokenStar:
			return left.(IntegerValue) * right.(IntegerValue)
		case TokenSlash:
			return left.(IntegerValue) / right.(IntegerValue)
		case TokenPercent:
			return left.(IntegerValue) % right.(IntegerValue)
		case TokenCaret:
			return IntegerValue(int(math.Pow(float64(left.(IntegerValue)), float64(right.(IntegerValue)))))
		case TokenLess:
			return BooleanValue(left.(IntegerValue) < right.(IntegerValue))
		case TokenLessEqual:
			return BooleanValue(left.(IntegerValue) <= right.(IntegerValue))
		case TokenGreater:
			return BooleanValue(left.(IntegerValue) > right.(IntegerValue))
		case TokenGreaterEqual:
			return BooleanValue(left.(IntegerValue) >= right.(IntegerValue))
		case TokenEqualEqual:
			return BooleanValue(left == right)
		case TokenBangEqual:
			return BooleanValue(left != right)
		}
	case *PrefixExpr:
		right := i.evaluate(expr.Right)
		switch expr.Operator.TokenType {
		case TokenMinus:
			return -right.(IntegerValue)
		}
	case *LogicalExpr:
		left := i.evaluate(expr.Left)
		right := i.evaluate(expr.Right)
		switch expr.Operator.TokenType {
		case TokenOr:
			return left.(BooleanValue) || right.(BooleanValue)
		case TokenAnd:
			return left.(BooleanValue) && right.(BooleanValue)
		}
	case *PostfixExpr:
		name := expr.Left.Name.Lexeme
		val := i.env.Get(name).(IntegerValue)
		switch expr.Operator.TokenType {
		case TokenPlusPlus:
			i.env.Assign(name, val+1)
			return val
		case TokenMinusMinus:
			i.env.Assign(name, val-1)
			return val
		}
	case *AssignExpr:
		name := expr.Name.Lexeme
		value := i.evaluate(expr.Value)
		i.env.Assign(name, value)
		return value
	case *CallExpr:
		callee := i.evaluate(expr.Callee).(CallableValue)
		args := make([]Value, len(expr.Arguments))
		for j, arg := range expr.Arguments {
			argValue := i.evaluate(arg)
			args[j] = argValue
		}
		return callee.Call(i, args)
	}
	panic(i.error("unexpected expression type: %s", expr))
}

func (i *Interpreter) error(format string, any ...any) error {
	return runtimeError{message: fmt.Sprintf(format, any...)}
}

func (i *Interpreter) beginScope() {
	i.enclosingEnv = i.env
	i.env = NewEnvironment[Value](i.enclosingEnv)
}

func (i *Interpreter) endScope() {
	i.env = i.enclosingEnv
	i.enclosingEnv = i.env.enclosing
}

func (i *Interpreter) predeclareFunction(name string, callable CallableValue) {
	i.env.Define(name, callable)
}
