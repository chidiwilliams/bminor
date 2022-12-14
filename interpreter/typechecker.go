package main

import (
	"fmt"
)

type typeError struct {
	message string
	line    int
}

func (e typeError) Error() string {
	return fmt.Sprintf("Error on line %d: %s", e.line+1, e.message)
}

var (
	typeInteger = newAtomicType[IntegerValue]("integer")
	typeBoolean = newAtomicType[BooleanValue]("boolean")
	typeChar    = newAtomicType[CharValue]("char")
	typeString  = newAtomicType[StringValue]("string")
	typeAny     = &anyType{}
	typeVoid    = &voidType{}
)

func NewTypeChecker() *TypeChecker {
	return &TypeChecker{env: NewEnvironment[Type](nil)}
}

type TypeChecker struct {
	env                        *Environment[Type]
	currentFunctionReturnType  Type
	hasCurrentFunctionReturned bool
}

func (c *TypeChecker) Check(statements []Stmt) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if recovered, ok := r.(typeError); ok {
				err = recovered
			} else if recovered, ok := r.(runtimeError); ok {
				err = recovered
			} else {
				panic(r)
			}
		}
	}()

	for _, stmt := range statements {
		c.checkStmt(stmt)
	}

	return nil
}

// checkStmt performs type checking on the given statement
func (c *TypeChecker) checkStmt(stmt Stmt) {
	switch stmt := stmt.(type) {
	case *VarStmt:
		declaredType := c.getType(stmt.Type)
		if stmt.Initializer != nil {
			resolvedType := c.resolveExpr(stmt.Initializer)
			c.expectExpr(stmt.Initializer, resolvedType, declaredType)
		} else {
			stmt.Initializer = &LiteralExpr{Value: declaredType.ZeroValue()}
		}
		c.env.Define(stmt.Name.Lexeme, declaredType)
	case *PrintStmt:
		for _, expression := range stmt.Expressions {
			c.resolveExpr(expression)
		}
	case *ExprStmt:
		c.resolveExpr(stmt.Expr)
	case *IfStmt:
		conditionType := c.resolveExpr(stmt.Condition)
		c.expectExpr(stmt.Condition, conditionType, typeBoolean)
		c.checkStmt(stmt.Consequent)
		if stmt.Alternative != nil {
			c.checkStmt(stmt.Alternative)
		}
	case *BlockStmt:
		previous := c.env
		c.env = NewEnvironment(previous)

		for _, innerStmt := range stmt.Statements {
			c.checkStmt(innerStmt)
		}

		c.env = previous
	case *FunctionStmt:
		if c.currentFunctionReturnType != nil {
			panic(c.error(stmt, "function definitions may not be nested"))
		}

		fnType := c.getType(stmt.TypeExpr)
		c.env.Define(stmt.Name.Lexeme, fnType)

		previous := c.env
		c.env = NewEnvironment(previous)
		c.hasCurrentFunctionReturned = false

		for _, param := range stmt.TypeExpr.Params {
			c.env.Define(param.Name.Lexeme, c.getType(param.Type))
		}

		c.currentFunctionReturnType = fnType.(*functionType).returnType

		c.checkStmt(stmt.Body)

		if c.currentFunctionReturnType != typeVoid && !c.hasCurrentFunctionReturned {
			panic(c.error(stmt, "expected function to return value of type '%s'", c.currentFunctionReturnType))
		}

		c.env = previous
	case *ReturnStmt:
		if stmt.Value != nil && c.currentFunctionReturnType == typeVoid {
			panic(c.error(stmt.Value, "not expecting any return value"))
		}
		valueType := c.resolveExpr(stmt.Value)
		c.expectExpr(stmt.Value, valueType, c.currentFunctionReturnType)
		c.hasCurrentFunctionReturned = true
	case *WhileStmt:
		conditionType := c.resolveExpr(stmt.Condition)
		c.expectExpr(stmt.Condition, conditionType, typeBoolean)
		c.checkStmt(stmt.Body)
	default:
		panic(c.error(stmt, fmt.Sprintf("unexpected statement type: %v", stmt)))
	}
}

