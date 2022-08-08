package main

import (
	"fmt"
)

type Type interface {
	Equals(other Type) bool
	ZeroValue() interface{}
}

// newAtomicType returns a new atomic type based on an underlying Go type.
func newAtomicType[UnderlyingGoType any](name string) Type {
	return atomicType[UnderlyingGoType]{name: name}
}

// atomicType represents one of the four atomic types
// in B-minor: "string", "boolean", "char", and "integer".
type atomicType[UnderlyingGoType any] struct {
	name      string
	zeroValue UnderlyingGoType
}

func (t atomicType[UnderlyingGoType]) Equals(other Type) bool {
	_, ok := other.(atomicType[UnderlyingGoType])
	return ok
}

// ZeroValue returns the default value to be assigned to
// B-minor variables that are declared without an initializer.
// For example:
//
//	x: integer; // x has the zero integer value, 0
//	y: boolean; // y has the zero boolean value, false
//
// The zero value of an atomic type is the zero value of
// its underlying Go type.
func (t atomicType[UnderlyingGoType]) ZeroValue() interface{} {
	return t.zeroValue
}

func newArrayTypeValue(elementType Type, length int) arrayType {
	return arrayType{elementType: elementType, length: length}
}

type arrayType struct {
	length      int
	elementType Type
}

func (a arrayType) Equals(other Type) bool {
	otherArrayTypeValue, ok := other.(arrayType)
	if !ok {
		return false
	}

	return a.elementType.Equals(otherArrayTypeValue.elementType)
}

func (a arrayType) ZeroValue() interface{} {
	arr := make([]interface{}, a.length)
	for i := 0; i < a.length; i++ {
		arr[i] = a.elementType.ZeroValue()
	}
	return arr
}

var integerType = newAtomicType[int]("integer")
var booleanType = newAtomicType[bool]("boolean")
var charType = newAtomicType[rune]("char")
var stringType = newAtomicType[string]("string")

func NewTypeChecker(statements []Stmt) TypeChecker {
	return TypeChecker{statements: statements, env: map[string]Type{}}
}

type TypeChecker struct {
	statements []Stmt
	types      map[string]Type
	env        map[string]Type
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
		declaredType := c.getType(stmt.Type)
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

func (c *TypeChecker) resolveType(expr Expr) Type {
	switch expr := expr.(type) {
	case LiteralExpr:
		switch expr.Value.(type) {
		case int:
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
	case ArrayExpr:
		// TODO: handle resolving array types with no elements
		firstElementType := c.resolveType(expr.Elements[0])
		return newArrayTypeValue(firstElementType, len(expr.Elements))
	}
	panic(fmt.Sprintf("unexpected expression type: %s", expr))
}

func (c *TypeChecker) getType(typeExpr TypeExpr) Type {
	switch typeExpr := typeExpr.(type) {
	case AtomicTypeExpr:
		switch typeExpr.Type {
		case AtomicTypeBoolean:
			return booleanType
		case AtomicTypeString:
			return stringType
		case AtomicTypeChar:
			return charType
		case AtomicTypeInteger:
			return integerType
		}
	case ArrayTypeExpr:
		elementType := c.getType(typeExpr.ElementType)
		return newArrayTypeValue(elementType, typeExpr.Length)
	}
	panic(fmt.Sprintf("unexpected type: %s", typeExpr))
}

func (c *TypeChecker) equals(resolvedType Type, declaredType Type) bool {
	return resolvedType.Equals(declaredType)
}
