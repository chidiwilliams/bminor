package main

import (
	"fmt"
)

type BMinorType interface {
	Equals(other BMinorType) bool
	ZeroValue() interface{}
}

type atomicType[T any] struct {
	name      string
	zeroValue T
}

func (t atomicType[T]) Equals(other BMinorType) bool {
	_, ok := other.(atomicType[T])
	return ok
}

func (t atomicType[T]) ZeroValue() interface{} {
	return t.zeroValue
}

var integerType = atomicType[int64]{name: "integer"}
var booleanType = atomicType[bool]{name: "boolean"}
var charType = atomicType[rune]{name: "char"}
var stringType = atomicType[string]{name: "string"}

var types = map[string]BMinorType{
	"integer": integerType,
	"boolean": booleanType,
	"char":    charType,
	"string":  stringType,
}

func NewTypeChecker(statements []Stmt) TypeChecker {
	return TypeChecker{statements: statements, env: map[string]BMinorType{}}
}

type TypeChecker struct {
	statements []Stmt
	types      map[string]BMinorType
	env        map[string]BMinorType
}

func (c *TypeChecker) Check() error {
	for _, stmt := range c.statements {
		err := c.checkStmt(stmt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *TypeChecker) checkStmt(stmt Stmt) error {
	switch stmt := stmt.(type) {
	case VarStmt:
		declaredType := c.getTypeByName(stmt.TypeDecl.Lexeme)
		if stmt.Initializer != nil {
			resolvedType := c.resolveType(stmt.Initializer)
			if !c.equals(resolvedType, declaredType) {
				return fmt.Errorf("expected value with type %s, but got %s", declaredType, resolvedType)
			}
		} else {
			stmt.Initializer = LiteralExpr{Value: declaredType.ZeroValue()}
		}
		c.env[stmt.Name.Lexeme] = declaredType
	case PrintStmt:
		for _, expression := range stmt.Expressions {
			c.resolveType(expression)
		}
	default:
		panic(fmt.Sprintf("unexpected statement type: %s", stmt))
	}
	return nil
}

func (c *TypeChecker) resolveType(expr Expr) BMinorType {
	switch expr := expr.(type) {
	case LiteralExpr:
		switch expr.Value.(type) {
		case int64:
			return integerType
		case bool:
			return booleanType
		case rune:
			return charType
		case string:
			return stringType
		}
	case VariableExpr:
		bMinorType, ok := c.env[expr.Name.Lexeme]
		if !ok {
			panic(fmt.Sprintf("could not resolve type for expression: %v", expr))
		}
		return bMinorType
	}
	panic(fmt.Sprintf("unexpected expression type: %s", expr))
}

func (c *TypeChecker) getTypeByName(name string) BMinorType {
	return types[name]
}

func (c *TypeChecker) equals(resolvedType BMinorType, declaredType BMinorType) bool {
	return resolvedType.Equals(declaredType)
}