// resolveExpr infers the type of an expression from its (possibly nested)
// literal values and operations while checking for type errors within the expression
func (c *TypeChecker) resolveExpr(expr Expr) Type {
	switch expr := expr.(type) {
	case *LiteralExpr:
		switch expr.Value.(type) {
		case IntegerValue:
			return typeInteger
		case BooleanValue:
			return typeBoolean
		case CharValue:
			return typeChar
		case StringValue:
			return typeString
		}
	case *VariableExpr:
		return c.env.Get(expr.Name.Lexeme)
	case *ArrayExpr:
		firstElementType := c.resolveExpr(expr.Elements[0])
		for _, element := range expr.Elements[1:] {
			elementType := c.resolveExpr(element)
			c.expectExpr(element, elementType, firstElementType)
		}
		return newArrayType(firstElementType, len(expr.Elements), false)
	case *MapExpr:
		if len(expr.Pairs) == 0 {
			return newMapType(typeAny, typeAny)
		}

		firstPair := expr.Pairs[0]
		firstKeyType := c.resolveExpr(firstPair.Key)
		firstValueType := c.resolveExpr(firstPair.Value)

		for _, pair := range expr.Pairs[1:] {
			keyType := c.resolveExpr(pair.Key)
			c.expectExpr(pair.Key, keyType, firstKeyType)
			valueType := c.resolveExpr(pair.Value)
			c.expectExpr(pair.Value, valueType, firstValueType)
		}

		return newMapType(firstKeyType, firstValueType)
	case *GetExpr:
		return c.resolveLookup(expr.Object, expr.Name)
	case *SetExpr:
		expectedValueType := c.resolveLookup(expr.Object, expr.Name)
		valueType := c.resolveExpr(expr.Value)
		c.expectExpr(expr.Value, valueType, expectedValueType)
		return expectedValueType
	case *BinaryExpr:
		leftType := c.resolveExpr(expr.Left)
		rightType := c.resolveExpr(expr.Right)

		switch expr.Operator.TokenType {
		case TokenPlus:
			c.expectExpr(expr.Left, leftType, typeInteger, typeString, typeChar)
		case TokenMinus, TokenStar, TokenSlash, TokenPercent,
			TokenCaret, TokenLess, TokenLessEqual, TokenGreater,
			TokenGreaterEqual:
			c.expectExpr(expr.Left, leftType, typeInteger)
		}

		if !leftType.Equals(rightType) {
			panic(c.error(expr.Left, "'%s' and '%s' are of different types", expr.Left, expr.Right))
		}

		switch expr.Operator.TokenType {
		case TokenLess, TokenGreater, TokenLessEqual,
			TokenGreaterEqual, TokenEqualEqual, TokenBangEqual:
			return typeBoolean
		default:
			return typeInteger
		}
	case *PrefixExpr:
		rightType := c.resolveExpr(expr.Right)
		switch expr.Operator.TokenType {
		case TokenMinus:
			c.expectExpr(expr.Right, rightType, typeInteger)
			return typeInteger
		case TokenBang:
			c.expectExpr(expr.Right, rightType, typeBoolean)
			return typeBoolean
		}
	case *PostfixExpr:
		leftType := c.resolveExpr(expr.Left)
		c.expectExpr(expr.Left, leftType, typeInteger)
		return typeInteger
	case *LogicalExpr:
		leftType := c.resolveExpr(expr.Left)
		c.expectExpr(expr.Left, leftType, typeBoolean)
		rightType := c.resolveExpr(expr.Right)
		c.expectExpr(expr.Right, rightType, typeBoolean)
		return typeBoolean
	case *AssignExpr:
		valueType := c.resolveExpr(expr.Value)
		varType := c.env.Get(expr.Name.Lexeme)
		c.expectExpr(expr.Value, valueType, varType)
		return valueType
	case *CallExpr:
		calleeType, ok := c.resolveExpr(expr.Callee).(*functionType)
		if !ok {
			panic(c.error(expr.Callee, "%s is not a function", expr.Callee))
		}

		for i, arg := range expr.Arguments {
			argType := c.resolveExpr(arg)
			expectedType := calleeType.paramTypes[i]
			c.expectExpr(arg, argType, expectedType)
		}

		return calleeType.returnType
	}
	panic(c.error(expr, "unexpected expression type: %s", expr))
}

func (c *TypeChecker) resolveLookup(object, name Expr) Type {
	objectType := c.resolveExpr(object)
	switch objectType := objectType.(type) {
	case *arrayType:
		indexType := c.resolveExpr(name)
		c.expectExpr(name, indexType, typeInteger)
		return objectType.elementType
	case *mapType:
		indexType := c.resolveExpr(name)
		c.expectExpr(name, indexType, objectType.keyType)
		return objectType.valueType
	default:
		panic(c.error(object, "can only index maps and arrays"))
	}
}

func (c *TypeChecker) getType(typeExpr TypeExpr) Type {
	switch typeExpr := typeExpr.(type) {
	case AtomicTypeExpr:
		switch typeExpr.Type {
		case AtomicTypeBoolean:
			return typeBoolean
		case AtomicTypeString:
			return typeString
		case AtomicTypeChar:
			return typeChar
		case AtomicTypeInteger:
			return typeInteger
		}
	case VoidTypeExpr:
		return typeVoid
	case ArrayTypeExpr:
		elementType := c.getType(typeExpr.ElementType)
		return newArrayType(elementType, typeExpr.Length, typeExpr.Reference)
	case MapTypeExpr:
		keyType := c.getType(typeExpr.KeyType)
		valueType := c.getType(typeExpr.ValueType)
		return newMapType(keyType, valueType)
	case FunctionTypeExpr:
		paramTypes := make([]Type, len(typeExpr.Params))
		for i, param := range typeExpr.Params {
			paramTypes[i] = c.getType(param.Type)
		}
		returnType := c.getType(typeExpr.ReturnType)
		return newFunctionType(paramTypes, returnType)
	}
	panic(c.error(typeExpr, "unexpected type: %s", typeExpr))
}

// expectExpr panics with a typeError if the given expression type is not equal to the expected type
func (c *TypeChecker) expectExpr(expr Expr, exprType Type, expectedTypes ...Type) {
	for _, expectedType := range expectedTypes {
		if exprType.Equals(expectedType) {
			return
		}
	}

	var err error
	if len(expectedTypes) == 1 {
		err = c.error(expr, "expected '%s' to be of type '%s', but got '%s'", expr, expectedTypes[0], exprType)
	} else {
		err = c.error(expr, "expected '%s' to be of type %s, but got '%s'", expr, expectedTypes, exprType)
	}
	panic(err)
}

// error returns a typeError at the given line
func (c *TypeChecker) error(node Node, format string, any ...any) error {
	return typeError{line: node.StartLine(), message: fmt.Sprintf(format, any...)}
}
