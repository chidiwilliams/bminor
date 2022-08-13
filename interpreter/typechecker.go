package main

import (
	"fmt"
)

type Type interface {
	// Equals returns true if both types are equal
	Equals(other Type) bool
	// ZeroValue returns the default value to be returned when a
	// variable of this type is declared without explicit initialization
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

func newArrayType(elementType Type, length int) arrayType {
	return arrayType{elementType: elementType, length: length}
}

type arrayType struct {
	length      int
	elementType Type
}

func (a arrayType) Equals(other Type) bool {
	otherArrayType, ok := other.(arrayType)
	if !ok {
		return false
	}

	return a.elementType.Equals(otherArrayType.elementType)
}

func (a arrayType) ZeroValue() interface{} {
	arr := make([]interface{}, a.length)
	for i := 0; i < a.length; i++ {
		arr[i] = a.elementType.ZeroValue()
	}
	return arr
}

type anyType struct {
}

func (a anyType) Equals(other Type) bool {
	return true
}

func (a anyType) ZeroValue() interface{} {
	panic("can't get zero value of any type")
}

func newMapType(keyType Type, valueType Type) mapType {
	return mapType{keyType: keyType, valueType: valueType}
}

type mapType struct {
	keyType   Type
	valueType Type
}

func (m mapType) Equals(other Type) bool {
	otherMapType, ok := other.(mapType)
	if !ok {
		return false
	}

	return m.keyType.Equals(otherMapType.keyType) &&
		m.valueType.Equals(otherMapType.valueType)
}

func (m mapType) ZeroValue() interface{} {
	return make(map[interface{}]interface{})
}

var integerType = newAtomicType[int]("integer")
var booleanType = newAtomicType[bool]("boolean")
var charType = newAtomicType[rune]("char")
var stringType = newAtomicType[string]("string")
var anyTypeType = anyType{}

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
			resolvedType := c.resolveExpr(stmt.Initializer)
			if !c.equals(resolvedType, declaredType) {
				return fmt.Errorf("expected value with type %s, but got %s", declaredType, resolvedType)
			}
		} else {
			stmt.Initializer = LiteralExpr{Value: declaredType.ZeroValue()}
		}
		c.env[stmt.Name.Lexeme] = declaredType
	case PrintStmt:
		for _, expression := range stmt.Expressions {
			c.resolveExpr(expression)
		}
	case ExprStmt:
		c.resolveExpr(stmt.Expr)
	default:
		return fmt.Errorf("unexpected statement type: %s", stmt)
	}
	return nil
}

func (c *TypeChecker) resolveExpr(expr Expr) Type {
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
		firstElementType := c.resolveExpr(expr.Elements[0])
		for _, element := range expr.Elements[1:] {
			elementType := c.resolveExpr(element)
			if !elementType.Equals(firstElementType) {
				panic(fmt.Sprintf("expected element to be of type: %s", elementType))
			}
		}
		return newArrayType(firstElementType, len(expr.Elements))
	case MapExpr:
		if len(expr.Pairs) == 0 {
			return newMapType(anyTypeType, anyTypeType)
		}

		firstPair := expr.Pairs[0]
		firstKeyType := c.resolveExpr(firstPair.Key)
		firstValueType := c.resolveExpr(firstPair.Value)

		for _, pair := range expr.Pairs[1:] {
			keyType := c.resolveExpr(pair.Key)
			if !keyType.Equals(firstKeyType) {
				panic(fmt.Sprintf("expected key to be of type: %s", firstKeyType))
			}
			valueType := c.resolveExpr(pair.Value)
			if !valueType.Equals(firstValueType) {
				panic(fmt.Sprintf("expected value to be of type: %s", firstValueType))
			}
		}

		return newMapType(firstKeyType, firstValueType)
	case GetExpr:
		return c.resolveLookup(expr.Object, expr.Name)
	case SetExpr:
		expectedValueType := c.resolveLookup(expr.Object, expr.Name)
		valueType := c.resolveExpr(expr.Value)
		if !expectedValueType.Equals(valueType) {
			panic(fmt.Sprintf("expected value to be of type: %s", expectedValueType))
		}
		return expectedValueType
	case BinaryExpr:
		left := c.resolveExpr(expr.Left)
		right := c.resolveExpr(expr.Right)
		if left.Equals(booleanType) || right.Equals(booleanType) {
			panic(fmt.Sprintf("cannot perform operation on boolean type"))
		}

		if !left.Equals(right) {
			panic(fmt.Sprintf("operands are of different types"))
		}

		switch expr.Operator.TokenType {
		case TokenLess, TokenGreater, TokenLessEqual,
			TokenGreaterEqual, TokenEqualEqual, TokenBangEqual:
			return booleanType
		default:
			if expr.Operator.TokenType == TokenPlus && left.Equals(stringType) {
				panic(fmt.Sprintf("cannot perform operation on string type"))
			}

			return integerType
		}
	case PrefixExpr:
		right := c.resolveExpr(expr.Right)
		switch expr.Operator.TokenType {
		case TokenMinus:
			if !right.Equals(integerType) {
				panic(fmt.Sprintf("must be an integer type"))
			}
			return integerType
		case TokenBang:
			if !right.Equals(booleanType) {
				panic(fmt.Sprintf("must be a boolean type"))
			}
			return booleanType
		}
	case PostfixExpr:
		left := c.resolveExpr(expr.Left)
		if !left.Equals(integerType) {
			panic(fmt.Sprintf("must be an integer type"))
		}
		return integerType
	case LogicalExpr:
		left := c.resolveExpr(expr.Left)
		right := c.resolveExpr(expr.Right)
		if !left.Equals(booleanType) {
			panic(fmt.Sprintf("must be a boolean type"))
		}
		if !left.Equals(right) {
			panic(fmt.Sprintf("operands are of different types"))
		}
		return booleanType
	}

	panic(fmt.Sprintf("unexpected expression type: %s", expr))
}

func (c *TypeChecker) resolveLookup(object, name Expr) Type {
	objectType := c.resolveExpr(object)
	switch objectType := objectType.(type) {
	case arrayType:
		nameType := c.resolveExpr(name)
		if !nameType.Equals(integerType) {
			panic(fmt.Sprintf("array index must be an integer"))
		}
		return objectType.elementType
	case mapType:
		nameType := c.resolveExpr(name)
		if !nameType.Equals(objectType.keyType) {
			panic(fmt.Sprintf("map key must be of type: %s", objectType.keyType))
		}
		return objectType.valueType
	default:
		panic(fmt.Sprintf("can only index maps and arrays"))
	}
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
		return newArrayType(elementType, typeExpr.Length)
	case MapTypeExpr:
		keyType := c.getType(typeExpr.KeyType)
		valueType := c.getType(typeExpr.ValueType)
		return newMapType(keyType, valueType)
	}
	panic(fmt.Sprintf("unexpected type: %s", typeExpr))
}

func (c *TypeChecker) equals(resolvedType Type, declaredType Type) bool {
	return resolvedType.Equals(declaredType)
}
